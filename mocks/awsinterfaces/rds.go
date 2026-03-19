package awsinterfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/stretchr/testify/mock"
)

// MockRDSClient is a mock of RDS ClientAPI
type MockRDSClient struct {
	mock.Mock
}

// DescribeDBInstances mocks the DescribeDBInstances API call.
func (m *MockRDSClient) DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*rds.DescribeDBInstancesOutput), args.Error(1)
}

// DescribeDBSnapshots mocks the DescribeDBSnapshots API call.
func (m *MockRDSClient) DescribeDBSnapshots(ctx context.Context, params *rds.DescribeDBSnapshotsInput, optFns ...func(*rds.Options)) (*rds.DescribeDBSnapshotsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*rds.DescribeDBSnapshotsOutput), args.Error(1)
}
