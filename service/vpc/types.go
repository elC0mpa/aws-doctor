package vpc

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/analyzer"
)

// ClientAPI is the interface for the AWS EC2 client methods used by the VPC service.
type ClientAPI interface {
	DescribeNatGateways(ctx context.Context, params *ec2.DescribeNatGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error)
}

type service struct {
	client         ClientAPI
	cwService      cloudwatchMetricsService
	pricingService pricingService
}

// cloudwatchMetricsService is a local interface for the CloudWatch metrics dependency.
type cloudwatchMetricsService interface {
	NATGatewayBytesOut(ctx context.Context, natGatewayID string, days int) (float64, error)
}

// pricingService is a local interface for the pricing dependency.
type pricingService interface {
	CalculateNATGatewayMonthlyCost() float64
}

// Service is the interface for AWS VPC service.
type Service interface {
	analyzer.WasteAnalyzer
	GetIdleNATGateways(ctx context.Context, idleDays int) ([]model.NATGatewayWasteInfo, error)
}

// natGatewayMetricResult stores the result of a CloudWatch metric query for a NAT Gateway.
type natGatewayMetricResult struct {
	bytesOut float64
	err      error
}
