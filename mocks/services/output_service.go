package services

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockOutputService is a mock implementation of the output service interface.
type MockOutputService struct {
	mock.Mock
}

// RenderCostComparison mocks the RenderCostComparison method.
func (m *MockOutputService) RenderCostComparison(input model.RenderCostComparisonInput) error {
	args := m.Called(input)
	return args.Error(0)
}

// RenderTrend mocks the RenderTrend method.
func (m *MockOutputService) RenderTrend(accountID string, costInfo []model.CostInfo) error {
	args := m.Called(accountID, costInfo)
	return args.Error(0)
}

// RenderWaste mocks the RenderWaste method.
func (m *MockOutputService) RenderWaste(input model.RenderWasteInput) error {
	args := m.Called(input)
	return args.Error(0)
}

// StopSpinner mocks the StopSpinner method.
func (m *MockOutputService) StopSpinner() {
	m.Called()
}
