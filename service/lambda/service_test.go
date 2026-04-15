package lambda

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCWLogsService struct {
	mock.Mock
}

func (m *mockCWLogsService) GetMaxMemoryUsed(ctx context.Context, logGroupName string, startTime, endTime time.Time) (int32, error) {
	args := m.Called(ctx, logGroupName, startTime, endTime)

	return args.Get(0).(int32), args.Error(1)
}

func TestGetOverProvisionedFunctions(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsService := new(mockCWLogsService)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsService:  mockLogsService,
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

	// Mock GetMaxMemoryUsed for over-provisioned function (uses 50 MB of 1024 MB = ~5%)
	mockLogsService.On("GetMaxMemoryUsed", mock.Anything, "/aws/lambda/over-provisioned-fn", mock.Anything, mock.Anything).Return(int32(50), nil)

	// Mock GetMaxMemoryUsed for normal function (uses 200 MB of 256 MB = ~78%)
	mockLogsService.On("GetMaxMemoryUsed", mock.Anything, "/aws/lambda/normal-fn", mock.Anything, mock.Anything).Return(int32(200), nil)

	result, err := s.GetOverProvisionedFunctions(context.Background(), 10)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "over-provisioned-fn", result[0].FunctionName)
	assert.Equal(t, int32(1024), result[0].ConfiguredMemoryMB)
	assert.Equal(t, int32(50), result[0].MaxMemoryUsedMB)
	assert.InDelta(t, 4.88, result[0].MemoryUtilization, 0.1)
	assert.Equal(t, int32(128), result[0].RecommendedMemoryMB) // 50*2=100, but minimum is 128
	assert.Equal(t, "nodejs20.x", result[0].Runtime)

	mockLambdaClient.AssertExpectations(t)
	mockLogsService.AssertExpectations(t)
}

func TestGetOverProvisionedFunctions_ListFunctionsError(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsService := new(mockCWLogsService)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsService:  mockLogsService,
	}

	mockLambdaClient.On("ListFunctions", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("Lambda API error"))

	_, err := s.GetOverProvisionedFunctions(context.Background(), 10)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list Lambda functions")
	mockLambdaClient.AssertExpectations(t)
}

func TestGetOverProvisionedFunctions_LogsError(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsService := new(mockCWLogsService)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsService:  mockLogsService,
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
	mockLogsService.On("GetMaxMemoryUsed", mock.Anything, "/aws/lambda/fn-with-no-logs", mock.Anything, mock.Anything).Return(int32(0), errors.New("log group not found"))

	result, err := s.GetOverProvisionedFunctions(context.Background(), 10)

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockLambdaClient.AssertExpectations(t)
	mockLogsService.AssertExpectations(t)
}

func TestGetOverProvisionedFunctions_NoFunctions(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsService := new(mockCWLogsService)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsService:  mockLogsService,
	}

	mockLambdaClient.On("ListFunctions", mock.Anything, mock.Anything, mock.Anything).Return(&awslambda.ListFunctionsOutput{
		Functions: []lambdatypes.FunctionConfiguration{},
	}, nil)

	result, err := s.GetOverProvisionedFunctions(context.Background(), 10)

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockLambdaClient.AssertExpectations(t)
}
