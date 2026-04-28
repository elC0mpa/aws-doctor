package services

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockLambdaService is a mock implementation of the Lambda service interface.
type MockLambdaService struct {
	mock.Mock
}

// GetOverProvisionedFunctions mocks the GetOverProvisionedFunctions method.
func (m *MockLambdaService) GetOverProvisionedFunctions(ctx context.Context, memoryThresholdPercent int, lookbackDays int) ([]model.LambdaOverProvisionedInfo, error) {
	args := m.Called(ctx, memoryThresholdPercent, lookbackDays)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]model.LambdaOverProvisionedInfo), args.Error(1)
}
