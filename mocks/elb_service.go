package mocks

import (
	"context"

	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/stretchr/testify/mock"
)

// MockELBService is a mock implementation of the ELB service interface.
type MockELBService struct {
	mock.Mock
}

// GetUnusedLoadBalancers mocks the GetUnusedLoadBalancers method.
func (m *MockELBService) GetUnusedLoadBalancers(ctx context.Context) ([]elbtypes.LoadBalancer, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]elbtypes.LoadBalancer), args.Error(1)
}
