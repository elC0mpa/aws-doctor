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
