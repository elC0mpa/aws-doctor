// Package orchestrator coordinates the execution of various AWS service checks.
package orchestrator

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/elC0mpa/aws-doctor/model"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	"github.com/elC0mpa/aws-doctor/utils/slice"
	"github.com/google/go-github/v62/github"
	"golang.org/x/sync/errgroup"
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

	if flags.Waste {
		return s.wasteWorkflow(flags.WasteChecks, flags.Report, flags.ReportPath)
	}

	if flags.Trend {
		return s.trendWorkflow(flags.TrendChecks, flags.Report, flags.ReportPath)
	}

	return s.defaultWorkflow(flags.Report, flags.ReportPath)
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

func (s *service) wasteWorkflow(wasteChecks []string, generateReport bool, reportPath string) error {
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
		// Fetch unused Load Balancers concurrently
		g.Go(func() error {
			var err error

			unusedLoadBalancers, err = s.elbService.GetUnusedLoadBalancers(ctx)

			return err
		})

		// Fetch idle Load Balancers concurrently
		g.Go(func() error {
			var err error

			idleLoadBalancers, err = s.elbService.GetIdleLoadBalancers(ctx)

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
		AccountID:           *stsResult.Account,
		ElasticIPs:          elasticIPInfo,
		UnusedVolumes:       availableEBSVolumesInfo,
		StoppedVolumes:      attachedToStoppedInstancesEBSVolumesInfo,
		Ris:                 expireReservedInstancesInfo,
		StoppedInstances:    stoppedInstancesMoreThan30Days,
		LoadBalancers:       unusedLoadBalancers,
		UnusedAMIs:          unusedAMIs,
		OrphanedSnapshots:   orphanedSnapshots,
		UnusedKeyPairs:      unusedKeyPairs,
		S3Buckets:           s3Buckets,
		S3MultipartUploads:  s3MultipartUploads,
		CloudWatchLogGroups: cloudwatchLogs,
		RDSInstances:        rdsInstances,
		RDSSnapshots:        rdsSnapshots,
		RDSIdleInstances:    rdsIdleInstances,
		IdleNATGateways:     idleNATGateways,
		IdleLoadBalancers:   idleLoadBalancers,
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
