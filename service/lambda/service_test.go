package lambda

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
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
	// GetLambdaMaxMemoryUsedBatchFn, when set, is called instead of the testify mock — useful
	// when a test needs per-call dynamic returns (e.g., per-batch contents).
	GetLambdaMaxMemoryUsedBatchFn func(ctx context.Context, logGroupNames []string, startTime, endTime time.Time) (map[string]int32, error)
}

func (m *mockCWLogsService) GetLambdaMaxMemoryUsedBatch(ctx context.Context, logGroupNames []string, startTime, endTime time.Time) (map[string]int32, error) {
	if m.GetLambdaMaxMemoryUsedBatchFn != nil {
		return m.GetLambdaMaxMemoryUsedBatchFn(ctx, logGroupNames, startTime, endTime)
	}

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

	result, err := s.GetOverProvisionedFunctions(context.Background(), 10, 14)

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

	_, err := s.GetOverProvisionedFunctions(context.Background(), 10, 14)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list Lambda functions")
	mockLambdaClient.AssertExpectations(t)
}

func TestGetOverProvisionedFunctions_BatchErrorPropagates(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsService := new(mockCWLogsService)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsService:  mockLogsService,
	}

	mockLambdaClient.On("ListFunctions", mock.Anything, mock.Anything, mock.Anything).Return(&awslambda.ListFunctionsOutput{
		Functions: []lambdatypes.FunctionConfiguration{
			{
				FunctionName: aws.String("fn-a"),
				MemorySize:   aws.Int32(512),
				Runtime:      lambdatypes.RuntimePython312,
			},
		},
	}, nil)

	mockLogsService.On("ListExistingLogGroups", mock.Anything, "/aws/lambda/").Return(map[string]struct{}{
		"/aws/lambda/fn-a": {},
	}, nil)

	mockLogsService.On("GetLambdaMaxMemoryUsedBatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return((map[string]int32)(nil), errors.New("insights throttled"))

	_, err := s.GetOverProvisionedFunctions(context.Background(), 10, 14)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insights throttled")
	mockLambdaClient.AssertExpectations(t)
	mockLogsService.AssertExpectations(t)
}

func TestGetOverProvisionedFunctions_ListLogGroupsErrorPropagates(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsService := new(mockCWLogsService)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsService:  mockLogsService,
	}

	mockLambdaClient.On("ListFunctions", mock.Anything, mock.Anything, mock.Anything).Return(&awslambda.ListFunctionsOutput{
		Functions: []lambdatypes.FunctionConfiguration{
			{FunctionName: aws.String("fn-a"), MemorySize: aws.Int32(512)},
		},
	}, nil)

	mockLogsService.On("ListExistingLogGroups", mock.Anything, "/aws/lambda/").Return((map[string]struct{})(nil), errors.New("access denied"))

	_, err := s.GetOverProvisionedFunctions(context.Background(), 10, 14)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list Lambda log groups")
}

func TestGetOverProvisionedFunctions_MissingLogGroupsAreSkipped(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsService := new(mockCWLogsService)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsService:  mockLogsService,
	}

	mockLambdaClient.On("ListFunctions", mock.Anything, mock.Anything, mock.Anything).Return(&awslambda.ListFunctionsOutput{
		Functions: []lambdatypes.FunctionConfiguration{
			{FunctionName: aws.String("fn-present"), MemorySize: aws.Int32(1024), Runtime: lambdatypes.RuntimePython312},
			{FunctionName: aws.String("fn-never-invoked"), MemorySize: aws.Int32(1024), Runtime: lambdatypes.RuntimePython312},
		},
	}, nil)

	mockLogsService.On("ListExistingLogGroups", mock.Anything, "/aws/lambda/").Return(map[string]struct{}{
		"/aws/lambda/fn-present": {},
	}, nil)

	capturedBatch := mock.MatchedBy(func(groups []string) bool {
		return len(groups) == 1 && groups[0] == "/aws/lambda/fn-present"
	})

	mockLogsService.On("GetLambdaMaxMemoryUsedBatch", mock.Anything, capturedBatch, mock.Anything, mock.Anything).Return(map[string]int32{
		"/aws/lambda/fn-present": 50,
	}, nil)

	result, err := s.GetOverProvisionedFunctions(context.Background(), 10, 14)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "fn-present", result[0].FunctionName)
	mockLogsService.AssertExpectations(t)
}

func TestGetOverProvisionedFunctions_AllLogGroupsMissing(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsService := new(mockCWLogsService)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsService:  mockLogsService,
	}

	mockLambdaClient.On("ListFunctions", mock.Anything, mock.Anything, mock.Anything).Return(&awslambda.ListFunctionsOutput{
		Functions: []lambdatypes.FunctionConfiguration{
			{FunctionName: aws.String("fn-a"), MemorySize: aws.Int32(512)},
		},
	}, nil)

	mockLogsService.On("ListExistingLogGroups", mock.Anything, "/aws/lambda/").Return(map[string]struct{}{}, nil)

	result, err := s.GetOverProvisionedFunctions(context.Background(), 10, 14)

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockLogsService.AssertNotCalled(t, "GetLambdaMaxMemoryUsedBatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetOverProvisionedFunctions_ChunksAtFiftyAndMerges(t *testing.T) {
	mockLambdaClient := new(awsinterfaces.MockLambdaClient)
	mockLogsService := new(mockCWLogsService)

	s := &service{
		lambdaClient: mockLambdaClient,
		logsService:  mockLogsService,
	}

	const totalFunctions = 101

	functions := make([]lambdatypes.FunctionConfiguration, 0, totalFunctions)
	existing := make(map[string]struct{}, totalFunctions)

	for i := 0; i < totalFunctions; i++ {
		name := fmt.Sprintf("fn-%03d", i)
		functions = append(functions, lambdatypes.FunctionConfiguration{
			FunctionName: aws.String(name),
			MemorySize:   aws.Int32(1024),
			Runtime:      lambdatypes.RuntimePython312,
		})
		existing["/aws/lambda/"+name] = struct{}{}
	}

	mockLambdaClient.On("ListFunctions", mock.Anything, mock.Anything, mock.Anything).Return(&awslambda.ListFunctionsOutput{Functions: functions}, nil)
	mockLogsService.On("ListExistingLogGroups", mock.Anything, "/aws/lambda/").Return(existing, nil)

	var (
		mu         sync.Mutex
		batchSizes []int
	)

	mockLogsService.GetLambdaMaxMemoryUsedBatchFn = func(_ context.Context, groups []string, _, _ time.Time) (map[string]int32, error) {
		mu.Lock()

		batchSizes = append(batchSizes, len(groups))
		mu.Unlock()

		out := make(map[string]int32, len(groups))
		for _, g := range groups {
			out[g] = 50
		}

		return out, nil
	}

	result, err := s.GetOverProvisionedFunctions(context.Background(), 10, 14)

	assert.NoError(t, err)
	assert.Len(t, result, totalFunctions, "all functions across batches should be merged into the result")

	sort.Ints(batchSizes)
	assert.Equal(t, []int{1, 50, 50}, batchSizes, "expected batches of 50, 50, and 1")
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

	result, err := s.GetOverProvisionedFunctions(context.Background(), 10, 14)

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockLambdaClient.AssertExpectations(t)
}

func TestAnalyzerMethods(t *testing.T) {
	svc := &service{}

	if svc.Name() == "" {
		t.Error("Name() should not be empty")
	}

	if svc.TabName() == "" {
		t.Error("TabName() should not be empty")
	}
}
