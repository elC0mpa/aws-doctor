package sagemaker

import (
	"context"

	sm "github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/elC0mpa/aws-doctor/model"
)

// ClientAPI is the subset of the AWS SageMaker client used by this service.
type ClientAPI interface {
	ListEndpoints(ctx context.Context, params *sm.ListEndpointsInput, optFns ...func(*sm.Options)) (*sm.ListEndpointsOutput, error)
	DescribeEndpoint(ctx context.Context, params *sm.DescribeEndpointInput, optFns ...func(*sm.Options)) (*sm.DescribeEndpointOutput, error)
	DescribeEndpointConfig(ctx context.Context, params *sm.DescribeEndpointConfigInput, optFns ...func(*sm.Options)) (*sm.DescribeEndpointConfigOutput, error)
}

type service struct {
	client         ClientAPI
	cwService      cloudWatchMetricsService
	pricingService pricingService
}

// cloudWatchMetricsService is the narrow interface on the CloudWatch metrics service that this
// package depends on.
type cloudWatchMetricsService interface {
	SageMakerVariantInvocations(ctx context.Context, endpointName, variantName string, days int) (float64, error)
}

// pricingService is a local interface for the pricing dependency.
type pricingService interface {
	CalculateSageMakerEndpointMonthlyCost(variants []model.SageMakerVariant) float64
}

// Service is the interface for SageMaker waste detection.
type Service interface {
	GetIdleEndpoints(ctx context.Context, idleDays int) ([]model.IdleSageMakerEndpointInfo, error)
}
