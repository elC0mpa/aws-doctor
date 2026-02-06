package renderers

import (
	"github.com/elC0mpa/aws-doctor/model"
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

// DrawTrendChart mocks DrawTrendChart
func (m *MockRenderer) DrawTrendChart(accountID string, costInfo []model.CostInfo) {
	m.Called(accountID, costInfo)
}

// OutputTrendJSON mocks OutputTrendJSON
func (m *MockRenderer) OutputTrendJSON(accountID string, costInfo []model.CostInfo) error {
	args := m.Called(accountID, costInfo)
	return args.Error(0)
}

// DrawWasteTable mocks DrawWasteTable
func (m *MockRenderer) DrawWasteTable(input model.RenderWasteInput) {
	m.Called(input)
}

// OutputWasteJSON mocks OutputWasteJSON
func (m *MockRenderer) OutputWasteJSON(input model.RenderWasteInput) error {
	args := m.Called(input)
	return args.Error(0)
}

// StopSpinner mocks StopSpinner
func (m *MockRenderer) StopSpinner() {
	m.Called()
}
