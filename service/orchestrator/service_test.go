package orchestrator

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testLatestVersion = "v1.3.0"

func TestOrchestrate_RouteToDefaultWorkflow(t *testing.T) {
	// Setup mocks
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)

	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	mockPricing := new(services.MockPricingService)

	// Create service

	config := Config{
		Registry: analyzer.NewRegistry(), STSService: mockSTS,
		CostService: mockCost,

		PricingService: mockPricing,
		OutputService:  mockOutput,
		UpdateService:  mockUpdate,
		ReportService:  mockReport,

		VersionInfo: model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
	}
	svc := NewService(config)
	// Setup expectations for default workflow
	mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	mockCost.On("GetLastMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	mockCost.On("GetCurrentMonthTotalCosts", mock.Anything).Return(aws.String("100.00"), nil)
	mockCost.On("GetLastMonthTotalCosts", mock.Anything).Return(aws.String("90.00"), nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("SetSpinnerMessage", mock.Anything).Return().Maybe()
	mockOutput.On("RenderCostComparison", mock.Anything).Return(nil)
	mockUpdate.On("CheckForUpdate", mock.Anything).Return(nil, nil)

	// Execute
	flags := model.Flags{Output: "json"}
	err := svc.Orchestrate(flags)

	// Assert
	assert.NoError(t, err)
	mockCost.AssertExpectations(t)
	mockSTS.AssertExpectations(t)
	mockOutput.AssertExpectations(t)
}

func TestOrchestrate_RouteToUpdateWorkflow(t *testing.T) {
	// Setup mocks
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)

	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	// Create service

	config := Config{
		Registry: analyzer.NewRegistry(), STSService: mockSTS,
		CostService: mockCost,

		PricingService: new(services.MockPricingService),
		OutputService:  mockOutput,
		UpdateService:  mockUpdate,
		ReportService:  mockReport,

		VersionInfo: model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
	}
	svc := NewService(config)
	// Setup expectations
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("SetSpinnerMessage", mock.Anything).Return().Maybe()
	mockUpdate.On("Update").Return(nil)

	// Execute with Update flag
	flags := model.Flags{Update: true}
	err := svc.Orchestrate(flags)

	// Assert
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
	mockUpdate.AssertExpectations(t)
}

func TestOrchestrate_UpdateWorkflow_HomebrewInstall(t *testing.T) {
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)

	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	config := Config{
		Registry: analyzer.NewRegistry(), STSService: mockSTS,
		CostService: mockCost,

		PricingService: new(services.MockPricingService),
		OutputService:  mockOutput,
		UpdateService:  mockUpdate,
		ReportService:  mockReport,

		VersionInfo: model.VersionInfo{Version: "v1.0.0", Commit: "abc", Date: "2024-01-01"},
	}
	svc := NewService(config)

	mockOutput.On("StopSpinner").Return()
	mockOutput.On("SetSpinnerMessage", mock.Anything).Return().Maybe()
	mockUpdate.On("Update").Return(model.ErrHomebrewInstall)
	mockOutput.On("PrintHomebrewUpdate").Return()

	flags := model.Flags{Update: true}
	err := svc.Orchestrate(flags)

	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
	mockUpdate.AssertExpectations(t)
}

func TestOrchestrate_UpdateWorkflow_GoInstall(t *testing.T) {
	svc, m := newTestServiceWithMocks(model.VersionInfo{Version: "v1.0.0", Commit: "abc", Date: "2024-01-01"})

	m.output.On("StopSpinner").Return()
	m.update.On("Update").Return(model.ErrGoInstall)
	m.output.On("PrintGoInstallUpdate").Return()

	err := svc.Orchestrate(model.Flags{Update: true})

	assert.NoError(t, err)
	m.output.AssertExpectations(t)
	m.update.AssertExpectations(t)
}

func TestOrchestrate_RouteToVersionWorkflow(t *testing.T) {
	// Setup mocks
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)

	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	// Create service

	versionInfo := model.VersionInfo{Version: "v1.2.3", Commit: "abc", Date: "today"}
	config := Config{
		Registry: analyzer.NewRegistry(), STSService: mockSTS,
		CostService: mockCost,

		PricingService: new(services.MockPricingService),
		OutputService:  mockOutput,
		UpdateService:  mockUpdate,
		ReportService:  mockReport,

		VersionInfo: versionInfo,
	}
	svc := NewService(config)

	// Setup expectations
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("SetSpinnerMessage", mock.Anything).Return().Maybe()
	mockOutput.On("RenderVersion", versionInfo).Return()

	// Execute with Version flag
	flags := model.Flags{Version: true}
	err := svc.Orchestrate(flags)

	// Assert
	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
}

func TestOrchestrate_RouteToTrendWorkflow(t *testing.T) {
	// Setup mocks
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)

	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	// Create service

	config := Config{
		Registry: analyzer.NewRegistry(), STSService: mockSTS,
		CostService: mockCost,

		PricingService: new(services.MockPricingService),
		OutputService:  mockOutput,
		UpdateService:  mockUpdate,
		ReportService:  mockReport,

		VersionInfo: model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
	}
	svc := NewService(config)
	// Setup expectations for trend workflow
	mockCost.On("GetLastSixMonthsCosts", mock.Anything, mock.Anything).Return([]model.CostInfo{}, nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("SetSpinnerMessage", mock.Anything).Return().Maybe()
	mockOutput.On("RenderTrend", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockUpdate.On("CheckForUpdate", mock.Anything).Return(nil, nil)

	// Execute with Trend flag
	flags := model.Flags{Trend: true, Output: "json"}
	err := svc.Orchestrate(flags)

	// Assert
	assert.NoError(t, err)
	mockCost.AssertExpectations(t)
	mockSTS.AssertExpectations(t)
	mockOutput.AssertExpectations(t)
}

func TestOrchestrate_WasteTakesPrecedenceOverTrend(t *testing.T) {
	// Setup mocks
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)

	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	mockPricing := new(services.MockPricingService)

	// Create service

	config := Config{
		Registry: analyzer.NewRegistry(), STSService: mockSTS,
		CostService: mockCost,

		PricingService: mockPricing,
		OutputService:  mockOutput,
		UpdateService:  mockUpdate,
		ReportService:  mockReport,
		VersionInfo:    model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
	}
	svc := NewService(config)
	// Setup expectations for waste workflow (should be called, not trend)
	mockPricing.On("LoadRegionRates", mock.Anything, mock.Anything).Return(nil)

	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("SetSpinnerMessage", mock.Anything).Return().Maybe()
	mockOutput.On("IsInteractive").Return(false).Maybe()
	mockOutput.On("RenderWaste", mock.Anything, mockPricing).Return(nil)
	mockUpdate.On("CheckForUpdate", mock.Anything).Return(nil, nil)

	// Execute with both flags - Waste should take precedence
	flags := model.Flags{Waste: true, Trend: true, Output: "json"}
	err := svc.Orchestrate(flags)

	// Assert - cost service should NOT be called for trend
	assert.NoError(t, err)
	mockCost.AssertNotCalled(t, "GetLastSixMonthsCosts", mock.Anything, mock.Anything)
}

func TestOrchestrate_TrendWorkflow_Mapping(t *testing.T) {
	// Setup mocks
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)

	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	// Create service

	config := Config{
		Registry: analyzer.NewRegistry(), STSService: mockSTS,
		CostService: mockCost,

		PricingService: new(services.MockPricingService),
		OutputService:  mockOutput,
		UpdateService:  mockUpdate,
		ReportService:  mockReport,

		VersionInfo: model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
	}
	svc := NewService(config)
	// Shorthand services to check
	trendChecks := []string{"ec2", "s3", "rds", "invalid"}
	// Expected mapped services
	expectedMapped := []string{
		"Amazon Elastic Compute Cloud - Compute",
		"Amazon Simple Storage Service",
		"Amazon Relational Database Service",
	}

	// Setup expectations
	mockCost.On("GetLastSixMonthsCosts", mock.Anything, expectedMapped).Return([]model.CostInfo{}, nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("SetSpinnerMessage", mock.Anything).Return().Maybe()
	mockOutput.On("RenderTrend", "123456789012", mock.Anything, mock.Anything).Return(nil)
	mockUpdate.On("CheckForUpdate", mock.Anything).Return(nil, nil)

	// Execute with Trend flag and checks
	flags := model.Flags{Trend: true, TrendChecks: trendChecks}
	err := svc.Orchestrate(flags)

	// Assert
	assert.NoError(t, err)
	mockCost.AssertExpectations(t)
}

func TestDefaultWorkflow_CostServiceError(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*services.MockCostService, *services.MockSTSService)
		expectedErr string
	}{
		{
			name: "GetCurrentMonthCostsByService_fails",
			setupMocks: func(mockCost *services.MockCostService, _ *services.MockSTSService) {
				mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return((*model.CostInfo)(nil), errors.New("cost API error"))
			},
			expectedErr: "cost API error",
		},
		{
			name: "GetLastMonthCostsByService_fails",
			setupMocks: func(mockCost *services.MockCostService, _ *services.MockSTSService) {
				mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetLastMonthCostsByService", mock.Anything).Return((*model.CostInfo)(nil), errors.New("last month error"))
			},
			expectedErr: "last month error",
		},
		{
			name: "GetCurrentMonthTotalCosts_fails",
			setupMocks: func(mockCost *services.MockCostService, _ *services.MockSTSService) {
				mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetLastMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetCurrentMonthTotalCosts", mock.Anything).Return((*string)(nil), errors.New("total cost error"))
			},
			expectedErr: "total cost error",
		},
		{
			name: "GetLastMonthTotalCosts_fails",
			setupMocks: func(mockCost *services.MockCostService, _ *services.MockSTSService) {
				mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetLastMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetCurrentMonthTotalCosts", mock.Anything).Return(aws.String("100.00"), nil)
				mockCost.On("GetLastMonthTotalCosts", mock.Anything).Return((*string)(nil), errors.New("last total error"))
			},
			expectedErr: "last total error",
		},
		{
			name: "GetCallerIdentity_fails",
			setupMocks: func(mockCost *services.MockCostService, mockSTS *services.MockSTSService) {
				mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetLastMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetCurrentMonthTotalCosts", mock.Anything).Return(aws.String("100.00"), nil)
				mockCost.On("GetLastMonthTotalCosts", mock.Anything).Return(aws.String("90.00"), nil)
				mockSTS.On("GetCallerIdentity", mock.Anything).Return((*sts.GetCallerIdentityOutput)(nil), errors.New("STS error"))
			},
			expectedErr: "STS error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSTS := new(services.MockSTSService)
			mockCost := new(services.MockCostService)

			mockOutput := new(services.MockOutputService)
			mockUpdate := new(services.MockUpdateService)
			mockReport := new(services.MockReportService)

			tt.setupMocks(mockCost, mockSTS)
			mockOutput.On("StopSpinner").Return().Maybe()
			mockOutput.On("SetSpinnerMessage", mock.Anything).Return().Maybe()
			mockOutput.On("RenderCostComparison", mock.Anything).Return(nil).Maybe()
			mockUpdate.On("CheckForUpdate", mock.Anything).Return(nil, nil).Maybe()

			config := Config{
				Registry: analyzer.NewRegistry(), STSService: mockSTS,
				CostService: mockCost,

				OutputService: mockOutput,
				UpdateService: mockUpdate,
				ReportService: mockReport,

				VersionInfo: model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
			}
			svc := NewService(config)
			err := svc.Orchestrate(model.Flags{Output: "json"})

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestTrendWorkflow_Error(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*services.MockCostService, *services.MockSTSService)
		expectedErr string
	}{
		{
			name: "GetLastSixMonthsCosts_fails",
			setupMocks: func(mockCost *services.MockCostService, _ *services.MockSTSService) {
				mockCost.On("GetLastSixMonthsCosts", mock.Anything, mock.Anything).Return(([]model.CostInfo)(nil), errors.New("trend API error"))
			},
			expectedErr: "trend API error",
		},
		{
			name: "GetCallerIdentity_fails",
			setupMocks: func(mockCost *services.MockCostService, mockSTS *services.MockSTSService) {
				mockCost.On("GetLastSixMonthsCosts", mock.Anything, mock.Anything).Return([]model.CostInfo{}, nil)
				mockSTS.On("GetCallerIdentity", mock.Anything).Return((*sts.GetCallerIdentityOutput)(nil), errors.New("STS error"))
			},
			expectedErr: "STS error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSTS := new(services.MockSTSService)
			mockCost := new(services.MockCostService)

			mockOutput := new(services.MockOutputService)
			mockUpdate := new(services.MockUpdateService)
			mockReport := new(services.MockReportService)

			tt.setupMocks(mockCost, mockSTS)
			mockOutput.On("StopSpinner").Return().Maybe()
			mockOutput.On("SetSpinnerMessage", mock.Anything).Return().Maybe()
			mockOutput.On("RenderTrend", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			mockUpdate.On("CheckForUpdate", mock.Anything).Return(nil, nil).Maybe()

			config := Config{
				Registry: analyzer.NewRegistry(), STSService: mockSTS,
				CostService: mockCost,

				OutputService: mockOutput,
				UpdateService: mockUpdate,
				ReportService: mockReport,

				VersionInfo: model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
			}
			svc := NewService(config)
			err := svc.Orchestrate(model.Flags{Trend: true, Output: "json"})

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestOrchestrate_RouteToReportWorkflow(t *testing.T) {
	// Setup mocks
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)

	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	// Create service
	config := Config{
		Registry: analyzer.NewRegistry(), STSService: mockSTS,
		CostService: mockCost,

		PricingService: new(services.MockPricingService),
		OutputService:  mockOutput,
		UpdateService:  mockUpdate,
		ReportService:  mockReport,

		VersionInfo: model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
	}
	svc := NewService(config)

	// Setup expectations for default workflow + report
	mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	mockCost.On("GetLastMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	mockCost.On("GetCurrentMonthTotalCosts", mock.Anything).Return(aws.String("100.00"), nil)
	mockCost.On("GetLastMonthTotalCosts", mock.Anything).Return(aws.String("90.00"), nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("SetSpinnerMessage", mock.Anything).Return().Maybe()

	// Mock report call
	reportPath := "report.pdf"
	mockReport.On("GenerateCostComparisonReport", mock.Anything, "report.html").Return(&reportPath, nil)
	mockOutput.On("PrintReportSuccess", reportPath).Return()
	mockUpdate.On("CheckForUpdate", mock.Anything).Return(nil, nil)

	// Execute with Report flag
	flags := model.Flags{Report: true, ReportPath: "report.html"}
	err := svc.Orchestrate(flags)

	// Assert
	assert.NoError(t, err)
	mockCost.AssertExpectations(t)
	mockSTS.AssertExpectations(t)
	mockOutput.AssertExpectations(t)
	mockReport.AssertExpectations(t)
	// Verify RenderCostComparison was NOT called
	mockOutput.AssertNotCalled(t, "RenderCostComparison", mock.Anything)
}

type testMocks struct {
	sts            *services.MockSTSService
	cost           *services.MockCostService
	ec2            *services.MockEC2Service
	elb            *services.MockELBService
	s3             *services.MockS3Service
	cloudWatch     *services.MockCloudWatchLogsService
	output         *services.MockOutputService
	update         *services.MockUpdateService
	report         *services.MockReportService
	rds            *services.MockRDSService
	sagemaker      *services.MockSageMakerService
	secretsmanager *services.MockSecretsManagerService
	ecr            *services.MockECRService
	pricing        *services.MockPricingService
	vpc            *services.MockVPCService
}

func newTestServiceWithMocks(versionInfo model.VersionInfo) (Service, *testMocks) {
	m := &testMocks{
		sts:            new(services.MockSTSService),
		cost:           new(services.MockCostService),
		ec2:            new(services.MockEC2Service),
		elb:            new(services.MockELBService),
		s3:             new(services.MockS3Service),
		cloudWatch:     new(services.MockCloudWatchLogsService),
		output:         new(services.MockOutputService),
		update:         new(services.MockUpdateService),
		report:         new(services.MockReportService),
		rds:            new(services.MockRDSService),
		sagemaker:      new(services.MockSageMakerService),
		secretsmanager: new(services.MockSecretsManagerService),
		ecr:            new(services.MockECRService),
		pricing:        new(services.MockPricingService),
		vpc:            new(services.MockVPCService),
	}

	svc := NewService(Config{
		Registry: analyzer.NewRegistry(), STSService: m.sts,
		CostService: m.cost,

		PricingService: m.pricing,

		OutputService: m.output,
		UpdateService: m.update,
		ReportService: m.report,
		VersionInfo:   versionInfo,
	})

	return svc, m
}

func (m *testMocks) setupDefaultWorkflow() {
	m.cost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	m.cost.On("GetLastMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	m.cost.On("GetCurrentMonthTotalCosts", mock.Anything).Return(aws.String("100.00"), nil)
	m.cost.On("GetLastMonthTotalCosts", mock.Anything).Return(aws.String("90.00"), nil)
	m.sts.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	m.output.On("StopSpinner").Return()
	m.output.On("RenderCostComparison", mock.Anything).Return(nil)
}

func TestOrchestrate_DefaultWorkflow_ShowsNewVersionNotification(t *testing.T) {
	versionInfo := model.VersionInfo{Version: "v1.2.0", Commit: "abc", Date: "today"}
	svc, m := newTestServiceWithMocks(versionInfo)
	m.setupDefaultWorkflow()

	latestVersion := testLatestVersion
	m.update.On("CheckForUpdate", mock.Anything).Return(&latestVersion, nil)
	m.output.On("PrintNewVersionAvailable", "v1.2.0", testLatestVersion).Return()

	err := svc.Orchestrate(model.Flags{})
	assert.NoError(t, err)
	m.output.AssertCalled(t, "PrintNewVersionAvailable", "v1.2.0", testLatestVersion)
}

func TestOrchestrate_DefaultWorkflow_VersionCheckError_SilentlyIgnored(t *testing.T) {
	svc, m := newTestServiceWithMocks(model.VersionInfo{Version: "v1.2.0", Commit: "abc", Date: "today"})
	m.setupDefaultWorkflow()

	m.update.On("CheckForUpdate", mock.Anything).Return(nil, errors.New("github error"))

	err := svc.Orchestrate(model.Flags{})
	assert.NoError(t, err)
	m.output.AssertNotCalled(t, "PrintNewVersionAvailable", mock.Anything, mock.Anything)
}

func TestOrchestrate_DefaultWorkflow_JSONFormat_RunsVersionCheck(t *testing.T) {
	versionInfo := model.VersionInfo{Version: "v1.2.0", Commit: "abc", Date: "today"}
	svc, m := newTestServiceWithMocks(versionInfo)
	m.setupDefaultWorkflow()

	latestVersion := testLatestVersion
	m.update.On("CheckForUpdate", mock.Anything).Return(&latestVersion, nil)
	m.output.On("PrintNewVersionAvailable", "v1.2.0", testLatestVersion).Return()

	err := svc.Orchestrate(model.Flags{Output: "json"})
	assert.NoError(t, err)
	m.update.AssertCalled(t, "CheckForUpdate", mock.Anything)
}

func TestOrchestrate_DefaultWorkflow_CSVFormat_RunsVersionCheck(t *testing.T) {
	versionInfo := model.VersionInfo{Version: "v1.2.0", Commit: "abc", Date: "today"}
	svc, m := newTestServiceWithMocks(versionInfo)
	m.setupDefaultWorkflow()

	latestVersion := testLatestVersion
	m.update.On("CheckForUpdate", mock.Anything).Return(&latestVersion, nil)
	m.output.On("PrintNewVersionAvailable", "v1.2.0", testLatestVersion).Return()

	err := svc.Orchestrate(model.Flags{Output: "csv"})
	assert.NoError(t, err)
	m.update.AssertCalled(t, "CheckForUpdate", mock.Anything)
}

func TestOrchestrate_UpdateWorkflow_NoVersionCheck(t *testing.T) {
	svc, m := newTestServiceWithMocks(model.VersionInfo{Version: "v1.2.0", Commit: "abc", Date: "today"})

	m.output.On("StopSpinner").Return()
	m.update.On("Update").Return(nil)

	err := svc.Orchestrate(model.Flags{Update: true})
	assert.NoError(t, err)
	m.update.AssertNotCalled(t, "CheckForUpdate", mock.Anything)
}

func TestOrchestrate_VersionWorkflow_NoVersionCheck(t *testing.T) {
	versionInfo := model.VersionInfo{Version: "v1.2.0", Commit: "abc", Date: "today"}
	svc, m := newTestServiceWithMocks(versionInfo)

	m.output.On("StopSpinner").Return()
	m.output.On("RenderVersion", versionInfo).Return()

	err := svc.Orchestrate(model.Flags{Version: true})
	assert.NoError(t, err)
	m.update.AssertNotCalled(t, "CheckForUpdate", mock.Anything)
}

func TestOrchestrate_UpdateWorkflow_RateLimit(t *testing.T) {
	svc, m := newTestServiceWithMocks(model.VersionInfo{Version: "v1.0.0"})

	m.output.On("StopSpinner").Return()
	m.update.On("Update").Return(model.ErrRateLimit)
	m.output.On("PrintRateLimitError").Return()

	err := svc.Orchestrate(model.Flags{Update: true})

	assert.Error(t, err)
	m.output.AssertExpectations(t)
}

func TestOrchestrate_UpdateWorkflow_Error(t *testing.T) {
	svc, m := newTestServiceWithMocks(model.VersionInfo{Version: "v1.0.0"})

	m.output.On("StopSpinner").Return()
	m.update.On("Update").Return(errors.New("generic error"))
	m.output.On("PrintUpdateError", mock.Anything).Return()

	err := svc.Orchestrate(model.Flags{Update: true})

	assert.Error(t, err)
	m.output.AssertExpectations(t)
}

func TestOrchestrate_HandleCostError_FirstDayOfMonth(t *testing.T) {
	svc, m := newTestServiceWithMocks(model.VersionInfo{Version: "v1.0.0"})

	m.cost.On("GetCurrentMonthCostsByService", mock.Anything).Return((*model.CostInfo)(nil), model.ErrFirstDayOfMonth)
	m.output.On("StopSpinner").Return()
	m.output.On("PrintFirstDayOfMonthError").Return()
	m.update.On("CheckForUpdate", mock.Anything).Return(nil, nil).Maybe()

	err := svc.Orchestrate(model.Flags{})

	assert.NoError(t, err)
	m.output.AssertExpectations(t)
}

func TestShouldRunCheck(t *testing.T) {
	tests := []struct {
		name        string
		wasteChecks []string
		checkName   string
		want        bool
	}{
		{"empty_checks_runs_all", []string{}, "ec2", true},
		{"nil_checks_runs_all", nil, "ec2", true},
		{"exact_match", []string{"ec2"}, "ec2", true},
		{"case_insensitive_match", []string{"EC2"}, "ec2", true},
		{"no_match", []string{"s3"}, "ec2", false},
		{"multiple_checks_match", []string{"s3", "ec2", "rds"}, "ec2", true},
		{"multiple_checks_no_match", []string{"s3", "rds"}, "ec2", false},
		{"hyphenated_name", []string{"secrets-manager"}, "secrets-manager", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldRunCheck(tt.wasteChecks, tt.checkName)
			if got != tt.want {
				t.Errorf("shouldRunCheck(%v, %q) = %v, want %v", tt.wasteChecks, tt.checkName, got, tt.want)
			}
		})
	}
}
