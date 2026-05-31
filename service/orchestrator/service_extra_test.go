package orchestrator

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/analyzer"
	"github.com/stretchr/testify/mock"
)

func TestHandleWasteReport(t *testing.T) {
	mockReportSvc := &services.MockReportService{}
	mockOutputSvc := &services.MockOutputService{}
	mockPricingSvc := &services.MockPricingService{}

	s := &service{
		reportService:  mockReportSvc,
		outputService:  mockOutputSvc,
		pricingService: mockPricingSvc,
	}

	path := "test.html"
	mockReportSvc.On("GenerateWasteReport", mock.Anything, mockPricingSvc, "").Return(&path, nil)
	mockOutputSvc.On("PrintReportSuccess", path).Return()

	err := s.handleWasteReport(model.RenderWasteInput{}, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	errPath := "error.html"
	mockReportSvc.On("GenerateWasteReport", mock.Anything, mockPricingSvc, errPath).Return((*string)(nil), errors.New("report error"))

	err = s.handleWasteReport(model.RenderWasteInput{}, errPath)
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func TestWasteWorkflow_NonInteractive(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockOutputSvc := &services.MockOutputService{}
	mockPricingSvc := &services.MockPricingService{}

	s := &service{
		registry:       analyzer.NewRegistry(),
		stsService:     mockSTSSvc,
		outputService:  mockOutputSvc,
		pricingService: mockPricingSvc,
	}

	acc := "123456789012"
	mockSTSSvc.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{Account: &acc}, nil)
	mockPricingSvc.On("LoadRegionRates", mock.Anything).Return(nil)
	mockOutputSvc.On("SetSpinnerMessage", mock.Anything).Return()
	mockOutputSvc.On("StopSpinner").Return()
	mockOutputSvc.On("IsInteractive").Return(false)
	mockOutputSvc.On("RenderWaste", mock.Anything, mockPricingSvc).Return(nil)

	err := s.wasteWorkflow([]string{"vpc"}, false, "", 0, 0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestWasteWorkflow_AnalyzerError(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockOutputSvc := &services.MockOutputService{}
	mockPricingSvc := &services.MockPricingService{}

	// Create a mock analyzer that returns an error
	mockAnalyzer := &services.MockAnalyzer{}
	mockAnalyzer.On("Name").Return("mock")
	mockAnalyzer.On("TabName").Return("MockTab")
	mockAnalyzer.On("Analyze", mock.Anything, mock.Anything).Return(model.ScopeResult{}, errors.New("mock analyzer failed"))

	reg := analyzer.NewRegistry()
	reg.Register(mockAnalyzer)

	s := &service{
		registry:       reg,
		stsService:     mockSTSSvc,
		outputService:  mockOutputSvc,
		pricingService: mockPricingSvc,
	}

	acc := "123456789012"
	mockSTSSvc.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{Account: &acc}, nil)
	mockPricingSvc.On("LoadRegionRates", mock.Anything).Return(nil)
	mockOutputSvc.On("SetSpinnerMessage", mock.Anything).Return()
	mockOutputSvc.On("StopSpinner").Return()
	mockOutputSvc.On("IsInteractive").Return(false)

	// Capture the RenderWasteInput to verify the Errors map
	mockOutputSvc.On("RenderWaste", mock.MatchedBy(func(input model.RenderWasteInput) bool {
		return input.Errors != nil && input.Errors["MockTab"] == "mock analyzer failed"
	}), mockPricingSvc).Return(nil)

	err := s.wasteWorkflow([]string{}, false, "", 0, 0, 0)
	if err != nil {
		t.Errorf("Unexpected workflow error: %v", err)
	}

	mockOutputSvc.AssertExpectations(t)
}

func TestTrendWorkflow(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockOutputSvc := &services.MockOutputService{}
	mockCostSvc := &services.MockCostService{}

	s := &service{
		registry:      analyzer.NewRegistry(),
		stsService:    mockSTSSvc,
		outputService: mockOutputSvc,
		costService:   mockCostSvc,
	}

	acc := "123456789012"
	mockSTSSvc.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{Account: &acc}, nil)
	mockCostSvc.On("GetLastSixMonthsCosts", mock.Anything, mock.Anything).Return([]model.CostInfo{}, nil)
	mockOutputSvc.On("StopSpinner").Return()
	mockOutputSvc.On("RenderTrend", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := s.trendWorkflow([]string{"ec2"}, false, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestWasteWorkflow_STSError(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockPricingSvc := &services.MockPricingService{}
	mockOutputSvc := &services.MockOutputService{}

	s := &service{
		registry:       analyzer.NewRegistry(),
		stsService:     mockSTSSvc,
		pricingService: mockPricingSvc,
		outputService:  mockOutputSvc,
	}

	mockPricingSvc.On("LoadRegionRates", mock.Anything).Return(nil)
	mockOutputSvc.On("SetSpinnerMessage", mock.Anything).Return()
	mockSTSSvc.On("GetCallerIdentity", mock.Anything).Return((*sts.GetCallerIdentityOutput)(nil), errors.New("sts error")).Once()

	err := s.wasteWorkflow([]string{"ec2"}, false, "", 0, 0, 0)
	if err == nil {
		t.Error("Expected error for STS failure")
	}
}

func TestWasteWorkflow_LoadRegionRatesError_Ignored(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockPricingSvc := &services.MockPricingService{}
	mockOutputSvc := &services.MockOutputService{}

	s := &service{
		registry:       analyzer.NewRegistry(),
		stsService:     mockSTSSvc,
		pricingService: mockPricingSvc,
		outputService:  mockOutputSvc,
	}

	acc := "123456789012"
	mockSTSSvc.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{Account: &acc}, nil)
	// Pricing fails, but workflow continues
	mockPricingSvc.On("LoadRegionRates", mock.Anything).Return(errors.New("pricing api failure")).Once()
	mockOutputSvc.On("SetSpinnerMessage", mock.Anything).Return()
	mockOutputSvc.On("StopSpinner").Return()
	mockOutputSvc.On("IsInteractive").Return(false)

	mockOutputSvc.On("RenderWaste", mock.Anything, mockPricingSvc).Return(nil)

	err := s.wasteWorkflow([]string{"vpc"}, false, "", 0, 0, 0)
	if err != nil {
		t.Errorf("Expected wasteWorkflow to succeed despite pricing error, got: %v", err)
	}
}

func TestWasteWorkflow_RenderWaste_Error(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockPricingSvc := &services.MockPricingService{}
	mockOutputSvc := &services.MockOutputService{}

	s := &service{
		registry:       analyzer.NewRegistry(),
		stsService:     mockSTSSvc,
		pricingService: mockPricingSvc,
		outputService:  mockOutputSvc,
	}

	acc := "123456789012"
	mockSTSSvc.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{Account: &acc}, nil)
	mockPricingSvc.On("LoadRegionRates", mock.Anything).Return(nil)
	mockOutputSvc.On("SetSpinnerMessage", mock.Anything).Return()
	mockOutputSvc.On("StopSpinner").Return()
	mockOutputSvc.On("IsInteractive").Return(false)

	mockOutputSvc.On("RenderWaste", mock.Anything, mockPricingSvc).Return(errors.New("render error")).Once()

	err := s.wasteWorkflow([]string{"vpc"}, false, "", 0, 0, 0)
	if err == nil {
		t.Error("Expected error for RenderWaste failure")
	}
}

func TestWasteWorkflow_Interactive_EOF(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockPricingSvc := &services.MockPricingService{}
	mockOutputSvc := &services.MockOutputService{}

	s := &service{
		registry:       analyzer.NewRegistry(),
		stsService:     mockSTSSvc,
		pricingService: mockPricingSvc,
		outputService:  mockOutputSvc,
	}

	acc := "123456789012"
	mockSTSSvc.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{Account: &acc}, nil)
	mockPricingSvc.On("LoadRegionRates", mock.Anything).Return(nil)
	mockOutputSvc.On("SetSpinnerMessage", mock.Anything).Return()
	mockOutputSvc.On("StopSpinner").Return()
	mockOutputSvc.On("IsInteractive").Return(true)

	mockOutputSvc.On("RenderWasteInteractive", acc, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	err := s.wasteWorkflow([]string{"vpc"}, false, "", 0, 0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
