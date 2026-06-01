package orchestrator

import (
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

func TestCompareCosts_Success(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockOutput := new(services.MockOutputService)
	mockReport := new(services.MockReportService)

	cfg := CostConfig{
		STSService:    mockSTS,
		CostService:   mockCost,
		OutputService: mockOutput,
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
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("RenderCostComparison", mock.Anything).Return(nil)

	err := svc.CompareCosts(false, "")

	assert.NoError(t, err)
	mockCost.AssertExpectations(t)
	mockSTS.AssertExpectations(t)
	mockOutput.AssertExpectations(t)
}

func TestUpdate_Success(t *testing.T) {
	mockUpdate := new(services.MockUpdateService)
	mockOutput := new(services.MockOutputService)

	cfg := SystemConfig{
		UpdateService: mockUpdate,
		OutputService: mockOutput,
		VersionInfo:   model.VersionInfo{Version: "dev"},
	}
	svc := NewSystemService(cfg)

	mockOutput.On("StopSpinner").Return()
	mockUpdate.On("Update").Return(nil)

	err := svc.Update()

	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
	mockUpdate.AssertExpectations(t)
}

func TestVersion_Success(t *testing.T) {
	mockOutput := new(services.MockOutputService)

	cfg := SystemConfig{
		OutputService: mockOutput,
		VersionInfo:   model.VersionInfo{Version: "dev"},
	}
	svc := NewSystemService(cfg)

	mockOutput.On("StopSpinner").Return()
	mockOutput.On("RenderVersion", mock.Anything).Return()

	err := svc.Version()

	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}

func TestAnalyzeTrends_Success(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockOutput := new(services.MockOutputService)
	mockReport := new(services.MockReportService)

	cfg := TrendConfig{
		STSService:    mockSTS,
		CostService:   mockCost,
		OutputService: mockOutput,
		ReportService: mockReport,
	}
	svc := NewTrendService(cfg)

	mockCost.On("GetLastSixMonthsCosts", mock.Anything, mock.Anything).Return([]model.CostInfo{}, nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("RenderTrend", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := svc.AnalyzeTrends([]string{"ec2"}, false, "")

	assert.NoError(t, err)
	mockCost.AssertExpectations(t)
	mockSTS.AssertExpectations(t)
	mockOutput.AssertExpectations(t)
}

func TestAnalyzeWaste_Success(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockPricing := new(services.MockPricingService)
	mockOutput := new(services.MockOutputService)
	mockReport := new(services.MockReportService)
	mockRegistry := analyzer.NewRegistry()

	cfg := WasteConfig{
		STSService:     mockSTS,
		PricingService: mockPricing,
		OutputService:  mockOutput,
		ReportService:  mockReport,
		Registry:       mockRegistry,
	}
	svc := NewWasteService(cfg)

	mockOutput.On("SetSpinnerMessage", mock.Anything).Return().Maybe()
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
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("RenderWaste", mock.Anything, mock.Anything).Return(nil)

	flags := model.Flags{WasteChecks: []string{"mock"}}
	err := svc.AnalyzeWaste(flags)

	assert.NoError(t, err)
	mockPricing.AssertExpectations(t)
	mockSTS.AssertExpectations(t)
	mockOutput.AssertExpectations(t)
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
