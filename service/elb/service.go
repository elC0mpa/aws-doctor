// Package elb provides a service for interacting with AWS Elastic Load Balancing.
package elb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"golang.org/x/sync/errgroup"
)

// NewService creates a new ELB service.
func NewService(awsconfig aws.Config, cwService cloudwatchMetricsService, pricingSvc pricingService) Service {
	client := elb.NewFromConfig(awsconfig)

	return &service{
		client:         client,
		cwService:      cwService,
		pricingService: pricingSvc,
	}
}

func (s *service) Name() string {
	return "elb"
}

func (s *service) Analyze(ctx context.Context, flags model.Flags) (model.ScopeResult, error) {
	start := time.Now()
	input := model.RenderWasteInput{}

	var errs []error

	unusedLBs, idleLBs, err := s.GetLoadBalancerWaste(ctx, flags.ELBIdleDays)
	if err != nil {
		errs = append(errs, err)
	} else {
		input.LoadBalancers = unusedLBs
		input.IdleLoadBalancers = idleLBs
	}

	var finalErr error
	if len(errs) > 0 {
		finalErr = fmt.Errorf("elb analyze errors: %v", errs)
	}

	return model.ScopeResult{
		Scope:    s.Name(),
		Input:    input,
		Duration: time.Since(start),
		Err:      finalErr,
	}, nil
}

// fetchLoadBalancersAndTargetGroups paginates all load balancers and target groups,
// returning the full LB list and a set of LB ARNs that have at least one target group.
func (s *service) fetchLoadBalancersAndTargetGroups(ctx context.Context) ([]types.LoadBalancer, map[string]bool, error) {
	var allLoadBalancers []types.LoadBalancer

	lbPaginator := elb.NewDescribeLoadBalancersPaginator(s.client, &elb.DescribeLoadBalancersInput{})

	for lbPaginator.HasMorePages() {
		lbOutput, err := lbPaginator.NextPage(ctx)
		if err != nil {
			return nil, nil, err
		}

		allLoadBalancers = append(allLoadBalancers, lbOutput.LoadBalancers...)
	}

	usedLbArns := make(map[string]bool)
	tgPaginator := elb.NewDescribeTargetGroupsPaginator(s.client, &elb.DescribeTargetGroupsInput{})

	for tgPaginator.HasMorePages() {
		tgOutput, err := tgPaginator.NextPage(ctx)
		if err != nil {
			return nil, nil, err
		}

		for _, tg := range tgOutput.TargetGroups {
			for _, lbArn := range tg.LoadBalancerArns {
				usedLbArns[lbArn] = true
			}
		}
	}

	return allLoadBalancers, usedLbArns, nil
}

// GetLoadBalancerWaste fetches all load balancers and target groups once, then partitions into
// unused (no target groups) and idle (has target groups but zero traffic via CloudWatch).
func (s *service) GetLoadBalancerWaste(ctx context.Context, idleDays int) ([]types.LoadBalancer, []model.ELBIdleInfo, error) {
	allLoadBalancers, usedLbArns, err := s.fetchLoadBalancersAndTargetGroups(ctx)
	if err != nil {
		return nil, nil, err
	}

	var (
		orphanedLbs []types.LoadBalancer
		candidates  []types.LoadBalancer
	)

	for _, lb := range allLoadBalancers {
		if lb.Type != types.LoadBalancerTypeEnumApplication && lb.Type != types.LoadBalancerTypeEnumNetwork {
			continue
		}

		arn := aws.ToString(lb.LoadBalancerArn)

		if !usedLbArns[arn] {
			orphanedLbs = append(orphanedLbs, lb)
		} else {
			candidates = append(candidates, lb)
		}
	}

	// Check CloudWatch metrics in parallel for LBs that have target groups
	var (
		mu      sync.Mutex
		idleLBs []model.ELBIdleInfo
	)

	g, ctx := errgroup.WithContext(ctx)

	for _, lb := range candidates {
		g.Go(func() error {
			arn := aws.ToString(lb.LoadBalancerArn)

			idle, cwErr := s.cwService.ELBHasZeroRequestsInPeriod(ctx, arn, lb.Type, idleDays)
			if cwErr != nil {
				return cwErr
			}

			if idle {
				mu.Lock()

				idleLBs = append(idleLBs, model.ELBIdleInfo{
					Name:                 aws.ToString(lb.LoadBalancerName),
					ARN:                  arn,
					Type:                 string(lb.Type),
					DaysChecked:          idleDays,
					EstimatedMonthlyCost: s.pricingService.CalculateLoadBalancerMonthlyCost(lb.Type),
				})
				mu.Unlock()
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, nil, err
	}

	return orphanedLbs, idleLBs, nil
}
