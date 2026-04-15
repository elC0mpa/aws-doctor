package awsinterfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/stretchr/testify/mock"
)

// MockCloudWatchLogsClient is a mock of CloudWatchLogs ClientAPI
type MockCloudWatchLogsClient struct {
	mock.Mock
}

// DescribeLogGroups mocks the DescribeLogGroups API call.
func (m *MockCloudWatchLogsClient) DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*cloudwatchlogs.DescribeLogGroupsOutput), args.Error(1)
}

// StartQuery mocks the StartQuery API call.
func (m *MockCloudWatchLogsClient) StartQuery(ctx context.Context, params *cloudwatchlogs.StartQueryInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.StartQueryOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*cloudwatchlogs.StartQueryOutput), args.Error(1)
}

// GetQueryResults mocks the GetQueryResults API call.
func (m *MockCloudWatchLogsClient) GetQueryResults(ctx context.Context, params *cloudwatchlogs.GetQueryResultsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetQueryResultsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*cloudwatchlogs.GetQueryResultsOutput), args.Error(1)
}
