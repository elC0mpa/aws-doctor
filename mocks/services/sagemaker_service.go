package services

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockSageMakerService mocks the SageMaker service for testing.
type MockSageMakerService struct {
	mock.Mock
}

// GetIdleEndpoints mocks the GetIdleEndpoints method.
func (m *MockSageMakerService) GetIdleEndpoints(ctx context.Context, idleDays int) ([]model.IdleSageMakerEndpointInfo, error) {
	args := m.Called(ctx, idleDays)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]model.IdleSageMakerEndpointInfo), args.Error(1)
}
