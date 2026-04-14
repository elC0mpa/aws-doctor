package orchestrator

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrchestrate_RouteToDefaultWorkflow(t *testing.T) {
	// Setup mocks
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockEC2 := new(services.MockEC2Service)
	mockELB := new(services.MockELBService)
	mockS3 := new(services.MockS3Service)
	mockCloudWatch := new(services.MockCloudWatchLogsService)
	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	// Create service
	mockRDS := new(services.MockRDSService)
	config := Config{
		STSService:            mockSTS,
		CostService:           mockCost,
		EC2Service:            mockEC2,
		ELBService:            mockELB,
		S3Service:             mockS3,
		CloudWatchLogsService: mockCloudWatch,
		RDSService:            mockRDS,
		OutputService:         mockOutput,
		UpdateService:         mockUpdate,
		ReportService:         mockReport,
		VersionInfo:           model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
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
	mockOutput.On("RenderCostComparison", mock.Anything).Return(nil)

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
	mockEC2 := new(services.MockEC2Service)
	mockELB := new(services.MockELBService)
	mockS3 := new(services.MockS3Service)
	mockCloudWatch := new(services.MockCloudWatchLogsService)
	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	// Create service
	mockRDS := new(services.MockRDSService)
	config := Config{
		STSService:            mockSTS,
		CostService:           mockCost,
		EC2Service:            mockEC2,
		ELBService:            mockELB,
		S3Service:             mockS3,
		CloudWatchLogsService: mockCloudWatch,
		RDSService:            mockRDS,
		OutputService:         mockOutput,
		UpdateService:         mockUpdate,
		ReportService:         mockReport,
		VersionInfo:           model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
	}
	svc := NewService(config)
	// Setup expectations
	mockOutput.On("StopSpinner").Return()
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
	mockEC2 := new(services.MockEC2Service)
	mockELB := new(services.MockELBService)
	mockS3 := new(services.MockS3Service)
	mockCloudWatch := new(services.MockCloudWatchLogsService)
	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)
	mockRDS := new(services.MockRDSService)

	config := Config{
		STSService:            mockSTS,
		CostService:           mockCost,
		EC2Service:            mockEC2,
		ELBService:            mockELB,
		S3Service:             mockS3,
		CloudWatchLogsService: mockCloudWatch,
		RDSService:            mockRDS,
		OutputService:         mockOutput,
		UpdateService:         mockUpdate,
		ReportService:         mockReport,
		VersionInfo:           model.VersionInfo{Version: "v1.0.0", Commit: "abc", Date: "2024-01-01"},
	}
	svc := NewService(config)

	mockOutput.On("StopSpinner").Return()
	mockUpdate.On("Update").Return(model.ErrHomebrewInstall)
	mockOutput.On("PrintHomebrewUpdate").Return()

	flags := model.Flags{Update: true}
	err := svc.Orchestrate(flags)

	assert.NoError(t, err)
	mockOutput.AssertExpectations(t)
	mockUpdate.AssertExpectations(t)
}

func TestOrchestrate_RouteToVersionWorkflow(t *testing.T) {
	// Setup mocks
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockEC2 := new(services.MockEC2Service)
	mockELB := new(services.MockELBService)
	mockS3 := new(services.MockS3Service)
	mockCloudWatch := new(services.MockCloudWatchLogsService)
	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	// Create service
	mockRDS := new(services.MockRDSService)
	versionInfo := model.VersionInfo{Version: "v1.2.3", Commit: "abc", Date: "today"}
	config := Config{
		STSService:            mockSTS,
		CostService:           mockCost,
		EC2Service:            mockEC2,
		ELBService:            mockELB,
		S3Service:             mockS3,
		CloudWatchLogsService: mockCloudWatch,
		RDSService:            mockRDS,
		OutputService:         mockOutput,
		UpdateService:         mockUpdate,
		ReportService:         mockReport,
		VersionInfo:           versionInfo,
	}
	svc := NewService(config)

	// Setup expectations
	mockOutput.On("StopSpinner").Return()
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
	mockEC2 := new(services.MockEC2Service)
	mockELB := new(services.MockELBService)
	mockS3 := new(services.MockS3Service)
	mockCloudWatch := new(services.MockCloudWatchLogsService)
	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	// Create service
	mockRDS := new(services.MockRDSService)
	config := Config{
		STSService:            mockSTS,
		CostService:           mockCost,
		EC2Service:            mockEC2,
		ELBService:            mockELB,
		S3Service:             mockS3,
		CloudWatchLogsService: mockCloudWatch,
		RDSService:            mockRDS,
		OutputService:         mockOutput,
		UpdateService:         mockUpdate,
		ReportService:         mockReport,
		VersionInfo:           model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
	}
	svc := NewService(config)
	// Setup expectations for trend workflow
	mockCost.On("GetLastSixMonthsCosts", mock.Anything, mock.Anything).Return([]model.CostInfo{}, nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("RenderTrend", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Execute with Trend flag
	flags := model.Flags{Trend: true, Output: "json"}
	err := svc.Orchestrate(flags)

	// Assert
	assert.NoError(t, err)
	mockCost.AssertExpectations(t)
	mockSTS.AssertExpectations(t)
	mockOutput.AssertExpectations(t)
}

func TestOrchestrate_RouteToWasteWorkflow(t *testing.T) {
	// Setup mocks
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockEC2 := new(services.MockEC2Service)
	mockELB := new(services.MockELBService)
	mockS3 := new(services.MockS3Service)
	mockCloudWatch := new(services.MockCloudWatchLogsService)
	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	// Create service
	mockRDS := new(services.MockRDSService)
	mockVPC := new(services.MockVPCService)
	config := Config{
		STSService:            mockSTS,
		CostService:           mockCost,
		EC2Service:            mockEC2,
		ELBService:            mockELB,
		S3Service:             mockS3,
		CloudWatchLogsService: mockCloudWatch,
		RDSService:            mockRDS,
		VPCService:            mockVPC,
		OutputService:         mockOutput,
		UpdateService:         mockUpdate,
		ReportService:         mockReport,
		VersionInfo:           model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
	}
	svc := NewService(config)
	// Setup expectations for waste workflow
	mockEC2.On("GetUnusedElasticIPAddressesInfo", mock.Anything).Return([]types.Address{}, nil)
	mockEC2.On("GetUnusedEBSVolumes", mock.Anything).Return([]types.Volume{}, nil)
	mockEC2.On("GetStoppedInstancesInfo", mock.Anything).Return([]types.Instance{}, []types.Volume{}, nil)
	mockEC2.On("GetReservedInstanceExpiringOrExpired30DaysWaste", mock.Anything).Return([]model.RiExpirationInfo{}, nil)
	mockEC2.On("GetUnusedAMIs", mock.Anything, mock.Anything).Return([]model.AMIWasteInfo{}, nil)
	mockEC2.On("GetOrphanedSnapshots", mock.Anything, mock.Anything).Return([]model.SnapshotWasteInfo{}, nil)
	mockEC2.On("GetUnusedKeyPairs", mock.Anything).Return([]model.KeyPairWasteInfo{}, nil)
	mockVPC.On("GetIdleNATGateways", mock.Anything, mock.Anything).Return([]model.NATGatewayWasteInfo{}, nil)
	mockS3.On("GetS3Waste", mock.Anything).Return([]model.S3BucketWasteInfo{}, []model.S3MultipartUploadWasteInfo{}, nil)
	mockCloudWatch.On("GetCloudWatchLogsWaste", mock.Anything).Return([]model.CloudWatchLogsWasteInfo{}, nil)
	mockRDS.On("GetRDSWaste", mock.Anything).Return([]model.RDSInstanceWasteInfo{}, []model.RDSSnapshotWasteInfo{}, []model.RDSIdleInstanceInfo{}, nil)
	mockELB.On("GetLoadBalancerWaste", mock.Anything).Return([]elbtypes.LoadBalancer{}, []model.ELBIdleInfo{}, nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("RenderWaste", mock.Anything).Return(nil)

	// Execute with Waste flag
	flags := model.Flags{Waste: true, Output: "json"}
	err := svc.Orchestrate(flags)

	// Assert
	assert.NoError(t, err)
	mockEC2.AssertExpectations(t)
	mockELB.AssertExpectations(t)
	mockSTS.AssertExpectations(t)
	mockOutput.AssertExpectations(t)
}

func TestOrchestrate_WasteTakesPrecedenceOverTrend(t *testing.T) {
	// Setup mocks
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockEC2 := new(services.MockEC2Service)
	mockELB := new(services.MockELBService)
	mockS3 := new(services.MockS3Service)
	mockCloudWatch := new(services.MockCloudWatchLogsService)
	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	// Create service
	mockRDS := new(services.MockRDSService)
	mockVPC := new(services.MockVPCService)
	config := Config{
		STSService:            mockSTS,
		CostService:           mockCost,
		EC2Service:            mockEC2,
		ELBService:            mockELB,
		S3Service:             mockS3,
		CloudWatchLogsService: mockCloudWatch,
		RDSService:            mockRDS,
		VPCService:            mockVPC,
		OutputService:         mockOutput,
		UpdateService:         mockUpdate,
		ReportService:         mockReport,
		VersionInfo:           model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
	}
	svc := NewService(config)
	// Setup expectations for waste workflow (should be called, not trend)
	mockEC2.On("GetUnusedElasticIPAddressesInfo", mock.Anything).Return([]types.Address{}, nil)
	mockEC2.On("GetUnusedEBSVolumes", mock.Anything).Return([]types.Volume{}, nil)
	mockEC2.On("GetStoppedInstancesInfo", mock.Anything).Return([]types.Instance{}, []types.Volume{}, nil)
	mockEC2.On("GetReservedInstanceExpiringOrExpired30DaysWaste", mock.Anything).Return([]model.RiExpirationInfo{}, nil)
	mockEC2.On("GetUnusedAMIs", mock.Anything, mock.Anything).Return([]model.AMIWasteInfo{}, nil)
	mockEC2.On("GetOrphanedSnapshots", mock.Anything, mock.Anything).Return([]model.SnapshotWasteInfo{}, nil)
	mockEC2.On("GetUnusedKeyPairs", mock.Anything).Return([]model.KeyPairWasteInfo{}, nil)
	mockVPC.On("GetIdleNATGateways", mock.Anything, mock.Anything).Return([]model.NATGatewayWasteInfo{}, nil)
	mockS3.On("GetS3Waste", mock.Anything).Return([]model.S3BucketWasteInfo{}, []model.S3MultipartUploadWasteInfo{}, nil)
	mockCloudWatch.On("GetCloudWatchLogsWaste", mock.Anything).Return([]model.CloudWatchLogsWasteInfo{}, nil)
	mockRDS.On("GetRDSWaste", mock.Anything).Return([]model.RDSInstanceWasteInfo{}, []model.RDSSnapshotWasteInfo{}, []model.RDSIdleInstanceInfo{}, nil)
	mockELB.On("GetLoadBalancerWaste", mock.Anything).Return([]elbtypes.LoadBalancer{}, []model.ELBIdleInfo{}, nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("RenderWaste", mock.Anything).Return(nil)

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
	mockEC2 := new(services.MockEC2Service)
	mockELB := new(services.MockELBService)
	mockS3 := new(services.MockS3Service)
	mockCloudWatch := new(services.MockCloudWatchLogsService)
	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)

	// Create service
	mockRDS := new(services.MockRDSService)
	config := Config{
		STSService:            mockSTS,
		CostService:           mockCost,
		EC2Service:            mockEC2,
		ELBService:            mockELB,
		S3Service:             mockS3,
		CloudWatchLogsService: mockCloudWatch,
		RDSService:            mockRDS,
		OutputService:         mockOutput,
		UpdateService:         mockUpdate,
		ReportService:         mockReport,
		VersionInfo:           model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
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
	mockOutput.On("RenderTrend", "123456789012", mock.Anything, mock.Anything).Return(nil)

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
			mockEC2 := new(services.MockEC2Service)
			mockELB := new(services.MockELBService)
			mockS3 := new(services.MockS3Service)
			mockCloudWatch := new(services.MockCloudWatchLogsService)
			mockOutput := new(services.MockOutputService)
			mockUpdate := new(services.MockUpdateService)
			mockReport := new(services.MockReportService)

			tt.setupMocks(mockCost, mockSTS)
			mockOutput.On("StopSpinner").Return().Maybe()
			mockOutput.On("RenderCostComparison", mock.Anything).Return(nil).Maybe()

			mockRDS := new(services.MockRDSService)
			config := Config{
				STSService:            mockSTS,
				CostService:           mockCost,
				EC2Service:            mockEC2,
				ELBService:            mockELB,
				S3Service:             mockS3,
				CloudWatchLogsService: mockCloudWatch,
				RDSService:            mockRDS,
				OutputService:         mockOutput,
				UpdateService:         mockUpdate,
				ReportService:         mockReport,
				VersionInfo:           model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
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
			mockEC2 := new(services.MockEC2Service)
			mockELB := new(services.MockELBService)
			mockS3 := new(services.MockS3Service)
			mockCloudWatch := new(services.MockCloudWatchLogsService)
			mockOutput := new(services.MockOutputService)
			mockUpdate := new(services.MockUpdateService)
			mockReport := new(services.MockReportService)

			tt.setupMocks(mockCost, mockSTS)
			mockOutput.On("StopSpinner").Return().Maybe()
			mockOutput.On("RenderTrend", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

			mockRDS := new(services.MockRDSService)
			config := Config{
				STSService:            mockSTS,
				CostService:           mockCost,
				EC2Service:            mockEC2,
				ELBService:            mockELB,
				S3Service:             mockS3,
				CloudWatchLogsService: mockCloudWatch,
				RDSService:            mockRDS,
				OutputService:         mockOutput,
				UpdateService:         mockUpdate,
				ReportService:         mockReport,
				VersionInfo:           model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
			}
			svc := NewService(config)
			err := svc.Orchestrate(model.Flags{Trend: true, Output: "json"})

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestWasteWorkflow_Error(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*services.MockEC2Service, *services.MockELBService, *services.MockS3Service, *services.MockCloudWatchLogsService, *services.MockRDSService, *services.MockSTSService, *services.MockVPCService)
		expectedErr string
	}{
		{
			name: "GetUnusedElasticIpAddressesInfo_fails",
			setupMocks: func(mockEC2 *services.MockEC2Service, mockELB *services.MockELBService, mockS3 *services.MockS3Service, mockCloudWatch *services.MockCloudWatchLogsService, mockRDS *services.MockRDSService, mockSTS *services.MockSTSService, mockVPC *services.MockVPCService) {
				mockEC2.On("GetUnusedElasticIPAddressesInfo", mock.Anything).Return(([]types.Address)(nil), errors.New("EIP error"))
				mockEC2.On("GetUnusedEBSVolumes", mock.Anything).Return([]types.Volume{}, nil)
				mockEC2.On("GetStoppedInstancesInfo", mock.Anything).Return([]types.Instance{}, []types.Volume{}, nil)
				mockEC2.On("GetReservedInstanceExpiringOrExpired30DaysWaste", mock.Anything).Return([]model.RiExpirationInfo{}, nil)
				mockEC2.On("GetUnusedAMIs", mock.Anything, mock.Anything).Return([]model.AMIWasteInfo{}, nil)
				mockEC2.On("GetOrphanedSnapshots", mock.Anything, mock.Anything).Return([]model.SnapshotWasteInfo{}, nil)
				mockEC2.On("GetUnusedKeyPairs", mock.Anything).Return([]model.KeyPairWasteInfo{}, nil)
				mockVPC.On("GetIdleNATGateways", mock.Anything, mock.Anything).Return([]model.NATGatewayWasteInfo{}, nil)
				mockS3.On("GetS3Waste", mock.Anything).Return([]model.S3BucketWasteInfo{}, []model.S3MultipartUploadWasteInfo{}, nil)
				mockCloudWatch.On("GetCloudWatchLogsWaste", mock.Anything).Return([]model.CloudWatchLogsWasteInfo{}, nil)
				mockRDS.On("GetRDSWaste", mock.Anything).Return([]model.RDSInstanceWasteInfo{}, []model.RDSSnapshotWasteInfo{}, []model.RDSIdleInstanceInfo{}, nil)

				mockELB.On("GetLoadBalancerWaste", mock.Anything).Return([]elbtypes.LoadBalancer{}, []model.ELBIdleInfo{}, nil)
				mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
					Account: aws.String("123456789012"),
				}, nil)
			},
			expectedErr: "EIP error",
		},
		{
			name: "GetUnusedEBSVolumes_fails",
			setupMocks: func(mockEC2 *services.MockEC2Service, mockELB *services.MockELBService, mockS3 *services.MockS3Service, mockCloudWatch *services.MockCloudWatchLogsService, mockRDS *services.MockRDSService, mockSTS *services.MockSTSService, mockVPC *services.MockVPCService) {
				mockEC2.On("GetUnusedElasticIPAddressesInfo", mock.Anything).Return([]types.Address{}, nil)
				mockEC2.On("GetUnusedEBSVolumes", mock.Anything).Return(([]types.Volume)(nil), errors.New("EBS error"))
				mockEC2.On("GetStoppedInstancesInfo", mock.Anything).Return([]types.Instance{}, []types.Volume{}, nil)
				mockEC2.On("GetReservedInstanceExpiringOrExpired30DaysWaste", mock.Anything).Return([]model.RiExpirationInfo{}, nil)
				mockEC2.On("GetUnusedAMIs", mock.Anything, mock.Anything).Return([]model.AMIWasteInfo{}, nil)
				mockEC2.On("GetOrphanedSnapshots", mock.Anything, mock.Anything).Return([]model.SnapshotWasteInfo{}, nil)
				mockEC2.On("GetUnusedKeyPairs", mock.Anything).Return([]model.KeyPairWasteInfo{}, nil)
				mockVPC.On("GetIdleNATGateways", mock.Anything, mock.Anything).Return([]model.NATGatewayWasteInfo{}, nil)
				mockS3.On("GetS3Waste", mock.Anything).Return([]model.S3BucketWasteInfo{}, []model.S3MultipartUploadWasteInfo{}, nil)
				mockCloudWatch.On("GetCloudWatchLogsWaste", mock.Anything).Return([]model.CloudWatchLogsWasteInfo{}, nil)
				mockRDS.On("GetRDSWaste", mock.Anything).Return([]model.RDSInstanceWasteInfo{}, []model.RDSSnapshotWasteInfo{}, []model.RDSIdleInstanceInfo{}, nil)

				mockELB.On("GetLoadBalancerWaste", mock.Anything).Return([]elbtypes.LoadBalancer{}, []model.ELBIdleInfo{}, nil)
				mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
					Account: aws.String("123456789012"),
				}, nil)
			},
			expectedErr: "EBS error",
		},
		{
			name: "GetLoadBalancerWaste_fails",
			setupMocks: func(mockEC2 *services.MockEC2Service, mockELB *services.MockELBService, mockS3 *services.MockS3Service, mockCloudWatch *services.MockCloudWatchLogsService, mockRDS *services.MockRDSService, mockSTS *services.MockSTSService, mockVPC *services.MockVPCService) {
				mockEC2.On("GetUnusedElasticIPAddressesInfo", mock.Anything).Return([]types.Address{}, nil)
				mockEC2.On("GetUnusedEBSVolumes", mock.Anything).Return([]types.Volume{}, nil)
				mockEC2.On("GetStoppedInstancesInfo", mock.Anything).Return([]types.Instance{}, []types.Volume{}, nil)
				mockEC2.On("GetReservedInstanceExpiringOrExpired30DaysWaste", mock.Anything).Return([]model.RiExpirationInfo{}, nil)
				mockEC2.On("GetUnusedAMIs", mock.Anything, mock.Anything).Return([]model.AMIWasteInfo{}, nil)
				mockEC2.On("GetOrphanedSnapshots", mock.Anything, mock.Anything).Return([]model.SnapshotWasteInfo{}, nil)
				mockEC2.On("GetUnusedKeyPairs", mock.Anything).Return([]model.KeyPairWasteInfo{}, nil)
				mockVPC.On("GetIdleNATGateways", mock.Anything, mock.Anything).Return([]model.NATGatewayWasteInfo{}, nil)
				mockS3.On("GetS3Waste", mock.Anything).Return([]model.S3BucketWasteInfo{}, []model.S3MultipartUploadWasteInfo{}, nil)
				mockCloudWatch.On("GetCloudWatchLogsWaste", mock.Anything).Return([]model.CloudWatchLogsWasteInfo{}, nil)
				mockRDS.On("GetRDSWaste", mock.Anything).Return([]model.RDSInstanceWasteInfo{}, []model.RDSSnapshotWasteInfo{}, []model.RDSIdleInstanceInfo{}, nil)

				mockELB.On("GetLoadBalancerWaste", mock.Anything).Return(nil, nil, errors.New("ELB error"))
				mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
					Account: aws.String("123456789012"),
				}, nil)
			},
			expectedErr: "ELB error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSTS := new(services.MockSTSService)
			mockCost := new(services.MockCostService)
			mockEC2 := new(services.MockEC2Service)
			mockELB := new(services.MockELBService)
			mockS3 := new(services.MockS3Service)
			mockCloudWatch := new(services.MockCloudWatchLogsService)
			mockOutput := new(services.MockOutputService)
			mockUpdate := new(services.MockUpdateService)
			mockRDS := new(services.MockRDSService)
			mockReport := new(services.MockReportService)
			mockVPC := new(services.MockVPCService)

			tt.setupMocks(mockEC2, mockELB, mockS3, mockCloudWatch, mockRDS, mockSTS, mockVPC)
			mockOutput.On("StopSpinner").Return().Maybe()
			mockOutput.On("RenderWaste", mock.Anything).Return(nil).Maybe()

			config := Config{
				STSService:            mockSTS,
				CostService:           mockCost,
				EC2Service:            mockEC2,
				ELBService:            mockELB,
				S3Service:             mockS3,
				CloudWatchLogsService: mockCloudWatch,
				RDSService:            mockRDS,
				VPCService:            mockVPC,
				OutputService:         mockOutput,
				UpdateService:         mockUpdate,
				ReportService:         mockReport,
				VersionInfo:           model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
			}
			svc := NewService(config)
			err := svc.Orchestrate(model.Flags{Waste: true, Output: "json"})

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestOrchestrate_RouteToReportWorkflow(t *testing.T) {
	// Setup mocks
	mockSTS := new(services.MockSTSService)
	mockCost := new(services.MockCostService)
	mockEC2 := new(services.MockEC2Service)
	mockELB := new(services.MockELBService)
	mockS3 := new(services.MockS3Service)
	mockCloudWatch := new(services.MockCloudWatchLogsService)
	mockOutput := new(services.MockOutputService)
	mockUpdate := new(services.MockUpdateService)
	mockReport := new(services.MockReportService)
	mockRDS := new(services.MockRDSService)

	// Create service
	config := Config{
		STSService:            mockSTS,
		CostService:           mockCost,
		EC2Service:            mockEC2,
		ELBService:            mockELB,
		S3Service:             mockS3,
		CloudWatchLogsService: mockCloudWatch,
		RDSService:            mockRDS,
		OutputService:         mockOutput,
		UpdateService:         mockUpdate,
		ReportService:         mockReport,
		VersionInfo:           model.VersionInfo{Version: "dev", Commit: "none", Date: "unknown"},
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

	// Mock report call
	reportPath := "report.pdf"
	mockReport.On("GenerateCostComparisonReport", mock.Anything, "report.html").Return(&reportPath, nil)
	mockOutput.On("PrintReportSuccess", reportPath).Return()

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
