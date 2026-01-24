package orchestrator

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/elC0mpa/aws-doctor/model"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	awsec2 "github.com/elC0mpa/aws-doctor/service/ec2"
	"github.com/elC0mpa/aws-doctor/service/elb"
	"github.com/elC0mpa/aws-doctor/service/output"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
	"golang.org/x/sync/errgroup"
)

func NewService(stsService awssts.STSService, costService awscostexplorer.CostService, ec2Service awsec2.EC2Service, elbService elb.ELBService, outputService output.Service) *service {
	return &service{
		stsService:    stsService,
		costService:   costService,
		ec2Service:    ec2Service,
		elbService:    elbService,
		outputService: outputService,
	}
}

func (s *service) Orchestrate(flags model.Flags) error {
	if flags.Waste {
		return s.wasteWorkflow()
	}

	if flags.Trend {
		return s.trendWorkflow()
	}

	return s.defaultWorkflow()
}

func (s *service) defaultWorkflow() error {
	currentMonthData, err := s.costService.GetCurrentMonthCostsByService(context.Background())
	if err != nil {
		return err
	}

	lastMonthData, err := s.costService.GetLastMonthCostsByService(context.Background())
	if err != nil {
		return err
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

	return s.outputService.RenderCostComparison(*stsResult.Account, *lastTotalCost, *currentTotalCost, lastMonthData, currentMonthData)
}

func (s *service) trendWorkflow() error {
	costInfo, err := s.costService.GetLastSixMonthsCosts(context.Background())
	if err != nil {
		return err
	}

	stsResult, err := s.stsService.GetCallerIdentity(context.Background())
	if err != nil {
		return err
	}

	s.outputService.StopSpinner()

	return s.outputService.RenderTrend(*stsResult.Account, costInfo)
}

func (s *service) wasteWorkflow() error {
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)

	// Results from concurrent API calls
	var elasticIpInfo []types.Address
	var availableEBSVolumesInfo []types.Volume
	var stoppedInstancesMoreThan30Days []types.Instance
	var attachedToStoppedInstancesEBSVolumesInfo []types.Volume
	var expireReservedInstancesInfo []model.RiExpirationInfo
	var unusedLoadBalancers []elbtypes.LoadBalancer
	var unusedAMIs []model.AMIWasteInfo
	var orphanedSnapshots []model.SnapshotWasteInfo
	var stsResult *sts.GetCallerIdentityOutput

	// Fetch unused Elastic IPs concurrently
	g.Go(func() error {
		var err error
		elasticIpInfo, err = s.ec2Service.GetUnusedElasticIPAddressesInfo(ctx)
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

	// Fetch unused Load Balancers concurrently
	g.Go(func() error {
		var err error
		unusedLoadBalancers, err = s.elbService.GetUnusedLoadBalancers(ctx)
		return err
	})

	// Fetch caller identity concurrently
	g.Go(func() error {
		var err error
		stsResult, err = s.stsService.GetCallerIdentity(ctx)
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

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return err
	}

	s.outputService.StopSpinner()

	return s.outputService.RenderWaste(
		*stsResult.Account,
		elasticIpInfo,
		availableEBSVolumesInfo,
		attachedToStoppedInstancesEBSVolumesInfo,
		expireReservedInstancesInfo,
		stoppedInstancesMoreThan30Days,
		unusedLoadBalancers,
		unusedAMIs,
		orphanedSnapshots,
	)
}
