package elb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

func NewService(awsconfig aws.Config) *service {
	client := elb.NewFromConfig(awsconfig)
	return &service{
		client: client,
	}
}

func (s *service) GetUnusedLoadBalancers(ctx context.Context) ([]types.LoadBalancer, error) {
	lbOutput, err := s.client.DescribeLoadBalancers(ctx, &elb.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, err
	}

	tgOutput, err := s.client.DescribeTargetGroups(ctx, &elb.DescribeTargetGroupsInput{})
	if err != nil {
		return nil, err
	}

	usedLbArns := make(map[string]bool)

	for _, tg := range tgOutput.TargetGroups {
		for _, lbArn := range tg.LoadBalancerArns {
			usedLbArns[lbArn] = true
		}
	}

	var orphanedLbs []types.LoadBalancer

	for _, lb := range lbOutput.LoadBalancers {
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
