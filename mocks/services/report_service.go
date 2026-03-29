package services

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockReportService is a mock implementation of the ReportService interface
type MockReportService struct {
	mock.Mock
}

func (m *MockReportService) GenerateCostComparisonReport(input model.RenderCostComparisonInput, reportPath string) (*string, error) {
	args := m.Called(input, reportPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockReportService) GenerateTrendReport(accountID string, costInfo []model.CostInfo, services []string, reportPath string) (*string, error) {
	args := m.Called(accountID, costInfo, services, reportPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockReportService) GenerateWasteReport(input model.RenderWasteInput, reportPath string) (*string, error) {
	args := m.Called(input, reportPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}
