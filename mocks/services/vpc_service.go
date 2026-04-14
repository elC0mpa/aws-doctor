package services

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockVPCService mocks the VPC service for testing.
type MockVPCService struct {
	mock.Mock
}

// IdleNATGateways mocks the IdleNATGateways method.
func (m *MockVPCService) IdleNATGateways(ctx context.Context, idleDays int) ([]model.NATGatewayWasteInfo, error) {
	args := m.Called(ctx, idleDays)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]model.NATGatewayWasteInfo), args.Error(1)
}
