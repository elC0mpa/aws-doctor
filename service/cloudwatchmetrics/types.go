package cloudwatchmetrics

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// ClientAPI is the interface for the AWS CloudWatch client methods used by the service.
type ClientAPI interface {
	GetMetricStatistics(ctx context.Context, params *cloudwatch.GetMetricStatisticsInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricStatisticsOutput, error)
}

// Service is the interface for CloudWatch metrics operations.
type Service interface {
	RDSHasZeroConnectionsInPeriod(ctx context.Context, dbInstanceID string, days int) (bool, error)
	NatGatewayBytesOut(ctx context.Context, natGatewayID string, days int) (float64, error)
	ELBHasZeroRequestsInPeriod(ctx context.Context, loadBalancerArn string, lbType elbtypes.LoadBalancerTypeEnum, days int) (bool, error)
}

type service struct {
	client ClientAPI
}
