package mocks

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockRoute53Service is a mock implementation of the Route 53 service interface.
type MockRoute53Service struct {
	mock.Mock
}

func (m *MockRoute53Service) GetEmptyHostedZones(ctx context.Context) ([]model.HostedZoneWasteInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.HostedZoneWasteInfo), args.Error(1)
}
