// Package elb provides a service for interacting with AWS Elastic Load Balancing.
package elb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/pricing"
)

const idleCheckDays = 7

// NewService creates a new ELB service.
func NewService(awsconfig aws.Config, cwService cloudwatchMetricsService) Service {
	client := elb.NewFromConfig(awsconfig)

	return &service{
		client:    client,
		cwService: cwService,
	}
}

func (s *service) GetUnusedLoadBalancers(ctx context.Context) ([]types.LoadBalancer, error) {
	// Collect all load balancers using pagination
	var allLoadBalancers []types.LoadBalancer

	lbPaginator := elb.NewDescribeLoadBalancersPaginator(s.client, &elb.DescribeLoadBalancersInput{})

	for lbPaginator.HasMorePages() {
		lbOutput, err := lbPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		allLoadBalancers = append(allLoadBalancers, lbOutput.LoadBalancers...)
	}

	// Collect all target groups using pagination
	usedLbArns := make(map[string]bool)
	tgPaginator := elb.NewDescribeTargetGroupsPaginator(s.client, &elb.DescribeTargetGroupsInput{})

	for tgPaginator.HasMorePages() {
		tgOutput, err := tgPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, tg := range tgOutput.TargetGroups {
			for _, lbArn := range tg.LoadBalancerArns {
				usedLbArns[lbArn] = true
			}
		}
	}

	// Find orphaned load balancers
	var orphanedLbs []types.LoadBalancer

	for _, lb := range allLoadBalancers {
		if lb.Type != types.LoadBalancerTypeEnumApplication && lb.Type != types.LoadBalancerTypeEnumNetwork {
			continue
		}

		arn := aws.ToString(lb.LoadBalancerArn)

		if !usedLbArns[arn] {
			orphanedLbs = append(orphanedLbs, lb)
		}
	}

	return orphanedLbs, nil
}

func (s *service) GetIdleLoadBalancers(ctx context.Context) ([]model.ELBIdleInfo, error) {
	// Collect all load balancers using pagination
	var allLoadBalancers []types.LoadBalancer

	lbPaginator := elb.NewDescribeLoadBalancersPaginator(s.client, &elb.DescribeLoadBalancersInput{})

	for lbPaginator.HasMorePages() {
		lbOutput, err := lbPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		allLoadBalancers = append(allLoadBalancers, lbOutput.LoadBalancers...)
	}

	// Collect all target groups to find which LBs have target groups
	lbsWithTargetGroups := make(map[string]bool)
	tgPaginator := elb.NewDescribeTargetGroupsPaginator(s.client, &elb.DescribeTargetGroupsInput{})

	for tgPaginator.HasMorePages() {
		tgOutput, err := tgPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, tg := range tgOutput.TargetGroups {
			for _, lbArn := range tg.LoadBalancerArns {
				lbsWithTargetGroups[lbArn] = true
			}
		}
	}

	// Check LBs that have target groups but may have zero traffic
	var idleLBs []model.ELBIdleInfo

	for _, lb := range allLoadBalancers {
		if lb.Type != types.LoadBalancerTypeEnumApplication && lb.Type != types.LoadBalancerTypeEnumNetwork {
			continue
		}

		arn := aws.ToString(lb.LoadBalancerArn)

		// Only check LBs that DO have target groups (those without are caught by GetUnusedLoadBalancers)
		if !lbsWithTargetGroups[arn] {
			continue
		}

		lbType := string(lb.Type)

		idle, err := s.cwService.ELBHasZeroRequestsInPeriod(ctx, arn, lbType, idleCheckDays)
		if err != nil {
			return nil, err
		}

		if idle {
			idleLBs = append(idleLBs, model.ELBIdleInfo{
				LoadBalancerName:     aws.ToString(lb.LoadBalancerName),
				LoadBalancerArn:      arn,
				Type:                 lbType,
				DaysChecked:          idleCheckDays,
				EstimatedMonthlyCost: pricing.CalculateLoadBalancerMonthlyCost(lb.Type),
			})
		}
	}

	return idleLBs, nil
}
