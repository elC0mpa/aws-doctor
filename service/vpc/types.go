package vpc

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elC0mpa/aws-doctor/model"
	awscloudwatchmetrics "github.com/elC0mpa/aws-doctor/service/cloudwatchmetrics"
)

// ClientAPI is the interface for the AWS EC2 client methods used by the VPC service.
type ClientAPI interface {
	DescribeNatGateways(ctx context.Context, params *ec2.DescribeNatGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error)
}

// Service is the interface for VPC-related waste detection
type Service interface {
	IdleNATGateways(ctx context.Context, idleDays int) ([]model.NATGatewayWasteInfo, error)
}

type service struct {
	client    ClientAPI
	cwService awscloudwatchmetrics.Service
}

// natGatewayWithIndex pairs a NAT Gateway with its position in the original slice.
type natGatewayWithIndex struct {
	index      int
	natGateway types.NatGateway
}
