// Package orchestrator coordinates the execution of various AWS service checks.
package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/elC0mpa/aws-doctor/model"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	"github.com/elC0mpa/aws-doctor/utils/slice"
	"github.com/google/go-github/v62/github"
	"golang.org/x/sync/errgroup"
)

// EC2-related thresholds for the various waste checks. The idle thresholds flag an instance only
// when the average CPU utilization AND the average daily network throughput (NetworkIn +
// NetworkOut combined) across the lookback window are both below the configured values, matching
// common Trusted-Advisor-style heuristics.
const (
	ec2StoppedDays            = 30
	ec2RiExpiringDays         = 30
	ec2AmiStaleDays           = 90
	ec2SnapshotStaleDays      = 90
	ec2IdleDays               = 14
	ec2IdleCPUPercent         = 5.0
	ec2IdleNetworkBytesPerDay = 5 * 1024 * 1024 // 5 MB/day, combined in+out
)

// Lookback windows and thresholds for the remaining waste checks.
const (
	sagemakerIdleDays  = 14
	vpcNatIdleDays     = 7
	elbIdleDays        = 7
	rdsIdleDays        = 7
	rdsSnapshotDays    = 30
	lambdaLookbackDays = 14
)

// NewService creates a new orchestrator service.
func NewService(cfg Config) Service {
	return &service{
		stsService:     cfg.STSService,
		costService:    cfg.CostService,
		pricingService: cfg.PricingService,
		outputService:  cfg.OutputService,
		updateService:  cfg.UpdateService,
		reportService:  cfg.ReportService,
		registry:       cfg.Registry,
		versionInfo:    cfg.VersionInfo,
	}
}

func (s *service) Orchestrate(flags model.Flags) error {
	if flags.Update {
		return s.updateWorkflow()
	}

	if flags.Version {
		return s.versionWorkflow()
	}

	// The notification prints to stderr, so running the check for every output format is safe for piping.
	versionCh := make(chan model.VersionCheckResult, 1)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		latest, err := s.updateService.CheckForUpdate(ctx)
		versionCh <- model.VersionCheckResult{LatestVersion: latest, Err: err}
	}()

	var workflowErr error

	switch {
	case flags.Waste:
		workflowErr = s.wasteWorkflow(flags.WasteChecks, flags.Report, flags.ReportPath, flags.LambdaMemoryThreshold, flags.SecretsIdleDays, flags.IAMIdleDays)
	case flags.Trend:
		workflowErr = s.trendWorkflow(flags.TrendChecks, flags.Report, flags.ReportPath)
	default:
		workflowErr = s.defaultWorkflow(flags.Report, flags.ReportPath)
	}

	select {
	case result := <-versionCh:
		if result.Err == nil && result.LatestVersion != nil {
			s.outputService.PrintNewVersionAvailable(s.versionInfo.Version, *result.LatestVersion)
		}
	case <-time.After(500 * time.Millisecond):
	}

	return workflowErr
}

func (s *service) versionWorkflow() error {
	s.outputService.StopSpinner()

	s.outputService.RenderVersion(s.versionInfo)

	return nil
}

func (s *service) updateWorkflow() error {
	s.outputService.StopSpinner()

	err := s.updateService.Update()
	if err == nil {
		return nil
	}

	if errors.Is(err, model.ErrHomebrewInstall) {
		s.outputService.PrintHomebrewUpdate()
		return nil
	}

	if errors.Is(err, model.ErrGoInstall) {
		s.outputService.PrintGoInstallUpdate()
		return nil
	}

	if errors.Is(err, model.ErrAlreadyLatest) {
		s.outputService.PrintAlreadyLatest(s.versionInfo.Version)
		return nil
	}

	if errors.Is(err, model.ErrRateLimit) {
		s.outputService.PrintRateLimitError()
		return err
	}

	var rateLimitErr *github.RateLimitError
	if errors.As(err, &rateLimitErr) {
		s.outputService.PrintRateLimitError()
		return err
	}

	s.outputService.PrintUpdateError(err)

	return err
}

func (s *service) defaultWorkflow(generateReport bool, reportPath string) error {
	currentMonthData, err := s.costService.GetCurrentMonthCostsByService(context.Background())
	if err != nil {
		return s.handleCostError(err)
	}

	lastMonthData, err := s.costService.GetLastMonthCostsByService(context.Background())
	if err != nil {
		return s.handleCostError(err)
	}

	currentTotalCost, err := s.costService.GetCurrentMonthTotalCosts(context.Background())
	if err != nil {
		return err
	}

	lastTotalCost, err := s.costService.GetLastMonthTotalCosts(context.Background())
	if err != nil {
		return err
	}

	stsResult, err := s.stsService.GetCallerIdentity(context.Background())
	if err != nil {
		return err
	}

	s.outputService.StopSpinner()

	input := model.RenderCostComparisonInput{
		AccountID:        *stsResult.Account,
		LastTotalCost:    *lastTotalCost,
		CurrentTotalCost: *currentTotalCost,
		LastMonth:        lastMonthData,
		CurrentMonth:     currentMonthData,
	}

	if generateReport {
		path, err := s.reportService.GenerateCostComparisonReport(input, reportPath)
		if err != nil {
			return err
		}

		s.outputService.PrintReportSuccess(*path)

		return nil
	}

	return s.outputService.RenderCostComparison(input)
}

func (s *service) trendWorkflow(trendChecks []string, generateReport bool, reportPath string) error {
	var mappedServices []string

	for _, svc := range trendChecks {
		if mapped, ok := awscostexplorer.ServiceNameMap[strings.ToLower(svc)]; ok {
			mappedServices = append(mappedServices, mapped)
		}
	}

	costInfo, err := s.costService.GetLastSixMonthsCosts(context.Background(), mappedServices)
	if err != nil {
		return err
	}

	stsResult, err := s.stsService.GetCallerIdentity(context.Background())
	if err != nil {
		return err
	}

	s.outputService.StopSpinner()

	if generateReport {
		path, err := s.reportService.GenerateTrendReport(*stsResult.Account, costInfo, trendChecks, reportPath)
		if err != nil {
			return err
		}

		s.outputService.PrintReportSuccess(*path)

		return nil
	}

	return s.outputService.RenderTrend(*stsResult.Account, costInfo, trendChecks)
}

func (s *service) wasteWorkflow(wasteChecks []string, generateReport bool, reportPath string, lambdaMemoryThreshold int, secretsIdleDays int, iamIdleDays int) error {
	ctx := context.Background()

	s.loadPricing(ctx)

	stsResult, err := s.stsService.GetCallerIdentity(ctx)
	if err != nil {
		return err
	}

	resultCh := make(chan model.ScopeResult, 20)
	g, ctx := errgroup.WithContext(ctx)

	flags := model.Flags{
		WasteChecks:               wasteChecks,
		LambdaMemoryThreshold:     lambdaMemoryThreshold,
		SecretsIdleDays:           secretsIdleDays,
		IAMIdleDays:               iamIdleDays,
		EC2StoppedDays:            ec2StoppedDays,
		EC2RiExpiringDays:         ec2RiExpiringDays,
		EC2AmiStaleDays:           ec2AmiStaleDays,
		EC2SnapshotStaleDays:      ec2SnapshotStaleDays,
		EC2IdleDays:               ec2IdleDays,
		EC2IdleCPUPercent:         ec2IdleCPUPercent,
		EC2IdleNetworkBytesPerDay: ec2IdleNetworkBytesPerDay,
		SageMakerIdleDays:         sagemakerIdleDays,
		VPCNatIdleDays:            vpcNatIdleDays,
		ELBIdleDays:               elbIdleDays,
		RDSIdleDays:               rdsIdleDays,
		RDSSnapshotDays:           rdsSnapshotDays,
		LambdaLookbackDays:        lambdaLookbackDays,
	}

	allScopes := []struct{ name, tab string }{
		{"ec2", "EC2"},
		{"vpc", "VPC"},
		{"elb", "ELB"},
		{"s3", "S3"},
		{"cloudwatch", "CloudWatch"},
		{"rds", "RDS"},
		{"lambda", "Lambda"},
		{"sagemaker", "SageMaker"},
		{"ecr", "ECR"},
		{"secrets-manager", "SecretsManager"},
		{"iam", "IAM"},
	}

	tabMap := make(map[string]string)
	for _, s := range allScopes {
		tabMap[s.name] = s.tab
	}

	analyzers := s.registry.GetAnalyzers()
	for _, a := range analyzers {
		if !shouldRunCheck(wasteChecks, a.Name()) {
			continue
		}

		analyzer := a

		g.Go(func() error {
			res, err := analyzer.Analyze(ctx, flags)

			// Map the lowercase analyzer name back to the expected Tab name
			if tab, ok := tabMap[analyzer.Name()]; ok {
				res.Scope = tab
			}

			if err != nil {
				// For now, we mimic old behavior by passing the error inside ScopeResult
				// The previous code didn't fail the errgroup for an individual check failure
				res.Err = err
			}

			resultCh <- res

			return nil
		})
	}

	// Wait and close channel in background
	go func() {
		_ = g.Wait()

		close(resultCh)
	}()

	s.outputService.StopSpinner()

	isInteractive := s.outputService.IsInteractive() && !generateReport
	if isInteractive {
		var scopes []string

		for _, s := range allScopes {
			if shouldRunCheck(wasteChecks, s.name) {
				scopes = append(scopes, s.tab)
			}
		}

		err := s.outputService.RenderWasteInteractive(*stsResult.Account, resultCh, scopes, s.pricingService)
		workflowErr := g.Wait()

		if err != nil {
			s.outputService.PrintWasteError(err)
			return err
		}

		return workflowErr
	}

	finalInput := model.RenderWasteInput{AccountID: *stsResult.Account}
	for res := range resultCh {
		finalInput.Merge(res.Input)
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if generateReport {
		return s.handleWasteReport(finalInput, reportPath)
	}

	return s.outputService.RenderWaste(finalInput, s.pricingService)
}

func shouldRunCheck(wasteChecks []string, name string) bool {
	return len(wasteChecks) == 0 || slice.ContainsIgnoreCase(wasteChecks, name)
}

func (s *service) handleWasteReport(input model.RenderWasteInput, reportPath string) error {
	path, err := s.reportService.GenerateWasteReport(input, s.pricingService, reportPath)
	if err != nil {
		return err
	}

	s.outputService.PrintReportSuccess(*path)

	return nil
}

func (s *service) handleCostError(err error) error {
	if errors.Is(err, model.ErrFirstDayOfMonth) {
		s.outputService.StopSpinner()

		s.outputService.PrintFirstDayOfMonthError()

		return nil
	}

	return err
}

// loadPricing fetches region-aware pricing at the start of the waste workflow so the Calculate*
// helpers can surface accurate rates instead of the hardcoded us-east-1 defaults. The call is
// best-effort: any Pricing API failures are surfaced to stderr and the fallback constants cover
// the missing entries. The spinner is updated in place so the user sees why startup is pausing.
func (s *service) loadPricing(ctx context.Context) {
	s.outputService.SetSpinnerMessage("Gathering pricing data...")

	pricingCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	if err := s.pricingService.LoadRegionRates(pricingCtx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: pricing API partial failure, falling back to defaults: %v\n", err)
	}

	s.outputService.SetSpinnerMessage("Please wait while data is being fetched...")
}
