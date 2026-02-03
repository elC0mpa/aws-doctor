package awsinterfaces

import (
	"context"

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/stretchr/testify/mock"
)

// MockELBClient is a mock of ClientAPI
type MockELBClient struct {
	mock.Mock
}

// DescribeLoadBalancers mocks the DescribeLoadBalancers API call.
func (m *MockELBClient) DescribeLoadBalancers(ctx context.Context, params *elb.DescribeLoadBalancersInput, optFns ...func(*elb.Options)) (*elb.DescribeLoadBalancersOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*elb.DescribeLoadBalancersOutput), args.Error(1)
}

// DescribeTargetGroups mocks the DescribeTargetGroups API call.
func (m *MockELBClient) DescribeTargetGroups(ctx context.Context, params *elb.DescribeTargetGroupsInput, optFns ...func(*elb.Options)) (*elb.DescribeTargetGroupsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*elb.DescribeTargetGroupsOutput), args.Error(1)
}
