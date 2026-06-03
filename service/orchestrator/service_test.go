package orchestrator

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testReportPath = "test.pdf"

func TestCompareCosts_Success(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockOutput := new(services.MockRenderer)
	mockReport := new(services.MockReportService)

	cfg := CostConfig{
		STSService:    mockSTS,
		CostService:   mockCost,
		Renderer:      mockOutput,
		ReportService: mockReport,
	}
	svc := NewCostService(cfg)

	mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	mockCost.On("GetLastMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	mockCost.On("GetCurrentMonthTotalCosts", mock.Anything).Return(aws.String("100.00"), nil)
	mockCost.On("GetLastMonthTotalCosts", mock.Anything).Return(aws.String("90.00"), nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("RenderCostComparison", mock.Anything).Return(nil)

	err := svc.CompareCosts(false, "")

	assert.NoError(t, err)
	mockCost.AssertExpectations(t)
	mockSTS.AssertExpectations(t)
}

func TestUpdate_Success(t *testing.T) {
	mockUpdate := new(services.MockUpdateService)
	mockOutput := new(services.MockRenderer)

	cfg := SystemConfig{
		UpdateService: mockUpdate,
		Renderer:      mockOutput,
		VersionInfo:   model.VersionInfo{Version: "dev"},
	}
	svc := NewSystemService(cfg)

	mockUpdate.On("Update").Return(nil)

	err := svc.Update()

	assert.NoError(t, err)
	mockUpdate.AssertExpectations(t)
}

func TestVersion_Success(t *testing.T) {
	mockOutput := new(services.MockRenderer)

	cfg := SystemConfig{
		Renderer:    mockOutput,
		VersionInfo: model.VersionInfo{Version: "dev"},
	}
	svc := NewSystemService(cfg)

	err := svc.Version()

	assert.NoError(t, err)
}

func TestAnalyzeTrends_Success(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockOutput := new(services.MockRenderer)
	mockReport := new(services.MockReportService)

	cfg := TrendConfig{
		STSService:    mockSTS,
		CostService:   mockCost,
		Renderer:      mockOutput,
		ReportService: mockReport,
	}
	svc := NewTrendService(cfg)

	mockCost.On("GetLastSixMonthsCosts", mock.Anything, mock.Anything).Return([]model.CostInfo{}, nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("RenderTrend", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := svc.AnalyzeTrends([]string{"ec2"}, false, "")

	assert.NoError(t, err)
	mockCost.AssertExpectations(t)
	mockSTS.AssertExpectations(t)
}

func TestAnalyzeWaste_Success(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockPricing := new(services.MockPricingService)
	mockOutput := new(services.MockRenderer)
	mockReport := new(services.MockReportService)
	mockRegistry := analyzer.NewRegistry()

	cfg := WasteConfig{
		STSService:     mockSTS,
		PricingService: mockPricing,
		Renderer:       mockOutput,
		ReportService:  mockReport,
		Registry:       mockRegistry,
	}
	svc := NewWasteService(cfg)

	mockPricing.On("LoadRegionRates", mock.Anything).Return(nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)

	mockAnalyzer := new(services.MockAnalyzer)
	mockAnalyzer.On("Name").Return("mock")
	mockAnalyzer.On("TabName").Return("MockTab")
	mockAnalyzer.On("Analyze", mock.Anything, mock.Anything).Return(model.ScopeResult{Input: model.RenderWasteInput{}}, nil)

	mockRegistry.Register(mockAnalyzer)

	mockOutput.On("IsInteractive").Return(false)
	mockOutput.On("RenderWaste", mock.Anything, mock.Anything).Return(nil)

	flags := model.Flags{WasteChecks: []string{"mock"}}
	err := svc.AnalyzeWaste(flags)

	assert.NoError(t, err)
	mockPricing.AssertExpectations(t)
	mockSTS.AssertExpectations(t)
	mockAnalyzer.AssertExpectations(t)
}

func TestCheckForUpdateInBackground(t *testing.T) {
	mockUpdate := new(services.MockUpdateService)

	cfg := SystemConfig{
		UpdateService: mockUpdate,
	}
	svc := NewSystemService(cfg)

	version := "v2.0.0"
	mockUpdate.On("CheckForUpdate", mock.Anything).Return(&version, nil)

	ch := svc.CheckForUpdateInBackground()

	select {
	case res := <-ch:
		assert.NoError(t, res.Err)
		assert.Equal(t, version, *res.LatestVersion)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for background update")
	}
}

func TestUpdate_Error(t *testing.T) {
	mockUpdate := new(services.MockUpdateService)
	mockOutput := new(services.MockRenderer)

	cfg := SystemConfig{
		UpdateService: mockUpdate,
		Renderer:      mockOutput,
		VersionInfo:   model.VersionInfo{Version: "dev"},
	}
	svc := NewSystemService(cfg)

	mockUpdate.On("Update").Return(errors.New("update failed"))

	err := svc.Update()

	assert.Error(t, err)
	assert.Equal(t, "update failed", err.Error())
	mockUpdate.AssertExpectations(t)
}

func TestCompareCosts_Error(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockOutput := new(services.MockRenderer)

	cfg := CostConfig{
		STSService:  mockSTS,
		CostService: mockCost,
		Renderer:    mockOutput,
	}
	svc := NewCostService(cfg)

	mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(nil, errors.New("cost error"))
	mockOutput.On("PrintCostError", mock.Anything).Return()

	err := svc.CompareCosts(false, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cost error")
}

func TestCompareCosts_ReportSuccess(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockOutput := new(services.MockRenderer)
	mockReport := new(services.MockReportService)

	cfg := CostConfig{
		STSService:    mockSTS,
		CostService:   mockCost,
		Renderer:      mockOutput,
		ReportService: mockReport,
	}
	svc := NewCostService(cfg)

	mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	mockCost.On("GetLastMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	mockCost.On("GetCurrentMonthTotalCosts", mock.Anything).Return(aws.String("100.00"), nil)
	mockCost.On("GetLastMonthTotalCosts", mock.Anything).Return(aws.String("90.00"), nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)

	mockOutput.On("IsInteractive").Return(false)

	path := testReportPath
	mockReport.On("GenerateCostComparisonReport", mock.Anything, testReportPath).Return(&path, nil)

	mockOutput.On("PrintReportError", mock.Anything).Return()

	err := svc.CompareCosts(true, testReportPath)

	assert.NoError(t, err)
}

func TestCompareCosts_ReportError(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockOutput := new(services.MockRenderer)
	mockReport := new(services.MockReportService)

	cfg := CostConfig{
		STSService:    mockSTS,
		CostService:   mockCost,
		Renderer:      mockOutput,
		ReportService: mockReport,
	}
	svc := NewCostService(cfg)

	mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	mockCost.On("GetLastMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	mockCost.On("GetCurrentMonthTotalCosts", mock.Anything).Return(aws.String("100.00"), nil)
	mockCost.On("GetLastMonthTotalCosts", mock.Anything).Return(aws.String("90.00"), nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)

	mockReport.On("GenerateCostComparisonReport", mock.Anything, testReportPath).Return((*string)(nil), errors.New("report failed"))

	mockOutput.On("PrintReportError", mock.Anything).Return()

	err := svc.CompareCosts(true, testReportPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "report failed")
}

func TestAnalyzeTrends_Error(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockOutput := new(services.MockRenderer)

	cfg := TrendConfig{
		STSService:  mockSTS,
		CostService: mockCost,
		Renderer:    mockOutput,
	}
	svc := NewTrendService(cfg)

	mockCost.On("GetLastSixMonthsCosts", mock.Anything, mock.Anything).Return([]model.CostInfo{}, errors.New("trend error"))
	mockOutput.On("PrintTrendError", mock.Anything).Return()

	err := svc.AnalyzeTrends(nil, false, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trend error")
}

func TestAnalyzeTrends_ReportSuccess(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockOutput := new(services.MockRenderer)
	mockReport := new(services.MockReportService)

	cfg := TrendConfig{
		STSService:    mockSTS,
		CostService:   mockCost,
		Renderer:      mockOutput,
		ReportService: mockReport,
	}
	svc := NewTrendService(cfg)

	mockCost.On("GetLastSixMonthsCosts", mock.Anything, mock.Anything).Return([]model.CostInfo{}, nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)

	mockOutput.On("IsInteractive").Return(false)

	path := testReportPath
	mockReport.On("GenerateTrendReport", mock.Anything, mock.Anything, mock.Anything, testReportPath).Return(&path, nil)

	mockOutput.On("PrintReportError", mock.Anything).Return()

	err := svc.AnalyzeTrends(nil, true, testReportPath)

	assert.NoError(t, err)
}

func TestAnalyzeWaste_Error(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockOutput := new(services.MockRenderer)
	mockPricing := new(services.MockPricingService)

	cfg := WasteConfig{
		STSService:     mockSTS,
		PricingService: mockPricing,
		Renderer:       mockOutput,
		Registry:       analyzer.NewRegistry(),
	}
	svc := NewWasteService(cfg)

	mockSTS.On("GetCallerIdentity", mock.Anything).Return(nil, errors.New("sts error"))
	mockPricing.On("LoadRegionRates", mock.Anything).Return(nil).Maybe()

	err := svc.AnalyzeWaste(model.Flags{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sts error")
}

func TestAnalyzeWaste_ReportSuccess(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockPricing := new(services.MockPricingService)
	mockOutput := new(services.MockRenderer)
	mockReport := new(services.MockReportService)
	mockRegistry := analyzer.NewRegistry()

	cfg := WasteConfig{
		STSService:     mockSTS,
		PricingService: mockPricing,
		Renderer:       mockOutput,
		ReportService:  mockReport,
		Registry:       mockRegistry,
	}
	svc := NewWasteService(cfg)

	mockPricing.On("LoadRegionRates", mock.Anything).Return(nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)

	mockOutput.On("IsInteractive").Return(false)

	path := testReportPath
	mockReport.On("GenerateWasteReport", mock.Anything, mockPricing, testReportPath).Return(&path, nil)

	flags := model.Flags{Report: true, ReportPath: testReportPath}
	err := svc.AnalyzeWaste(flags)

	assert.NoError(t, err)
}
