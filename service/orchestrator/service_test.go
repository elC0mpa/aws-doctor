package orchestrator

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/elC0mpa/aws-doctor/mocks"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrchestrate_RouteToDefaultWorkflow(t *testing.T) {
	// Setup mocks
	mockSTS := new(mocks.MockSTSService)
	mockCost := new(mocks.MockCostService)
	mockEC2 := new(mocks.MockEC2Service)
	mockELB := new(mocks.MockELBService)
	mockRoute53 := new(mocks.MockRoute53Service)
	mockOutput := new(mocks.MockOutputService)

	// Create service
	svc := NewService(mockSTS, mockCost, mockEC2, mockELB, mockRoute53, mockOutput)

	// Setup expectations for default workflow
	mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	mockCost.On("GetLastMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
	mockCost.On("GetCurrentMonthTotalCosts", mock.Anything).Return(aws.String("100.00"), nil)
	mockCost.On("GetLastMonthTotalCosts", mock.Anything).Return(aws.String("90.00"), nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("RenderCostComparison", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Execute
	flags := model.Flags{Output: "json"}
	err := svc.Orchestrate(flags)

	// Assert
	assert.NoError(t, err)
	mockCost.AssertExpectations(t)
	mockSTS.AssertExpectations(t)
	mockOutput.AssertExpectations(t)
}

func TestOrchestrate_RouteToTrendWorkflow(t *testing.T) {
	// Setup mocks
	mockSTS := new(mocks.MockSTSService)
	mockCost := new(mocks.MockCostService)
	mockEC2 := new(mocks.MockEC2Service)
	mockELB := new(mocks.MockELBService)
	mockRoute53 := new(mocks.MockRoute53Service)
	mockOutput := new(mocks.MockOutputService)

	// Create service
	svc := NewService(mockSTS, mockCost, mockEC2, mockELB, mockRoute53, mockOutput)

	// Setup expectations for trend workflow
	mockCost.On("GetLastSixMonthsCosts", mock.Anything).Return([]model.CostInfo{}, nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("RenderTrend", mock.Anything, mock.Anything).Return(nil)

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
	mockSTS := new(mocks.MockSTSService)
	mockCost := new(mocks.MockCostService)
	mockEC2 := new(mocks.MockEC2Service)
	mockELB := new(mocks.MockELBService)
	mockRoute53 := new(mocks.MockRoute53Service)
	mockOutput := new(mocks.MockOutputService)

	// Create service
	svc := NewService(mockSTS, mockCost, mockEC2, mockELB, mockRoute53, mockOutput)

	// Setup expectations for waste workflow
	mockEC2.On("GetUnusedElasticIpAddressesInfo", mock.Anything).Return([]types.Address{}, nil)
	mockEC2.On("GetUnusedEBSVolumes", mock.Anything).Return([]types.Volume{}, nil)
	mockEC2.On("GetStoppedInstancesInfo", mock.Anything).Return([]types.Instance{}, []types.Volume{}, nil)
	mockEC2.On("GetReservedInstanceExpiringOrExpired30DaysWaste", mock.Anything).Return([]model.RiExpirationInfo{}, nil)
	mockEC2.On("GetUnusedAMIs", mock.Anything, mock.Anything).Return([]model.AMIWasteInfo{}, nil)
	mockEC2.On("GetOrphanedSnapshots", mock.Anything, mock.Anything).Return([]model.SnapshotWasteInfo{}, nil)
	mockELB.On("GetUnusedLoadBalancers", mock.Anything).Return([]elbtypes.LoadBalancer{}, nil)
	mockRoute53.On("GetEmptyHostedZones", mock.Anything).Return([]model.HostedZoneWasteInfo{}, nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("RenderWaste", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Execute with Waste flag
	flags := model.Flags{Waste: true, Output: "json"}
	err := svc.Orchestrate(flags)

	// Assert
	assert.NoError(t, err)
	mockEC2.AssertExpectations(t)
	mockELB.AssertExpectations(t)
	mockRoute53.AssertExpectations(t)
	mockSTS.AssertExpectations(t)
	mockOutput.AssertExpectations(t)
}

func TestOrchestrate_WasteTakesPrecedenceOverTrend(t *testing.T) {
	// Setup mocks
	mockSTS := new(mocks.MockSTSService)
	mockCost := new(mocks.MockCostService)
	mockEC2 := new(mocks.MockEC2Service)
	mockELB := new(mocks.MockELBService)
	mockRoute53 := new(mocks.MockRoute53Service)
	mockOutput := new(mocks.MockOutputService)

	// Create service
	svc := NewService(mockSTS, mockCost, mockEC2, mockELB, mockRoute53, mockOutput)

	// Setup expectations for waste workflow (should be called, not trend)
	mockEC2.On("GetUnusedElasticIpAddressesInfo", mock.Anything).Return([]types.Address{}, nil)
	mockEC2.On("GetUnusedEBSVolumes", mock.Anything).Return([]types.Volume{}, nil)
	mockEC2.On("GetStoppedInstancesInfo", mock.Anything).Return([]types.Instance{}, []types.Volume{}, nil)
	mockEC2.On("GetReservedInstanceExpiringOrExpired30DaysWaste", mock.Anything).Return([]model.RiExpirationInfo{}, nil)
	mockEC2.On("GetUnusedAMIs", mock.Anything, mock.Anything).Return([]model.AMIWasteInfo{}, nil)
	mockEC2.On("GetOrphanedSnapshots", mock.Anything, mock.Anything).Return([]model.SnapshotWasteInfo{}, nil)
	mockELB.On("GetUnusedLoadBalancers", mock.Anything).Return([]elbtypes.LoadBalancer{}, nil)
	mockRoute53.On("GetEmptyHostedZones", mock.Anything).Return([]model.HostedZoneWasteInfo{}, nil)
	mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)
	mockOutput.On("StopSpinner").Return()
	mockOutput.On("RenderWaste", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Execute with both flags - Waste should take precedence
	flags := model.Flags{Waste: true, Trend: true, Output: "json"}
	err := svc.Orchestrate(flags)

	// Assert - cost service should NOT be called for trend
	assert.NoError(t, err)
	mockCost.AssertNotCalled(t, "GetLastSixMonthsCosts", mock.Anything)
}

func TestDefaultWorkflow_CostServiceError(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mocks.MockCostService, *mocks.MockSTSService)
		expectedErr string
	}{
		{
			name: "GetCurrentMonthCostsByService_fails",
			setupMocks: func(mockCost *mocks.MockCostService, mockSTS *mocks.MockSTSService) {
				mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(nil, errors.New("cost API error"))
			},
			expectedErr: "cost API error",
		},
		{
			name: "GetLastMonthCostsByService_fails",
			setupMocks: func(mockCost *mocks.MockCostService, mockSTS *mocks.MockSTSService) {
				mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetLastMonthCostsByService", mock.Anything).Return(nil, errors.New("last month error"))
			},
			expectedErr: "last month error",
		},
		{
			name: "GetCurrentMonthTotalCosts_fails",
			setupMocks: func(mockCost *mocks.MockCostService, mockSTS *mocks.MockSTSService) {
				mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetLastMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetCurrentMonthTotalCosts", mock.Anything).Return(nil, errors.New("total cost error"))
			},
			expectedErr: "total cost error",
		},
		{
			name: "GetLastMonthTotalCosts_fails",
			setupMocks: func(mockCost *mocks.MockCostService, mockSTS *mocks.MockSTSService) {
				mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetLastMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetCurrentMonthTotalCosts", mock.Anything).Return(aws.String("100.00"), nil)
				mockCost.On("GetLastMonthTotalCosts", mock.Anything).Return(nil, errors.New("last total error"))
			},
			expectedErr: "last total error",
		},
		{
			name: "GetCallerIdentity_fails",
			setupMocks: func(mockCost *mocks.MockCostService, mockSTS *mocks.MockSTSService) {
				mockCost.On("GetCurrentMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetLastMonthCostsByService", mock.Anything).Return(&model.CostInfo{}, nil)
				mockCost.On("GetCurrentMonthTotalCosts", mock.Anything).Return(aws.String("100.00"), nil)
				mockCost.On("GetLastMonthTotalCosts", mock.Anything).Return(aws.String("90.00"), nil)
				mockSTS.On("GetCallerIdentity", mock.Anything).Return(nil, errors.New("STS error"))
			},
			expectedErr: "STS error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSTS := new(mocks.MockSTSService)
			mockCost := new(mocks.MockCostService)
			mockEC2 := new(mocks.MockEC2Service)
			mockELB := new(mocks.MockELBService)
			mockRoute53 := new(mocks.MockRoute53Service)
			mockOutput := new(mocks.MockOutputService)

			tt.setupMocks(mockCost, mockSTS)
			mockOutput.On("StopSpinner").Return().Maybe()
			mockOutput.On("RenderCostComparison", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

			svc := NewService(mockSTS, mockCost, mockEC2, mockELB, mockRoute53, mockOutput)
			err := svc.Orchestrate(model.Flags{Output: "json"})

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestTrendWorkflow_Error(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mocks.MockCostService, *mocks.MockSTSService)
		expectedErr string
	}{
		{
			name: "GetLastSixMonthsCosts_fails",
			setupMocks: func(mockCost *mocks.MockCostService, mockSTS *mocks.MockSTSService) {
				mockCost.On("GetLastSixMonthsCosts", mock.Anything).Return(nil, errors.New("trend API error"))
			},
			expectedErr: "trend API error",
		},
		{
			name: "GetCallerIdentity_fails",
			setupMocks: func(mockCost *mocks.MockCostService, mockSTS *mocks.MockSTSService) {
				mockCost.On("GetLastSixMonthsCosts", mock.Anything).Return([]model.CostInfo{}, nil)
				mockSTS.On("GetCallerIdentity", mock.Anything).Return(nil, errors.New("STS error"))
			},
			expectedErr: "STS error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSTS := new(mocks.MockSTSService)
			mockCost := new(mocks.MockCostService)
			mockEC2 := new(mocks.MockEC2Service)
			mockELB := new(mocks.MockELBService)
			mockRoute53 := new(mocks.MockRoute53Service)
			mockOutput := new(mocks.MockOutputService)

			tt.setupMocks(mockCost, mockSTS)
			mockOutput.On("StopSpinner").Return().Maybe()
			mockOutput.On("RenderTrend", mock.Anything, mock.Anything).Return(nil).Maybe()

			svc := NewService(mockSTS, mockCost, mockEC2, mockELB, mockRoute53, mockOutput)
			err := svc.Orchestrate(model.Flags{Trend: true, Output: "json"})

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestWasteWorkflow_Error(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mocks.MockEC2Service, *mocks.MockELBService, *mocks.MockRoute53Service, *mocks.MockSTSService)
		expectedErr string
	}{
		{
			name: "GetUnusedElasticIpAddressesInfo_fails",
			setupMocks: func(mockEC2 *mocks.MockEC2Service, mockELB *mocks.MockELBService, mockRoute53 *mocks.MockRoute53Service, mockSTS *mocks.MockSTSService) {
				mockEC2.On("GetUnusedElasticIpAddressesInfo", mock.Anything).Return(nil, errors.New("EIP error"))
				mockEC2.On("GetUnusedEBSVolumes", mock.Anything).Return([]types.Volume{}, nil)
				mockEC2.On("GetStoppedInstancesInfo", mock.Anything).Return([]types.Instance{}, []types.Volume{}, nil)
				mockEC2.On("GetReservedInstanceExpiringOrExpired30DaysWaste", mock.Anything).Return([]model.RiExpirationInfo{}, nil)
				mockEC2.On("GetUnusedAMIs", mock.Anything, mock.Anything).Return([]model.AMIWasteInfo{}, nil)
				mockEC2.On("GetOrphanedSnapshots", mock.Anything, mock.Anything).Return([]model.SnapshotWasteInfo{}, nil)
				mockELB.On("GetUnusedLoadBalancers", mock.Anything).Return([]elbtypes.LoadBalancer{}, nil)
				mockRoute53.On("GetEmptyHostedZones", mock.Anything).Return([]model.HostedZoneWasteInfo{}, nil)
				mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
					Account: aws.String("123456789012"),
				}, nil)
			},
			expectedErr: "EIP error",
		},
		{
			name: "GetUnusedEBSVolumes_fails",
			setupMocks: func(mockEC2 *mocks.MockEC2Service, mockELB *mocks.MockELBService, mockRoute53 *mocks.MockRoute53Service, mockSTS *mocks.MockSTSService) {
				mockEC2.On("GetUnusedElasticIpAddressesInfo", mock.Anything).Return([]types.Address{}, nil)
				mockEC2.On("GetUnusedEBSVolumes", mock.Anything).Return(nil, errors.New("EBS error"))
				mockEC2.On("GetStoppedInstancesInfo", mock.Anything).Return([]types.Instance{}, []types.Volume{}, nil)
				mockEC2.On("GetReservedInstanceExpiringOrExpired30DaysWaste", mock.Anything).Return([]model.RiExpirationInfo{}, nil)
				mockEC2.On("GetUnusedAMIs", mock.Anything, mock.Anything).Return([]model.AMIWasteInfo{}, nil)
				mockEC2.On("GetOrphanedSnapshots", mock.Anything, mock.Anything).Return([]model.SnapshotWasteInfo{}, nil)
				mockELB.On("GetUnusedLoadBalancers", mock.Anything).Return([]elbtypes.LoadBalancer{}, nil)
				mockRoute53.On("GetEmptyHostedZones", mock.Anything).Return([]model.HostedZoneWasteInfo{}, nil)
				mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
					Account: aws.String("123456789012"),
				}, nil)
			},
			expectedErr: "EBS error",
		},
		{
			name: "GetUnusedLoadBalancers_fails",
			setupMocks: func(mockEC2 *mocks.MockEC2Service, mockELB *mocks.MockELBService, mockRoute53 *mocks.MockRoute53Service, mockSTS *mocks.MockSTSService) {
				mockEC2.On("GetUnusedElasticIpAddressesInfo", mock.Anything).Return([]types.Address{}, nil)
				mockEC2.On("GetUnusedEBSVolumes", mock.Anything).Return([]types.Volume{}, nil)
				mockEC2.On("GetStoppedInstancesInfo", mock.Anything).Return([]types.Instance{}, []types.Volume{}, nil)
				mockEC2.On("GetReservedInstanceExpiringOrExpired30DaysWaste", mock.Anything).Return([]model.RiExpirationInfo{}, nil)
				mockEC2.On("GetUnusedAMIs", mock.Anything, mock.Anything).Return([]model.AMIWasteInfo{}, nil)
				mockEC2.On("GetOrphanedSnapshots", mock.Anything, mock.Anything).Return([]model.SnapshotWasteInfo{}, nil)
				mockELB.On("GetUnusedLoadBalancers", mock.Anything).Return(nil, errors.New("ELB error"))
				mockRoute53.On("GetEmptyHostedZones", mock.Anything).Return([]model.HostedZoneWasteInfo{}, nil)
				mockSTS.On("GetCallerIdentity", mock.Anything).Return(&sts.GetCallerIdentityOutput{
					Account: aws.String("123456789012"),
				}, nil)
			},
			expectedErr: "ELB error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSTS := new(mocks.MockSTSService)
			mockCost := new(mocks.MockCostService)
			mockEC2 := new(mocks.MockEC2Service)
			mockELB := new(mocks.MockELBService)
			mockRoute53 := new(mocks.MockRoute53Service)
			mockOutput := new(mocks.MockOutputService)

			tt.setupMocks(mockEC2, mockELB, mockRoute53, mockSTS)
			mockOutput.On("StopSpinner").Return().Maybe()
			mockOutput.On("RenderWaste", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

			svc := NewService(mockSTS, mockCost, mockEC2, mockELB, mockRoute53, mockOutput)
			err := svc.Orchestrate(model.Flags{Waste: true, Output: "json"})

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
