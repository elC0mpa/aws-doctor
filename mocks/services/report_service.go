package services

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockReportService is a mock implementation of the ReportService interface
type MockReportService struct {
	mock.Mock
}

func (m *MockReportService) GenerateCostComparisonReport(input model.RenderCostComparisonInput, reportPath string) error {
	args := m.Called(input, reportPath)
	return args.Error(0)
}

func (m *MockReportService) GenerateTrendReport(accountID string, costInfo []model.CostInfo, services []string, reportPath string) error {
	args := m.Called(accountID, costInfo, services, reportPath)
	return args.Error(0)
}

func (m *MockReportService) GenerateWasteReport(input model.RenderWasteInput, reportPath string) error {
	args := m.Called(input, reportPath)
	return args.Error(0)
}
