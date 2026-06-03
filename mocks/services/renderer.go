package services

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	"github.com/stretchr/testify/mock"
)

// MockRenderer is a mock implementation of the output service.
type MockRenderer struct {
	mock.Mock
}

// RenderCostComparison mocks the RenderCostComparison method.
func (m *MockRenderer) RenderCostComparison(input model.RenderCostComparisonInput) error {
	args := m.Called(input)
	return args.Error(0)
}

// RenderTrend mocks the RenderTrend method.
func (m *MockRenderer) RenderTrend(accountID string, costInfo []model.CostInfo, services []string) error {
	args := m.Called(accountID, costInfo, services)
	return args.Error(0)
}

// RenderWaste mocks the RenderWaste method.
func (m *MockRenderer) RenderWaste(input model.RenderWasteInput, pricingSvc pricing.Service) error {
	args := m.Called(input, pricingSvc)
	return args.Error(0)
}

// IsInteractive mocks the IsInteractive method.
func (m *MockRenderer) IsInteractive() bool {
	args := m.Called()
	return args.Bool(0)
}

// RenderWasteInteractive mocks the RenderWasteInteractive method.
func (m *MockRenderer) RenderWasteInteractive(accountID string, resultCh <-chan model.ScopeResult, scopes []string, pricingSvc pricing.Service) error {
	args := m.Called(accountID, resultCh, scopes, pricingSvc)
	return args.Error(0)
}
