package renderers

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	"github.com/stretchr/testify/mock"
)

// MockRenderer is a mock implementation of output.Renderer
type MockRenderer struct {
	mock.Mock
}

// DrawCostTable mocks DrawCostTable
func (m *MockRenderer) DrawCostTable(input model.RenderCostComparisonInput) {
	m.Called(input)
}

// OutputCostComparisonJSON mocks OutputCostComparisonJSON
func (m *MockRenderer) OutputCostComparisonJSON(input model.RenderCostComparisonInput) error {
	args := m.Called(input)
	return args.Error(0)
}

// OutputCostComparisonCSV mocks OutputCostComparisonCSV
func (m *MockRenderer) OutputCostComparisonCSV(input model.RenderCostComparisonInput) error {
	args := m.Called(input)
	return args.Error(0)
}

// DrawTrendChart mocks DrawTrendChart
func (m *MockRenderer) DrawTrendChart(accountID string, costInfo []model.CostInfo) {
	m.Called(accountID, costInfo)
}

// OutputTrendJSON mocks OutputTrendJSON
func (m *MockRenderer) OutputTrendJSON(accountID string, costInfo []model.CostInfo, services []string) error {
	args := m.Called(accountID, costInfo, services)
	return args.Error(0)
}

// OutputTrendCSV mocks OutputTrendCSV
func (m *MockRenderer) OutputTrendCSV(costInfo []model.CostInfo, services []string) error {
	args := m.Called(costInfo, services)
	return args.Error(0)
}

// DrawWasteTable mocks DrawWasteTable
func (m *MockRenderer) DrawWasteTable(input model.RenderWasteInput, pricingSvc pricing.Service) {
	m.Called(input, pricingSvc)
}

// OutputWasteJSON mocks OutputWasteJSON
func (m *MockRenderer) OutputWasteJSON(input model.RenderWasteInput, pricingSvc pricing.Service) error {
	args := m.Called(input, pricingSvc)
	return args.Error(0)
}

// OutputWasteCSV mocks OutputWasteCSV
func (m *MockRenderer) OutputWasteCSV(input model.RenderWasteInput, pricingSvc pricing.Service) error {
	args := m.Called(input, pricingSvc)
	return args.Error(0)
}

// RenderWasteInteractive mocks the RenderWasteInteractive method
func (m *MockRenderer) RenderWasteInteractive(accountID string, resultCh <-chan model.ScopeResult, scopes []string, pricingSvc pricing.Service) error {
	args := m.Called(accountID, resultCh, scopes, pricingSvc)
	return args.Error(0)
}
