package orchestrator

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	awsec2 "github.com/elC0mpa/aws-doctor/service/ec2"
	awssts "github.com/elC0mpa/aws-doctor/service/sts"
	"github.com/elC0mpa/aws-doctor/utils"
)

func NewService(stsService awssts.STSService, costService awscostexplorer.CostService, ec2Service awsec2.EC2Service) *service {
	return &service{
		stsService:  stsService,
		costService: costService,
		ec2Service:  ec2Service,
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

	utils.StopSpinner()

	utils.DrawCostTable(*stsResult.Account, *lastTotalCost, *currentTotalCost, lastMonthData, currentMonthData, "UnblendedCost")
	return nil
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

	utils.StopSpinner()

	utils.DrawTrendChart(*stsResult.Account, costInfo)

	return nil
}

func (s *service) wasteWorkflow() error {
	elasticIpInfo, err := s.ec2Service.GetUnusedElasticIpAddressesInfo(context.Background())
	if err != nil {
		return err
	}

	availableEBSVolumesInfo, err := s.ec2Service.GetUnusedEBSVolumes(context.Background())
	if err != nil {
		return err
	}

	stoppedInstancesMoreThan30Days, attachedToStoppedInstancesEBSVolumesInfo, err := s.ec2Service.GetStoppedInstancesInfo(context.Background())
	if err != nil {
		return err
	}

	expireReservedInstancesInfo, err := s.ec2Service.GetReservedInstanceExpiringOrExpired30DaysWaste(context.Background())
	if err != nil {
		return err
	}

	stsResult, err := s.stsService.GetCallerIdentity(context.Background())
	if err != nil {
		return err
	}

	utils.StopSpinner()

	utils.DrawWasteTable(*stsResult.Account, elasticIpInfo, availableEBSVolumesInfo, attachedToStoppedInstancesEBSVolumesInfo, expireReservedInstancesInfo, stoppedInstancesMoreThan30Days)

	return nil
}
