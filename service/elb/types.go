package elb

import (
	"context"

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// ClientAPI is the interface for the AWS ELB client methods used by the service.
type ClientAPI interface {
	DescribeLoadBalancers(ctx context.Context, params *elb.DescribeLoadBalancersInput, optFns ...func(*elb.Options)) (*elb.DescribeLoadBalancersOutput, error)
	DescribeTargetGroups(ctx context.Context, params *elb.DescribeTargetGroupsInput, optFns ...func(*elb.Options)) (*elb.DescribeTargetGroupsOutput, error)
}

type service struct {
	client ClientAPI
}

// Service defines the interface for AWS ELB service.
type Service interface {
	GetUnusedLoadBalancers(ctx context.Context) ([]types.LoadBalancer, error)
}
