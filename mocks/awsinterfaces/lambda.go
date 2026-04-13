package awsinterfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/stretchr/testify/mock"
)

// MockLambdaClient is a mock of ClientAPI.
type MockLambdaClient struct {
	mock.Mock
}

// ListFunctions mocks the ListFunctions API call.
func (m *MockLambdaClient) ListFunctions(ctx context.Context, params *awslambda.ListFunctionsInput, optFns ...func(*awslambda.Options)) (*awslambda.ListFunctionsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*awslambda.ListFunctionsOutput), args.Error(1)
}

// MockLambdaLogsClient is a mock of LogsClientAPI for the Lambda service.
type MockLambdaLogsClient struct {
	mock.Mock
}

// FilterLogEvents mocks the FilterLogEvents API call.
func (m *MockLambdaLogsClient) FilterLogEvents(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*cloudwatchlogs.FilterLogEventsOutput), args.Error(1)
}
