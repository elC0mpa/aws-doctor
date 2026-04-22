// Package orchestrator coordinates the execution of various AWS service checks.
package orchestrator

import (
	"context"
	"errors"
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
const sagemakerIdleDays = 14

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
		outputService:         cfg.OutputService,
		updateService:         cfg.UpdateService,
		reportService:         cfg.ReportService,
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
		workflowErr = s.wasteWorkflow(flags.WasteChecks, flags.Report, flags.ReportPath, flags.LambdaMemoryThreshold)
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
func (s *service) wasteWorkflow(wasteChecks []string, generateReport bool, reportPath string, lambdaMemoryThreshold int) error {
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)

	// Determine which checks to run
	runAll := len(wasteChecks) == 0
	runEC2 := runAll || slice.ContainsIgnoreCase(wasteChecks, "ec2")
	runELB := runAll || slice.ContainsIgnoreCase(wasteChecks, "elb")
	runS3 := runAll || slice.ContainsIgnoreCase(wasteChecks, "s3")
	runCloudWatchLogs := runAll || slice.ContainsIgnoreCase(wasteChecks, "cloudwatch")
	runRDS := runAll || slice.ContainsIgnoreCase(wasteChecks, "rds")
	runVPC := runAll || slice.ContainsIgnoreCase(wasteChecks, "vpc")
	runLambda := runAll || slice.ContainsIgnoreCase(wasteChecks, "lambda")
	runSagemaker := runAll || slice.ContainsIgnoreCase(wasteChecks, "sagemaker")

	var input model.RenderWasteInput

	var stsResult *sts.GetCallerIdentityOutput

	if runEC2 {
		s.queueEC2Checks(ctx, g, &input)
	}

	if runVPC {
		s.queueVPCChecks(ctx, g, &input)
	}

	if runELB {
		s.queueELBChecks(ctx, g, &input)
	}

	if runS3 {
		s.queueS3Checks(ctx, g, &input)
	}

	if runCloudWatchLogs {
		s.queueCloudWatchLogsChecks(ctx, g, &input)
	}

	if runRDS {
		s.queueRDSChecks(ctx, g, &input)
	}

	if runLambda {
		s.queueLambdaChecks(ctx, g, &input, lambdaMemoryThreshold)
	}

	if runSagemaker {
		s.queueSagemakerChecks(ctx, g, &input)
	}

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
		path, err := s.reportService.GenerateWasteReport(input, reportPath)
		if err != nil {
			return err
		}

		s.outputService.PrintReportSuccess(*path)

		return nil
	}

	return s.outputService.RenderWaste(input)
}

func (s *service) handleCostError(err error) error {
	if errors.Is(err, model.ErrFirstDayOfMonth) {
		s.outputService.StopSpinner()

		s.outputService.PrintFirstDayOfMonthError()

		return nil
	}

	return err
}
