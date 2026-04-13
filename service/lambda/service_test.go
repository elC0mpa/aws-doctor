package lambda

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwlogstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetOverProvisionedFunctions(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsClient := new(awsinterfaces.MockLambdaLogsClient)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsClient:   mockLogsClient,
	}

	// Mock ListFunctions: one over-provisioned function, one normal function
	mockLambdaClient.On("ListFunctions", mock.Anything, mock.Anything, mock.Anything).Return(&awslambda.ListFunctionsOutput{
		Functions: []lambdatypes.FunctionConfiguration{
			{
				FunctionName: aws.String("over-provisioned-fn"),
				MemorySize:   aws.Int32(1024),
				Runtime:      lambdatypes.RuntimeNodejs20x,
			},
			{
				FunctionName: aws.String("normal-fn"),
				MemorySize:   aws.Int32(256),
				Runtime:      lambdatypes.RuntimePython312,
			},
		},
	}, nil)

	// Mock FilterLogEvents for over-provisioned function (uses 50 MB of 1024 MB = ~5%)
	mockLogsClient.On("FilterLogEvents", mock.Anything, mock.MatchedBy(func(input *cloudwatchlogs.FilterLogEventsInput) bool {
		return aws.ToString(input.LogGroupName) == "/aws/lambda/over-provisioned-fn"
	}), mock.Anything).Return(&cloudwatchlogs.FilterLogEventsOutput{
		Events: []cwlogstypes.FilteredLogEvent{
			{Message: aws.String("REPORT RequestId: abc-123 Duration: 100.00 ms Billed Duration: 100 ms Memory Size: 1024 MB Max Memory Used: 50 MB")},
			{Message: aws.String("REPORT RequestId: def-456 Duration: 200.00 ms Billed Duration: 200 ms Memory Size: 1024 MB Max Memory Used: 40 MB")},
		},
	}, nil)

	// Mock FilterLogEvents for normal function (uses 200 MB of 256 MB = ~78%)
	mockLogsClient.On("FilterLogEvents", mock.Anything, mock.MatchedBy(func(input *cloudwatchlogs.FilterLogEventsInput) bool {
		return aws.ToString(input.LogGroupName) == "/aws/lambda/normal-fn"
	}), mock.Anything).Return(&cloudwatchlogs.FilterLogEventsOutput{
		Events: []cwlogstypes.FilteredLogEvent{
			{Message: aws.String("REPORT RequestId: ghi-789 Duration: 300.00 ms Billed Duration: 300 ms Memory Size: 256 MB Max Memory Used: 200 MB")},
		},
	}, nil)

	result, err := s.GetOverProvisionedFunctions(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "over-provisioned-fn", result[0].FunctionName)
	assert.Equal(t, int32(1024), result[0].ConfiguredMemoryMB)
	assert.Equal(t, int32(50), result[0].MaxMemoryUsedMB)
	assert.InDelta(t, 4.88, result[0].MemoryUtilization, 0.1)
	assert.Equal(t, int32(128), result[0].RecommendedMemoryMB) // 50*2=100, but minimum is 128
	assert.Equal(t, "nodejs20.x", result[0].Runtime)

	mockLambdaClient.AssertExpectations(t)
	mockLogsClient.AssertExpectations(t)
}

func TestGetOverProvisionedFunctions_ListFunctionsError(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsClient := new(awsinterfaces.MockLambdaLogsClient)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsClient:   mockLogsClient,
	}

	mockLambdaClient.On("ListFunctions", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("Lambda API error"))

	_, err := s.GetOverProvisionedFunctions(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list Lambda functions")
	mockLambdaClient.AssertExpectations(t)
}

func TestGetOverProvisionedFunctions_LogsError(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsClient := new(awsinterfaces.MockLambdaLogsClient)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsClient:   mockLogsClient,
	}

	mockLambdaClient.On("ListFunctions", mock.Anything, mock.Anything, mock.Anything).Return(&awslambda.ListFunctionsOutput{
		Functions: []lambdatypes.FunctionConfiguration{
			{
				FunctionName: aws.String("fn-with-no-logs"),
				MemorySize:   aws.Int32(512),
				Runtime:      lambdatypes.RuntimePython312,
			},
		},
	}, nil)

	// Logs error should be skipped, not returned
	mockLogsClient.On("FilterLogEvents", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("log group not found"))

	result, err := s.GetOverProvisionedFunctions(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockLambdaClient.AssertExpectations(t)
	mockLogsClient.AssertExpectations(t)
}

func TestGetOverProvisionedFunctions_NoFunctions(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsClient := new(awsinterfaces.MockLambdaLogsClient)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsClient:   mockLogsClient,
	}

	mockLambdaClient.On("ListFunctions", mock.Anything, mock.Anything, mock.Anything).Return(&awslambda.ListFunctionsOutput{
		Functions: []lambdatypes.FunctionConfiguration{},
	}, nil)

	result, err := s.GetOverProvisionedFunctions(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockLambdaClient.AssertExpectations(t)
}
