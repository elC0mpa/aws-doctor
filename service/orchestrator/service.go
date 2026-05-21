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

func (s *service) wasteWorkflow(wasteChecks []string, generateReport bool, reportPath string, lambdaMemoryThreshold int, secretsIdleDays int) error {
	ctx := context.Background()

	s.loadPricing(ctx)

	stsResult, err := s.stsService.GetCallerIdentity(ctx)
	if err != nil {
		return err
	}

	resultCh := make(chan model.ScopeResult, 20)
	g, ctx := errgroup.WithContext(ctx)

	s.dispatchWasteChecks(ctx, g, resultCh, wasteChecks, lambdaMemoryThreshold, secretsIdleDays)

	// Wait and close channel in background
	go func() {
		_ = g.Wait()

		close(resultCh)
	}()

	s.outputService.StopSpinner()

	isInteractive := s.outputService.IsInteractive() && !generateReport
	if isInteractive {
		var scopes []string

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
		}
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
		mergeScopeResult(&finalInput, res.Input)
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if generateReport {
		return s.handleWasteReport(finalInput, reportPath)
	}

	return s.outputService.RenderWaste(finalInput, s.pricingService)
}

func (s *service) dispatchWasteChecks(ctx context.Context, g *errgroup.Group, resultCh chan<- model.ScopeResult, wasteChecks []string, lambdaMemoryThreshold int, secretsIdleDays int) {
	if shouldRunCheck(wasteChecks, "ec2") {
		s.queueEC2Checks(ctx, g, resultCh)
	}

	if shouldRunCheck(wasteChecks, "vpc") {
		s.queueVPCChecks(ctx, g, resultCh)
	}

	if shouldRunCheck(wasteChecks, "elb") {
		s.queueELBChecks(ctx, g, resultCh)
	}

	if shouldRunCheck(wasteChecks, "s3") {
		s.queueS3Checks(ctx, g, resultCh)
	}

	if shouldRunCheck(wasteChecks, "cloudwatch") {
		s.queueCloudWatchLogsChecks(ctx, g, resultCh)
	}

	if shouldRunCheck(wasteChecks, "rds") {
		s.queueRDSChecks(ctx, g, resultCh)
	}

	if shouldRunCheck(wasteChecks, "lambda") {
		s.queueLambdaChecks(ctx, g, resultCh, lambdaMemoryThreshold)
	}

	if shouldRunCheck(wasteChecks, "sagemaker") {
		s.queueSagemakerChecks(ctx, g, resultCh)
	}

	if shouldRunCheck(wasteChecks, "ecr") {
		s.queueECRChecks(ctx, g, resultCh)
	}

	if shouldRunCheck(wasteChecks, "secrets-manager") {
		s.queueSecretsManagerChecks(ctx, g, resultCh, secretsIdleDays)
	}
}

//nolint:gocyclo
func mergeScopeResult(dest *model.RenderWasteInput, src model.RenderWasteInput) {
	if len(src.ElasticIPs) > 0 {
		dest.ElasticIPs = append(dest.ElasticIPs, src.ElasticIPs...)
	}

	if len(src.UnusedVolumes) > 0 {
		dest.UnusedVolumes = append(dest.UnusedVolumes, src.UnusedVolumes...)
	}

	if len(src.StoppedVolumes) > 0 {
		dest.StoppedVolumes = append(dest.StoppedVolumes, src.StoppedVolumes...)
	}

	if len(src.Ris) > 0 {
		dest.Ris = append(dest.Ris, src.Ris...)
	}

	if len(src.StoppedInstances) > 0 {
		dest.StoppedInstances = append(dest.StoppedInstances, src.StoppedInstances...)
	}

	if len(src.IdleEC2Instances) > 0 {
		dest.IdleEC2Instances = append(dest.IdleEC2Instances, src.IdleEC2Instances...)
	}

	if len(src.LoadBalancers) > 0 {
		dest.LoadBalancers = append(dest.LoadBalancers, src.LoadBalancers...)
	}

	if len(src.UnusedAMIs) > 0 {
		dest.UnusedAMIs = append(dest.UnusedAMIs, src.UnusedAMIs...)
	}

	if len(src.OrphanedSnapshots) > 0 {
		dest.OrphanedSnapshots = append(dest.OrphanedSnapshots, src.OrphanedSnapshots...)
	}

	if len(src.UnusedKeyPairs) > 0 {
		dest.UnusedKeyPairs = append(dest.UnusedKeyPairs, src.UnusedKeyPairs...)
	}

	if len(src.S3Buckets) > 0 {
		dest.S3Buckets = append(dest.S3Buckets, src.S3Buckets...)
	}

	if len(src.S3MultipartUploads) > 0 {
		dest.S3MultipartUploads = append(dest.S3MultipartUploads, src.S3MultipartUploads...)
	}

	if len(src.CloudWatchLogGroups) > 0 {
		dest.CloudWatchLogGroups = append(dest.CloudWatchLogGroups, src.CloudWatchLogGroups...)
	}

	if len(src.RDSInstances) > 0 {
		dest.RDSInstances = append(dest.RDSInstances, src.RDSInstances...)
	}

	if len(src.RDSSnapshots) > 0 {
		dest.RDSSnapshots = append(dest.RDSSnapshots, src.RDSSnapshots...)
	}

	if len(src.RDSIdleInstances) > 0 {
		dest.RDSIdleInstances = append(dest.RDSIdleInstances, src.RDSIdleInstances...)
	}

	if len(src.IdleNATGateways) > 0 {
		dest.IdleNATGateways = append(dest.IdleNATGateways, src.IdleNATGateways...)
	}

	if len(src.IdleLoadBalancers) > 0 {
		dest.IdleLoadBalancers = append(dest.IdleLoadBalancers, src.IdleLoadBalancers...)
	}

	if len(src.OverProvisionedLambdas) > 0 {
		dest.OverProvisionedLambdas = append(dest.OverProvisionedLambdas, src.OverProvisionedLambdas...)
	}

	if len(src.IdleSageMakerEndpoints) > 0 {
		dest.IdleSageMakerEndpoints = append(dest.IdleSageMakerEndpoints, src.IdleSageMakerEndpoints...)
	}

	if len(src.ECRNoLifecyclePolicies) > 0 {
		dest.ECRNoLifecyclePolicies = append(dest.ECRNoLifecyclePolicies, src.ECRNoLifecyclePolicies...)
	}

	if len(src.ECREmptyRepositories) > 0 {
		dest.ECREmptyRepositories = append(dest.ECREmptyRepositories, src.ECREmptyRepositories...)
	}

	if len(src.ECRUntaggedImages) > 0 {
		dest.ECRUntaggedImages = append(dest.ECRUntaggedImages, src.ECRUntaggedImages...)
	}

	if len(src.UnusedSecrets) > 0 {
		dest.UnusedSecrets = append(dest.UnusedSecrets, src.UnusedSecrets...)
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
