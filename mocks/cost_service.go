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

func (m *MockCostService) GetCurrentMonthCostsByService(ctx context.Context) (*model.CostInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CostInfo), args.Error(1)
}

func (m *MockCostService) GetLastMonthCostsByService(ctx context.Context) (*model.CostInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CostInfo), args.Error(1)
}

func (m *MockCostService) GetMonthCostsByService(ctx context.Context, endDate time.Time) (*model.CostInfo, error) {
	args := m.Called(ctx, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CostInfo), args.Error(1)
}

func (m *MockCostService) GetCurrentMonthTotalCosts(ctx context.Context) (*string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockCostService) GetLastMonthTotalCosts(ctx context.Context) (*string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockCostService) GetLastSixMonthsCosts(ctx context.Context) ([]model.CostInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.CostInfo), args.Error(1)
}
