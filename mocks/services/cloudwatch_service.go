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

// GetMaxMemoryUsed mocks the GetMaxMemoryUsed method.
func (m *MockCloudWatchLogsService) GetMaxMemoryUsed(ctx context.Context, logGroupName string, startTime, endTime time.Time) (int32, error) {
	args := m.Called(ctx, logGroupName, startTime, endTime)

	return args.Get(0).(int32), args.Error(1)
}
