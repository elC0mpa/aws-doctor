package services

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockCloudWatchMetricsService is a mock of CloudWatch metrics Service
type MockCloudWatchMetricsService struct {
	mock.Mock
}

// RDSHasZeroConnectionsInPeriod mocks the RDSHasZeroConnectionsInPeriod method.
func (m *MockCloudWatchMetricsService) RDSHasZeroConnectionsInPeriod(ctx context.Context, dbInstanceID string, days int) (bool, error) {
	args := m.Called(ctx, dbInstanceID, days)

	return args.Bool(0), args.Error(1)
}
