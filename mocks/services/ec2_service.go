package services

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockEC2Service is a mock implementation of the EC2 service interface.
type MockEC2Service struct {
	mock.Mock
}

// GetElasticIPAddressesInfo mocks the GetElasticIPAddressesInfo method.
func (m *MockEC2Service) GetElasticIPAddressesInfo(ctx context.Context) (*model.ElasticIPInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.ElasticIPInfo), args.Error(1)
}

// GetUnusedElasticIPAddressesInfo mocks the GetUnusedElasticIPAddressesInfo method.
func (m *MockEC2Service) GetUnusedElasticIPAddressesInfo(ctx context.Context) ([]types.Address, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]types.Address), args.Error(1)
}

// GetUnusedEBSVolumes mocks the GetUnusedEBSVolumes method.
func (m *MockEC2Service) GetUnusedEBSVolumes(ctx context.Context) ([]types.Volume, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]types.Volume), args.Error(1)
}

// GetStoppedInstancesInfo mocks the GetStoppedInstancesInfo method.
func (m *MockEC2Service) GetStoppedInstancesInfo(ctx context.Context, stoppedDays int) ([]types.Instance, []types.Volume, error) {
	args := m.Called(ctx, stoppedDays)

	var (
		instances []types.Instance
		volumes   []types.Volume
	)

	if args.Get(0) != nil {
		instances = args.Get(0).([]types.Instance)
	}

	if args.Get(1) != nil {
		volumes = args.Get(1).([]types.Volume)
	}

	return instances, volumes, args.Error(2)
}

// GetReservedInstanceExpiringOrExpiredWaste mocks the GetReservedInstanceExpiringOrExpiredWaste method.
func (m *MockEC2Service) GetReservedInstanceExpiringOrExpiredWaste(ctx context.Context, expiringDays int) ([]model.RiExpirationInfo, error) {
	args := m.Called(ctx, expiringDays)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]model.RiExpirationInfo), args.Error(1)
}

// GetUnusedAMIs mocks the GetUnusedAMIs method.
func (m *MockEC2Service) GetUnusedAMIs(ctx context.Context, staleDays int) ([]model.AMIWasteInfo, error) {
	args := m.Called(ctx, staleDays)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]model.AMIWasteInfo), args.Error(1)
}

// GetOrphanedSnapshots mocks the GetOrphanedSnapshots method.
func (m *MockEC2Service) GetOrphanedSnapshots(ctx context.Context, staleDays int) ([]model.SnapshotWasteInfo, error) {
	args := m.Called(ctx, staleDays)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]model.SnapshotWasteInfo), args.Error(1)
}

// GetUnusedKeyPairs mocks the GetUnusedKeyPairs method.
func (m *MockEC2Service) GetUnusedKeyPairs(ctx context.Context) ([]model.KeyPairWasteInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]model.KeyPairWasteInfo), args.Error(1)
}

// GetIdleInstances mocks the GetIdleInstances method.
func (m *MockEC2Service) GetIdleInstances(ctx context.Context, idleDays int, cpuThresholdPercent float64, networkBytesPerDayThreshold float64) ([]model.EC2IdleInstanceInfo, error) {
	args := m.Called(ctx, idleDays, cpuThresholdPercent, networkBytesPerDayThreshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]model.EC2IdleInstanceInfo), args.Error(1)
}

// GetPublicIPv4Summary mocks the GetPublicIPv4Summary method.
func (m *MockEC2Service) GetPublicIPv4Summary(ctx context.Context) (*model.PublicIPv4Summary, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.PublicIPv4Summary), args.Error(1)
}
