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
func (m *MockOutputService) RenderTrend(accountID string, costInfo []model.CostInfo, services []string) error {
	args := m.Called(accountID, costInfo, services)
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

// PrintReportSuccess mocks the PrintReportSuccess method.
func (m *MockOutputService) PrintReportSuccess(path string) {
	m.Called(path)
}

// PrintAlreadyLatest mocks the PrintAlreadyLatest method.
func (m *MockOutputService) PrintAlreadyLatest(version string) {
	m.Called(version)
}

// PrintRateLimitError mocks the PrintRateLimitError method.
func (m *MockOutputService) PrintRateLimitError() {
	m.Called()
}

// PrintUpdateError mocks the PrintUpdateError method.
func (m *MockOutputService) PrintUpdateError(err error) {
	m.Called(err)
}

// RenderVersion mocks the RenderVersion method.
func (m *MockOutputService) RenderVersion(versionInfo model.VersionInfo) {
	m.Called(versionInfo)
}

// PrintFirstDayOfMonthError mocks the PrintFirstDayOfMonthError method.
func (m *MockOutputService) PrintFirstDayOfMonthError() {
	m.Called()
}
