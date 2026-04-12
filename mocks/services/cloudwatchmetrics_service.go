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

// GetNatGatewayBytesOut mocks the GetNatGatewayBytesOut method.
func (m *MockCloudWatchMetricsService) GetNatGatewayBytesOut(ctx context.Context, natGatewayID string, days int) (float64, error) {
	args := m.Called(ctx, natGatewayID, days)
	return args.Get(0).(float64), args.Error(1)
}
