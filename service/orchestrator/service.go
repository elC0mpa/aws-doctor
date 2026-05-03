// Package orchestrator coordinates the execution of various AWS service checks.
package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/elC0mpa/aws-doctor/model"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	"github.com/elC0mpa/aws-doctor/utils/slice"
	"github.com/google/go-github/v62/github"
	"golang.org/x/sync/errgroup"
)

// sagemakerIdleDays is the lookback window for flagging SageMaker endpoints with zero
// invocations as idle. Matches the hardcoded windows used by other waste checks
// (unused AMIs, orphaned snapshots, idle NAT gateways).
const (
	sagemakerIdleDays    = 14
	ec2StoppedDays       = 30
	ec2RiExpiringDays    = 30
	ec2AmiStaleDays      = 90
	ec2SnapshotStaleDays = 90
	vpcNatIdleDays       = 7
	elbIdleDays          = 7
	rdsIdleDays          = 7
	rdsSnapshotDays      = 30
	lambdaLookbackDays   = 14
)

// NewService creates a new orchestrator service.
func NewService(cfg Config) Service {
	return &service{
		stsService:            cfg.STSService,
		costService:           cfg.CostService,
		ec2Service:            cfg.EC2Service,
		elbService:            cfg.ELBService,
		s3Service:             cfg.S3Service,
		cloudwatchlogsService: cfg.CloudWatchLogsService,
		rdsService:            cfg.RDSService,
		lambdaService:         cfg.LambdaService,
		sagemakerService:      cfg.SageMakerService,
		secretsmanagerService: cfg.SecretsManagerService,
		pricingService:        cfg.PricingService,
		outputService:         cfg.OutputService,
		updateService:         cfg.UpdateService,
		reportService:         cfg.ReportService,
		ecrService:            cfg.ECRService,
		versionInfo:           cfg.VersionInfo,
		vpcService:            cfg.VPCService,
	}
}

func (s *service) Orchestrate(flags model.Flags) error {
	if flags.Update {
		return s.updateWorkflow()
	}

	if flags.Version {
		return s.versionWorkflow()
	}

	// TODO: cache the version check result locally to avoid hitting the GitHub API on every run.
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
		workflowErr = s.wasteWorkflow(flags.WasteChecks, flags.Report, flags.ReportPath, flags.LambdaMemoryThreshold, flags.SecretsIdleDays)
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

// wasteWorkflow dispatches one check per AWS service concurrently.
func (s *service) wasteWorkflow(wasteChecks []string, generateReport bool, reportPath string, lambdaMemoryThreshold int, secretsIdleDays int) error {
	ctx := context.Background()

	s.loadPricing(ctx)

	g, ctx := errgroup.WithContext(ctx)

	input := model.RenderWasteInput{}

	s.dispatchWasteChecks(ctx, g, &input, wasteChecks, lambdaMemoryThreshold, secretsIdleDays)

	var stsResult *sts.GetCallerIdentityOutput

	// Fetch caller identity concurrently (always required for output)
	g.Go(func() error {
		var err error

		stsResult, err = s.stsService.GetCallerIdentity(ctx)

		return err
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return err
	}

	s.outputService.StopSpinner()

	input.AccountID = *stsResult.Account

	if generateReport {
		return s.handleWasteReport(input, reportPath)
	}

	return s.outputService.RenderWaste(input, s.pricingService)
}

func (s *service) dispatchWasteChecks(ctx context.Context, g *errgroup.Group, input *model.RenderWasteInput, wasteChecks []string, lambdaMemoryThreshold int, secretsIdleDays int) {
	if shouldRunCheck(wasteChecks, "ec2") {
		s.queueEC2Checks(ctx, g, input)
	}

	if shouldRunCheck(wasteChecks, "vpc") {
		s.queueVPCChecks(ctx, g, input)
	}

	if shouldRunCheck(wasteChecks, "elb") {
		s.queueELBChecks(ctx, g, input)
	}

	if shouldRunCheck(wasteChecks, "s3") {
		s.queueS3Checks(ctx, g, input)
	}

	if shouldRunCheck(wasteChecks, "cloudwatch") {
		s.queueCloudWatchLogsChecks(ctx, g, input)
	}

	if shouldRunCheck(wasteChecks, "rds") {
		s.queueRDSChecks(ctx, g, input)
	}

	if shouldRunCheck(wasteChecks, "lambda") {
		s.queueLambdaChecks(ctx, g, input, lambdaMemoryThreshold)
	}

	if shouldRunCheck(wasteChecks, "sagemaker") {
		s.queueSagemakerChecks(ctx, g, input)
	}

	if shouldRunCheck(wasteChecks, "ecr") {
		s.queueECRChecks(ctx, g, input)
	}

	if shouldRunCheck(wasteChecks, "secrets-manager") {
		s.queueSecretsManagerChecks(ctx, g, input, secretsIdleDays)
	}
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
