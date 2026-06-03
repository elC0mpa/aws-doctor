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

// StopSpinner mocks the StopSpinner method.
func (m *MockRenderer) StopSpinner() {
	m.Called()
}

// SetSpinnerMessage mocks the SetSpinnerMessage method.
func (m *MockRenderer) SetSpinnerMessage(message string) {
	m.Called(message)
}

// PrintReportSuccess mocks the PrintReportSuccess method.
func (m *MockRenderer) PrintReportSuccess(path string) {
	m.Called(path)
}

// PrintAlreadyLatest mocks the PrintAlreadyLatest method.
func (m *MockRenderer) PrintAlreadyLatest(version string) {
	m.Called(version)
}

// PrintHomebrewUpdate mocks the PrintHomebrewUpdate method.
func (m *MockRenderer) PrintHomebrewUpdate() {
	m.Called()
}

// PrintGoInstallUpdate mocks the PrintGoInstallUpdate method.
func (m *MockRenderer) PrintGoInstallUpdate() {
	m.Called()
}

// PrintRateLimitError mocks the PrintRateLimitError method.
func (m *MockRenderer) PrintRateLimitError() {
	m.Called()
}

// PrintUpdateError mocks PrintUpdateError
func (m *MockRenderer) PrintUpdateError(err error) {
	m.Called(err)
}

// PrintWasteError mocks PrintWasteError
func (m *MockRenderer) PrintWasteError(err error) {
	m.Called(err)
}

// RenderVersion mocks the RenderVersion method.
func (m *MockRenderer) RenderVersion(versionInfo model.VersionInfo) {
	m.Called(versionInfo)
}

// PrintFirstDayOfMonthError mocks the PrintFirstDayOfMonthError method.
func (m *MockRenderer) PrintFirstDayOfMonthError() {
	m.Called()
}

// PrintNewVersionAvailable mocks the PrintNewVersionAvailable method.
func (m *MockRenderer) PrintNewVersionAvailable(currentVersion, latestVersion string) {
	m.Called(currentVersion, latestVersion)
}
