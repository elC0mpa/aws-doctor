package orchestrator

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/mock"
	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/elC0mpa/aws-doctor/model"
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
	mockVPCSvc := &services.MockVPCService{}
	
	s := &service{
		stsService:     mockSTSSvc,
		outputService:  mockOutputSvc,
		pricingService: mockPricingSvc,
		vpcService:     mockVPCSvc,
	}

	acc := "123456789012"
	mockSTSSvc.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{Account: &acc}, nil)
	mockPricingSvc.On("LoadPricing", mock.Anything).Return(nil)
	mockPricingSvc.On("LoadRegionRates", mock.Anything).Return(nil)
	mockOutputSvc.On("SetSpinnerMessage", mock.Anything).Return()
	mockOutputSvc.On("StopSpinner").Return()
	mockOutputSvc.On("IsInteractive").Return(false)
	mockOutputSvc.On("RenderWaste", mock.Anything, mockPricingSvc).Return(nil)

	mockVPCSvc.On("GetIdleNATGateways", mock.Anything, mock.Anything).Return(nil, nil)

	err := s.wasteWorkflow([]string{"vpc"}, false, "", 0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestTrendWorkflow(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockOutputSvc := &services.MockOutputService{}
	mockCostSvc := &services.MockCostService{}
	
	s := &service{
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

func TestWasteWorkflow_Errors_Extra(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockPricingSvc := &services.MockPricingService{}
	mockOutputSvc := &services.MockOutputService{}
	
	s := &service{
		stsService:     mockSTSSvc,
		pricingService: mockPricingSvc,
		outputService:  mockOutputSvc,
	}

	mockOutputSvc.On("SetSpinnerMessage", mock.Anything).Return()
	
	// LoadPricing error
	mockPricingSvc.On("LoadPricing", mock.Anything).Return(errors.New("pricing error")).Once()
	err := s.wasteWorkflow([]string{"ec2"}, false, "", 0, 0)
	if err == nil {
		t.Error("Expected error for LoadPricing failure")
	}
	
	// GetCallerIdentity error
	mockPricingSvc.On("LoadPricing", mock.Anything).Return(nil)
	mockPricingSvc.On("LoadRegionRates", mock.Anything).Return(nil)
	mockSTSSvc.On("GetCallerIdentity", mock.Anything).Return((*sts.GetCallerIdentityOutput)(nil), errors.New("sts error")).Once()
	err = s.wasteWorkflow([]string{"ec2"}, false, "", 0, 0)
	if err == nil {
		t.Error("Expected error for STS failure")
	}
}

func TestWasteWorkflow_STS_Error_Safe(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockPricingSvc := &services.MockPricingService{}
	mockOutputSvc := &services.MockOutputService{}
	
	s := &service{
		stsService:     mockSTSSvc,
		pricingService: mockPricingSvc,
		outputService:  mockOutputSvc,
	}

	mockOutputSvc.On("SetSpinnerMessage", mock.Anything).Return()
	mockPricingSvc.On("LoadPricing", mock.Anything).Return(nil)
	mockPricingSvc.On("LoadRegionRates", mock.Anything).Return(nil)
	
	mockSTSSvc.On("GetCallerIdentity", mock.Anything).Return((*sts.GetCallerIdentityOutput)(nil), errors.New("sts error")).Once()
	err := s.wasteWorkflow([]string{"ec2"}, false, "", 0, 0)
	if err == nil {
		t.Error("Expected error for STS failure")
	}
}

func TestWasteWorkflow_LoadPricing_Error_Safe(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockPricingSvc := &services.MockPricingService{}
	mockOutputSvc := &services.MockOutputService{}
	
	s := &service{
		stsService:     mockSTSSvc,
		pricingService: mockPricingSvc,
		outputService:  mockOutputSvc,
	}

	mockOutputSvc.On("SetSpinnerMessage", mock.Anything).Return()
	mockPricingSvc.On("LoadPricing", mock.Anything).Return(errors.New("load error")).Once()
	
	err := s.wasteWorkflow([]string{"ec2"}, false, "", 0, 0)
	if err == nil {
		t.Error("Expected error for LoadPricing failure")
	}
}

func TestWasteWorkflow_LoadRegionRates_Error_Safe(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockPricingSvc := &services.MockPricingService{}
	mockOutputSvc := &services.MockOutputService{}
	
	s := &service{
		stsService:     mockSTSSvc,
		pricingService: mockPricingSvc,
		outputService:  mockOutputSvc,
	}

	mockOutputSvc.On("SetSpinnerMessage", mock.Anything).Return()
	mockPricingSvc.On("LoadPricing", mock.Anything).Return(nil)
	mockPricingSvc.On("LoadRegionRates", mock.Anything).Return(errors.New("region error")).Once()
	
	err := s.wasteWorkflow([]string{"ec2"}, false, "", 0, 0)
	if err == nil {
		t.Error("Expected error for LoadRegionRates failure")
	}
}

func TestWasteWorkflow_RenderWaste_Error_Safe(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockPricingSvc := &services.MockPricingService{}
	mockOutputSvc := &services.MockOutputService{}
	mockVPCSvc := &services.MockVPCService{}
	
	s := &service{
		stsService:     mockSTSSvc,
		pricingService: mockPricingSvc,
		outputService:  mockOutputSvc,
		vpcService:     mockVPCSvc,
	}

	acc := "123456789012"
	mockSTSSvc.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{Account: &acc}, nil)
	mockPricingSvc.On("LoadPricing", mock.Anything).Return(nil)
	mockPricingSvc.On("LoadRegionRates", mock.Anything).Return(nil)
	mockOutputSvc.On("SetSpinnerMessage", mock.Anything).Return()
	mockOutputSvc.On("StopSpinner").Return()
	mockOutputSvc.On("IsInteractive").Return(false)
	
	mockVPCSvc.On("GetIdleNATGateways", mock.Anything, mock.Anything).Return(nil, nil)
	
	mockOutputSvc.On("RenderWaste", mock.Anything, mockPricingSvc).Return(errors.New("render error")).Once()

	err := s.wasteWorkflow([]string{"vpc"}, false, "", 0, 0)
	if err == nil {
		t.Error("Expected error for RenderWaste failure")
	}
}

func TestWasteWorkflow_Interactive_EOF_Safe(t *testing.T) {
	mockSTSSvc := &services.MockSTSService{}
	mockPricingSvc := &services.MockPricingService{}
	mockOutputSvc := &services.MockOutputService{}
	mockVPCSvc := &services.MockVPCService{}
	
	s := &service{
		stsService:     mockSTSSvc,
		pricingService: mockPricingSvc,
		outputService:  mockOutputSvc,
		vpcService:     mockVPCSvc,
	}

	acc := "123456789012"
	mockSTSSvc.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{Account: &acc}, nil)
	mockPricingSvc.On("LoadPricing", mock.Anything).Return(nil)
	mockPricingSvc.On("LoadRegionRates", mock.Anything).Return(nil)
	mockOutputSvc.On("SetSpinnerMessage", mock.Anything).Return()
	mockOutputSvc.On("StopSpinner").Return()
	mockOutputSvc.On("IsInteractive").Return(true)
	
	mockVPCSvc.On("GetIdleNATGateways", mock.Anything, mock.Anything).Return(nil, nil)
	
	mockOutputSvc.On("RenderWasteInteractive", mock.Anything, acc, mock.Anything, mockPricingSvc).Return(nil).Once()

	err := s.wasteWorkflow([]string{"vpc"}, false, "", 0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
