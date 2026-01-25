// Package elb provides a service for interacting with AWS Elastic Load Balancing.
package elb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// NewService creates a new ELB service.
func NewService(awsconfig aws.Config) *service {
	client := elb.NewFromConfig(awsconfig)
	return &service{
		client: client,
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
