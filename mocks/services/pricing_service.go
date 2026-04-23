package services

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockPricingService is a mock of the pricing service
type MockPricingService struct {
	mock.Mock
}

type PricingService interface {
	LoadRegionRates(ctx context.Context) error
	CalculateEBSMonthlyCost(sizeGiB int32, volumeType types.VolumeType) float64
	CalculateEBSSnapshotMonthlyCost(sizeGB int64) float64
	CalculateEIPMonthlyCost() float64
	CalculateLoadBalancerMonthlyCost(lbType elbtypes.LoadBalancerTypeEnum) float64
	CalculateCloudWatchLogsMonthlyCost(storedBytes int64) float64
	CalculateNATGatewayMonthlyCost() float64
	CalculateRDSInstanceMonthlyCost(allocatedGB int32, multiAZ bool) float64
	CalculateRDSSnapshotMonthlyCost(allocatedGB int32) float64
	CalculateRDSIdleInstanceMonthlyCost(instanceClass string, allocatedGB int32, multiAZ bool) float64
	CalculateSageMakerEndpointMonthlyCost(variants []model.SageMakerVariant) float64
}

func (m *MockPricingService) LoadRegionRates(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPricingService) CalculateEBSMonthlyCost(sizeGiB int32, volumeType types.VolumeType) float64 {
	args := m.Called(sizeGiB, volumeType)
	return args.Get(0).(float64)
}

func (m *MockPricingService) CalculateEBSSnapshotMonthlyCost(sizeGB int64) float64 {
	args := m.Called(sizeGB)
	return args.Get(0).(float64)
}

func (m *MockPricingService) CalculateEIPMonthlyCost() float64 {
	args := m.Called()
	return args.Get(0).(float64)
}

func (m *MockPricingService) CalculateLoadBalancerMonthlyCost(lbType elbtypes.LoadBalancerTypeEnum) float64 {
	args := m.Called(lbType)
	return args.Get(0).(float64)
}

func (m *MockPricingService) CalculateCloudWatchLogsMonthlyCost(storedBytes int64) float64 {
	args := m.Called(storedBytes)
	return args.Get(0).(float64)
}

func (m *MockPricingService) CalculateNATGatewayMonthlyCost() float64 {
	args := m.Called()
	return args.Get(0).(float64)
}

func (m *MockPricingService) CalculateRDSInstanceMonthlyCost(allocatedGB int32, multiAZ bool) float64 {
	args := m.Called(allocatedGB, multiAZ)
	return args.Get(0).(float64)
}

func (m *MockPricingService) CalculateRDSSnapshotMonthlyCost(allocatedGB int32) float64 {
	args := m.Called(allocatedGB)
	return args.Get(0).(float64)
}

func (m *MockPricingService) CalculateRDSIdleInstanceMonthlyCost(instanceClass string, allocatedGB int32, multiAZ bool) float64 {
	args := m.Called(instanceClass, allocatedGB, multiAZ)
	return args.Get(0).(float64)
}

func (m *MockPricingService) CalculateSageMakerEndpointMonthlyCost(variants []model.SageMakerVariant) float64 {
	args := m.Called(variants)
	return args.Get(0).(float64)
}

// NewMockPricingService returns a MockPricingService with default expectations set for all methods.
func NewMockPricingService() *MockPricingService {
	m := new(MockPricingService)
	m.On("LoadRegionRates", mock.Anything).Return(nil).Maybe()
	m.On("CalculateEBSMonthlyCost", mock.Anything, mock.Anything).Return(10.0).Maybe()
	m.On("CalculateEBSSnapshotMonthlyCost", mock.Anything).Return(5.0).Maybe()
	m.On("CalculateEIPMonthlyCost").Return(3.65).Maybe()
	m.On("CalculateLoadBalancerMonthlyCost", mock.Anything).Return(16.43).Maybe()
	m.On("CalculateCloudWatchLogsMonthlyCost", mock.Anything).Return(1.0).Maybe()
	m.On("CalculateNATGatewayMonthlyCost").Return(32.85).Maybe()
	m.On("CalculateRDSInstanceMonthlyCost", mock.Anything, mock.Anything).Return(20.0).Maybe()
	m.On("CalculateRDSSnapshotMonthlyCost", mock.Anything).Return(5.0).Maybe()
	m.On("CalculateRDSIdleInstanceMonthlyCost", mock.Anything, mock.Anything, mock.Anything).Return(45.0).Maybe()
	m.On("CalculateSageMakerEndpointMonthlyCost", mock.Anything).Return(50.0).Maybe()
	return m
}
