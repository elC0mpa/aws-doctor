package cloudwatchmetrics

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// ClientAPI is the interface for the AWS CloudWatch client methods used by the service.
type ClientAPI interface {
	GetMetricStatistics(ctx context.Context, params *cloudwatch.GetMetricStatisticsInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricStatisticsOutput, error)
	GetMetricData(ctx context.Context, params *cloudwatch.GetMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricDataOutput, error)
}

type service struct {
	client ClientAPI
}

// Service is the interface for AWS CloudWatch metrics service.
type Service interface {
	RDSHasZeroConnectionsInPeriod(ctx context.Context, dbInstanceID string, days int) (bool, error)
	NATGatewayBytesOut(ctx context.Context, natGatewayID string, days int) (float64, error)
	ELBHasZeroRequestsInPeriod(ctx context.Context, loadBalancerArn string, lbType elbtypes.LoadBalancerTypeEnum, days int) (bool, error)
	SageMakerVariantInvocations(ctx context.Context, endpointName, variantName string, days int) (float64, error)
	EC2InstanceIdleStats(ctx context.Context, instanceID string, days int) (cpuAvgPercent, networkBytesPerDay float64, err error)
}
