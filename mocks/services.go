// Package mocks provides mock implementations of service interfaces for testing.
package mocks

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// STSService mock

type MockSTSService struct {
	mock.Mock
}

func (m *MockSTSService) GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sts.GetCallerIdentityOutput), args.Error(1)
}

// CostService mock

type MockCostService struct {
	mock.Mock
}

func (m *MockCostService) GetCurrentMonthCostsByService(ctx context.Context) (*model.CostInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CostInfo), args.Error(1)
}

func (m *MockCostService) GetLastMonthCostsByService(ctx context.Context) (*model.CostInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CostInfo), args.Error(1)
}

func (m *MockCostService) GetMonthCostsByService(ctx context.Context, endDate time.Time) (*model.CostInfo, error) {
	args := m.Called(ctx, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CostInfo), args.Error(1)
}

func (m *MockCostService) GetCurrentMonthTotalCosts(ctx context.Context) (*string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockCostService) GetLastMonthTotalCosts(ctx context.Context) (*string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockCostService) GetLastSixMonthsCosts(ctx context.Context) ([]model.CostInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.CostInfo), args.Error(1)
}

// EC2Service mock

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

// ELBService mock

type MockELBService struct {
	mock.Mock
}

func (m *MockELBService) GetUnusedLoadBalancers(ctx context.Context) ([]elbtypes.LoadBalancer, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]elbtypes.LoadBalancer), args.Error(1)
}
