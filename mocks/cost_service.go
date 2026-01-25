// Package mocks provides mock implementations for testing.
package mocks

import (
	"context"
	"time"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockCostService is a mock implementation of the Cost service interface.
type MockCostService struct {
	mock.Mock
}

// GetCurrentMonthCostsByService mocks the GetCurrentMonthCostsByService method.
func (m *MockCostService) GetCurrentMonthCostsByService(ctx context.Context) (*model.CostInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CostInfo), args.Error(1)
}

// GetLastMonthCostsByService mocks the GetLastMonthCostsByService method.
func (m *MockCostService) GetLastMonthCostsByService(ctx context.Context) (*model.CostInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CostInfo), args.Error(1)
}

// GetMonthCostsByService mocks the GetMonthCostsByService method.
func (m *MockCostService) GetMonthCostsByService(ctx context.Context, endDate time.Time) (*model.CostInfo, error) {
	args := m.Called(ctx, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CostInfo), args.Error(1)
}

// GetCurrentMonthTotalCosts mocks the GetCurrentMonthTotalCosts method.
func (m *MockCostService) GetCurrentMonthTotalCosts(ctx context.Context) (*string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

// GetLastMonthTotalCosts mocks the GetLastMonthTotalCosts method.
func (m *MockCostService) GetLastMonthTotalCosts(ctx context.Context) (*string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

// GetLastSixMonthsCosts mocks the GetLastSixMonthsCosts method.
func (m *MockCostService) GetLastSixMonthsCosts(ctx context.Context) ([]model.CostInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.CostInfo), args.Error(1)
}
