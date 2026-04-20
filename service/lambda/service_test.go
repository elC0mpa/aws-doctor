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

func (m *mockCWLogsService) GetLambdaMaxMemoryUsedBatch(ctx context.Context, logGroupNames []string, startTime, endTime time.Time) (map[string]int32, error) {
	args := m.Called(ctx, logGroupNames, startTime, endTime)

	res, _ := args.Get(0).(map[string]int32)

	return res, args.Error(1)
}

func (m *mockCWLogsService) ListExistingLogGroups(ctx context.Context, prefix string) (map[string]struct{}, error) {
	args := m.Called(ctx, prefix)

	res, _ := args.Get(0).(map[string]struct{})

	return res, args.Error(1)
}

func TestGetOverProvisionedFunctions(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsService := new(mockCWLogsService)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsService:  mockLogsService,
	}

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

	mockLogsService.On("ListExistingLogGroups", mock.Anything, "/aws/lambda/").Return(map[string]struct{}{
		"/aws/lambda/over-provisioned-fn": {},
		"/aws/lambda/normal-fn":           {},
	}, nil)

	mockLogsService.On("GetLambdaMaxMemoryUsedBatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]int32{
		"/aws/lambda/over-provisioned-fn": 50,
		"/aws/lambda/normal-fn":           200,
	}, nil)

	result, err := s.GetOverProvisionedFunctions(context.Background(), 10)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "over-provisioned-fn", result[0].FunctionName)
	assert.Equal(t, int32(1024), result[0].ConfiguredMemoryMB)
	assert.Equal(t, int32(50), result[0].MaxMemoryUsedMB)
	assert.InDelta(t, 4.88, result[0].MemoryUtilization, 0.1)
	assert.Equal(t, int32(128), result[0].RecommendedMemoryMB)
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

func TestGetOverProvisionedFunctions_LogsBatchErrorIsSkipped(t *testing.T) {
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

	mockLogsService.On("ListExistingLogGroups", mock.Anything, "/aws/lambda/").Return(map[string]struct{}{
		"/aws/lambda/fn-with-no-logs": {},
	}, nil)

	mockLogsService.On("GetLambdaMaxMemoryUsedBatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return((map[string]int32)(nil), errors.New("insights error"))

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
