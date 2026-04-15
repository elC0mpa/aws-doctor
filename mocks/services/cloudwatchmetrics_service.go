package services

import (
	"context"

	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
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

// NatGatewayBytesOut mocks the NatGatewayBytesOut method.
func (m *MockCloudWatchMetricsService) NatGatewayBytesOut(ctx context.Context, natGatewayID string, days int) (float64, error) {
	args := m.Called(ctx, natGatewayID, days)
	return args.Get(0).(float64), args.Error(1)
}

// ELBHasZeroRequestsInPeriod mocks the ELBHasZeroRequestsInPeriod method.
func (m *MockCloudWatchMetricsService) ELBHasZeroRequestsInPeriod(ctx context.Context, loadBalancerArn string, lbType elbtypes.LoadBalancerTypeEnum, days int) (bool, error) {
	args := m.Called(ctx, loadBalancerArn, lbType, days)

	return args.Bool(0), args.Error(1)
}
