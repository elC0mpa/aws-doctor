package services

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	"github.com/stretchr/testify/mock"
)

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
func (m *MockOutputService) RenderWaste(input model.RenderWasteInput, pricingSvc pricing.Service) error {
	args := m.Called(input, pricingSvc)
	return args.Error(0)
}

// StopSpinner mocks the StopSpinner method.
func (m *MockOutputService) StopSpinner() {
	m.Called()
}

// SetSpinnerMessage mocks the SetSpinnerMessage method.
func (m *MockOutputService) SetSpinnerMessage(message string) {
	m.Called(message)
}

// PrintReportSuccess mocks the PrintReportSuccess method.
func (m *MockOutputService) PrintReportSuccess(path string) {
	m.Called(path)
}

// PrintAlreadyLatest mocks the PrintAlreadyLatest method.
func (m *MockOutputService) PrintAlreadyLatest(version string) {
	m.Called(version)
}

// PrintHomebrewUpdate mocks the PrintHomebrewUpdate method.
func (m *MockOutputService) PrintHomebrewUpdate() {
	m.Called()
}

// PrintGoInstallUpdate mocks the PrintGoInstallUpdate method.
func (m *MockOutputService) PrintGoInstallUpdate() {
	m.Called()
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

// PrintNewVersionAvailable mocks the PrintNewVersionAvailable method.
func (m *MockOutputService) PrintNewVersionAvailable(currentVersion, latestVersion string) {
	m.Called(currentVersion, latestVersion)
}
