package awsinterfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/stretchr/testify/mock"
)

// MockCloudWatchClient is a mock of CloudWatch ClientAPI
type MockCloudWatchClient struct {
	mock.Mock
}

// GetMetricStatistics mocks the GetMetricStatistics API call.
func (m *MockCloudWatchClient) GetMetricStatistics(ctx context.Context, params *cloudwatch.GetMetricStatisticsInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricStatisticsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*cloudwatch.GetMetricStatisticsOutput), args.Error(1)
}

// GetMetricData mocks the GetMetricData API call.
func (m *MockCloudWatchClient) GetMetricData(ctx context.Context, params *cloudwatch.GetMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricDataOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*cloudwatch.GetMetricDataOutput), args.Error(1)
}
