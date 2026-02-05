package awsinterfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/stretchr/testify/mock"
)

// MockEC2Client is a mock of EC2ClientAPI
type MockEC2Client struct {
	mock.Mock
}

// DescribeAddresses mocks the DescribeAddresses API call.
func (m *MockEC2Client) DescribeAddresses(ctx context.Context, params *ec2.DescribeAddressesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeAddressesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ec2.DescribeAddressesOutput), args.Error(1)
}

// DescribeNetworkInterfaces mocks the DescribeNetworkInterfaces API call.
func (m *MockEC2Client) DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ec2.DescribeNetworkInterfacesOutput), args.Error(1)
}

// DescribeVolumes mocks the DescribeVolumes API call.
func (m *MockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ec2.DescribeVolumesOutput), args.Error(1)
}

// DescribeInstances mocks the DescribeInstances API call.
func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ec2.DescribeInstancesOutput), args.Error(1)
}

// DescribeReservedInstances mocks the DescribeReservedInstances API call.
func (m *MockEC2Client) DescribeReservedInstances(ctx context.Context, params *ec2.DescribeReservedInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeReservedInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ec2.DescribeReservedInstancesOutput), args.Error(1)
}

// DescribeImages mocks the DescribeImages API call.
func (m *MockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ec2.DescribeImagesOutput), args.Error(1)
}

// DescribeSnapshots mocks the DescribeSnapshots API call.
func (m *MockEC2Client) DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ec2.DescribeSnapshotsOutput), args.Error(1)
}

// DescribeKeyPairs mocks the DescribeKeyPairs API call.
func (m *MockEC2Client) DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ec2.DescribeKeyPairsOutput), args.Error(1)
}
