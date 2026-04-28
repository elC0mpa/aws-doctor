package services

import (
	"context"

	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockELBService is a mock implementation of the ELB service interface.
type MockELBService struct {
	mock.Mock
}

// GetLoadBalancerWaste mocks the GetLoadBalancerWaste method.
func (m *MockELBService) GetLoadBalancerWaste(ctx context.Context, idleDays int) ([]elbtypes.LoadBalancer, []model.ELBIdleInfo, error) {
	args := m.Called(ctx, idleDays)

	var unused []elbtypes.LoadBalancer

	if args.Get(0) != nil {
		unused = args.Get(0).([]elbtypes.LoadBalancer)
	}

	var idle []model.ELBIdleInfo

	if args.Get(1) != nil {
		idle = args.Get(1).([]model.ELBIdleInfo)
	}

	return unused, idle, args.Error(2)
}
