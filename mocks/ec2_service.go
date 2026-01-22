package mocks

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

func (m *MockEC2Service) GetElasticIpAddressesInfo(ctx context.Context) (*model.ElasticIpInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ElasticIpInfo), args.Error(1)
}

func (m *MockEC2Service) GetUnusedElasticIpAddressesInfo(ctx context.Context) ([]types.Address, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.Address), args.Error(1)
}

func (m *MockEC2Service) GetUnusedEBSVolumes(ctx context.Context) ([]types.Volume, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.Volume), args.Error(1)
}

func (m *MockEC2Service) GetStoppedInstancesInfo(ctx context.Context) ([]types.Instance, []types.Volume, error) {
	args := m.Called(ctx)
	var instances []types.Instance
	var volumes []types.Volume
	if args.Get(0) != nil {
		instances = args.Get(0).([]types.Instance)
	}
	if args.Get(1) != nil {
		volumes = args.Get(1).([]types.Volume)
	}
	return instances, volumes, args.Error(2)
}

func (m *MockEC2Service) GetReservedInstanceExpiringOrExpired30DaysWaste(ctx context.Context) ([]model.RiExpirationInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.RiExpirationInfo), args.Error(1)
}

func (m *MockEC2Service) GetUnusedAMIs(ctx context.Context, staleDays int) ([]model.AMIWasteInfo, error) {
	args := m.Called(ctx, staleDays)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.AMIWasteInfo), args.Error(1)
}
