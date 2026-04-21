package services

import (
	"context"
	"time"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockCloudWatchLogsService is a mock of CloudWatchLogs Service
type MockCloudWatchLogsService struct {
	mock.Mock
}

// GetCloudWatchLogsWaste mocks the GetCloudWatchLogsWaste method.
func (m *MockCloudWatchLogsService) GetCloudWatchLogsWaste(ctx context.Context) ([]model.CloudWatchLogsWasteInfo, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]model.CloudWatchLogsWasteInfo), args.Error(1)
}

// GetLambdaMaxMemoryUsedBatch mocks the GetLambdaMaxMemoryUsedBatch method.
func (m *MockCloudWatchLogsService) GetLambdaMaxMemoryUsedBatch(ctx context.Context, logGroupNames []string, startTime, endTime time.Time) (map[string]int32, error) {
	args := m.Called(ctx, logGroupNames, startTime, endTime)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[string]int32), args.Error(1)
}

// ListExistingLogGroups mocks the ListExistingLogGroups method.
func (m *MockCloudWatchLogsService) ListExistingLogGroups(ctx context.Context, prefix string) (map[string]struct{}, error) {
	args := m.Called(ctx, prefix)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[string]struct{}), args.Error(1)
}
