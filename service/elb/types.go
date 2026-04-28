package elb

import (
	"context"

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
)

// ClientAPI is the interface for the AWS ELB client methods used by the service.
type ClientAPI interface {
	DescribeLoadBalancers(ctx context.Context, params *elb.DescribeLoadBalancersInput, optFns ...func(*elb.Options)) (*elb.DescribeLoadBalancersOutput, error)
	DescribeTargetGroups(ctx context.Context, params *elb.DescribeTargetGroupsInput, optFns ...func(*elb.Options)) (*elb.DescribeTargetGroupsOutput, error)
}

type service struct {
	client         ClientAPI
	cwService      cloudwatchMetricsService
	pricingService pricingService
}

// cloudwatchMetricsService is a local interface for the CloudWatch metrics dependency.
type cloudwatchMetricsService interface {
	ELBHasZeroRequestsInPeriod(ctx context.Context, loadBalancerArn string, lbType types.LoadBalancerTypeEnum, days int) (bool, error)
}

// pricingService is a local interface for the pricing dependency.
type pricingService interface {
	CalculateLoadBalancerMonthlyCost(lbType types.LoadBalancerTypeEnum) float64
}

// Service is the interface for AWS ELB service.
type Service interface {
	GetLoadBalancerWaste(ctx context.Context, idleDays int) ([]types.LoadBalancer, []model.ELBIdleInfo, error)
}
