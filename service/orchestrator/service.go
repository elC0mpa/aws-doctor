// Package orchestrator coordinates the execution of various AWS service checks.
package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/elC0mpa/aws-doctor/model"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	"github.com/elC0mpa/aws-doctor/utils/pricing"
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
		awsConfig:             cfg.AWSConfig,
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

//nolint:gocyclo // waste workflow dispatches one check per AWS service; each adds complexity
func (s *service) wasteWorkflow(wasteChecks []string, generateReport bool, reportPath string, lambdaMemoryThreshold int) error {
	ctx := context.Background()

	s.loadPricing(ctx)

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

	// Results from concurrent API calls
	var (
		elasticIPInfo                            []types.Address
		availableEBSVolumesInfo                  []types.Volume
		stoppedInstancesMoreThan30Days           []types.Instance
		attachedToStoppedInstancesEBSVolumesInfo []types.Volume
		expireReservedInstancesInfo              []model.RiExpirationInfo
		unusedLoadBalancers                      []elbtypes.LoadBalancer
		idleLoadBalancers                        []model.ELBIdleInfo
		unusedAMIs                               []model.AMIWasteInfo
		orphanedSnapshots                        []model.SnapshotWasteInfo
		unusedKeyPairs                           []model.KeyPairWasteInfo
		s3Buckets                                []model.S3BucketWasteInfo
		s3MultipartUploads                       []model.S3MultipartUploadWasteInfo
		cloudwatchLogs                           []model.CloudWatchLogsWasteInfo
		rdsInstances                             []model.RDSInstanceWasteInfo
		rdsSnapshots                             []model.RDSSnapshotWasteInfo
		rdsIdleInstances                         []model.RDSIdleInstanceInfo
		idleNATGateways                          []model.NATGatewayWasteInfo
		overProvisionedLambdas                   []model.LambdaOverProvisionedInfo
		idleSageMakerEndpoints                   []model.IdleSageMakerEndpointInfo
		stsResult                                *sts.GetCallerIdentityOutput
	)

	if runEC2 {
		// Fetch unused Elastic IPs concurrently
		g.Go(func() error {
			var err error

			elasticIPInfo, err = s.ec2Service.GetUnusedElasticIPAddressesInfo(ctx)

			return err
		})

		// Fetch unused EBS volumes concurrently
		g.Go(func() error {
			var err error

			availableEBSVolumesInfo, err = s.ec2Service.GetUnusedEBSVolumes(ctx)

			return err
		})

		// Fetch stopped instances info concurrently
		g.Go(func() error {
			var err error

			stoppedInstancesMoreThan30Days, attachedToStoppedInstancesEBSVolumesInfo, err = s.ec2Service.GetStoppedInstancesInfo(ctx)

			return err
		})

		// Fetch reserved instance expiration info concurrently
		g.Go(func() error {
			var err error

			expireReservedInstancesInfo, err = s.ec2Service.GetReservedInstanceExpiringOrExpired30DaysWaste(ctx)

			return err
		})

		// Fetch unused AMIs concurrently
		g.Go(func() error {
			var err error

			unusedAMIs, err = s.ec2Service.GetUnusedAMIs(ctx, 90)

			return err
		})

		// Fetch orphaned EBS snapshots concurrently
		g.Go(func() error {
			var err error

			orphanedSnapshots, err = s.ec2Service.GetOrphanedSnapshots(ctx, 90)

			return err
		})

		// Fetch unused keypairs concurrently
		g.Go(func() error {
			var err error

			unusedKeyPairs, err = s.ec2Service.GetUnusedKeyPairs(ctx)

			return err
		})
	}

	if runVPC {
		// Fetch idle NAT Gateways concurrently
		g.Go(func() error {
			var err error

			idleNATGateways, err = s.vpcService.GetIdleNATGateways(ctx, 7)

			return err
		})
	}

	if runELB {
		// Fetch unused and idle Load Balancers concurrently
		g.Go(func() error {
			var err error

			unusedLoadBalancers, idleLoadBalancers, err = s.elbService.GetLoadBalancerWaste(ctx)

			return err
		})
	}

	if runS3 {
		// Fetch S3 waste concurrently
		g.Go(func() error {
			var err error

			s3Buckets, s3MultipartUploads, err = s.s3Service.GetS3Waste(ctx)

			return err
		})
	}

	if runCloudWatchLogs {
		// Fetch CloudWatch Logs waste concurrently
		g.Go(func() error {
			var err error

			cloudwatchLogs, err = s.cloudwatchlogsService.GetCloudWatchLogsWaste(ctx)

			return err
		})
	}

	if runRDS {
		// Fetch RDS waste concurrently
		g.Go(func() error {
			var err error

			rdsInstances, rdsSnapshots, rdsIdleInstances, err = s.rdsService.GetRDSWaste(ctx)

			return err
		})
	}

	if runLambda {
		// Fetch over-provisioned Lambda functions concurrently
		g.Go(func() error {
			var err error

			overProvisionedLambdas, err = s.lambdaService.GetOverProvisionedFunctions(ctx, lambdaMemoryThreshold)

			return err
		})
	}

	if s.sagemakerService != nil && (runAll || slice.ContainsIgnoreCase(wasteChecks, "sagemaker")) {
		g.Go(func() error {
			var err error

			idleSageMakerEndpoints, err = s.sagemakerService.GetIdleEndpoints(ctx, sagemakerIdleDays)

			return err
		})
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

	input := model.RenderWasteInput{
		AccountID:              *stsResult.Account,
		ElasticIPs:             elasticIPInfo,
		UnusedVolumes:          availableEBSVolumesInfo,
		StoppedVolumes:         attachedToStoppedInstancesEBSVolumesInfo,
		Ris:                    expireReservedInstancesInfo,
		StoppedInstances:       stoppedInstancesMoreThan30Days,
		LoadBalancers:          unusedLoadBalancers,
		UnusedAMIs:             unusedAMIs,
		OrphanedSnapshots:      orphanedSnapshots,
		UnusedKeyPairs:         unusedKeyPairs,
		S3Buckets:              s3Buckets,
		S3MultipartUploads:     s3MultipartUploads,
		CloudWatchLogGroups:    cloudwatchLogs,
		RDSInstances:           rdsInstances,
		RDSSnapshots:           rdsSnapshots,
		RDSIdleInstances:       rdsIdleInstances,
		IdleNATGateways:        idleNATGateways,
		IdleLoadBalancers:      idleLoadBalancers,
		OverProvisionedLambdas: overProvisionedLambdas,
		IdleSageMakerEndpoints: idleSageMakerEndpoints,
	}

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

// loadPricing fetches region-aware pricing at the start of the waste workflow so the Calculate*
// helpers can surface accurate rates instead of the hardcoded us-east-1 defaults. The call is
// best-effort: any Pricing API failures are surfaced to stderr and the fallback constants cover
// the missing entries. The spinner is updated in place so the user sees why startup is pausing.
func (s *service) loadPricing(ctx context.Context) {
	s.outputService.SetSpinnerMessage("Gathering pricing data...")

	pricingCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := pricing.Load(pricingCtx, s.awsConfig); err != nil {
		fmt.Fprintf(os.Stderr, "warning: pricing API partial failure, falling back to defaults: %v\n", err)
	}

	s.outputService.SetSpinnerMessage("Please wait while data is being fetched...")
}
