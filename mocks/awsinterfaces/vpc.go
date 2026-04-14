package awsinterfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/stretchr/testify/mock"
)

// MockVPCClient mocks the VPC client for testing.
type MockVPCClient struct {
	mock.Mock
}

// DescribeNatGateways mocks the DescribeNatGateways method.
func (m *MockVPCClient) DescribeNatGateways(ctx context.Context, params *ec2.DescribeNatGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ec2.DescribeNatGatewaysOutput), args.Error(1)
}
