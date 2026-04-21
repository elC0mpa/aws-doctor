package awsinterfaces

import (
	"context"

	sm "github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/stretchr/testify/mock"
)

// MockSageMakerClient is a mock of the SageMaker ClientAPI.
type MockSageMakerClient struct {
	mock.Mock
}

// ListEndpoints mocks the ListEndpoints API call.
func (m *MockSageMakerClient) ListEndpoints(ctx context.Context, params *sm.ListEndpointsInput, optFns ...func(*sm.Options)) (*sm.ListEndpointsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*sm.ListEndpointsOutput), args.Error(1)
}

// DescribeEndpoint mocks the DescribeEndpoint API call.
func (m *MockSageMakerClient) DescribeEndpoint(ctx context.Context, params *sm.DescribeEndpointInput, optFns ...func(*sm.Options)) (*sm.DescribeEndpointOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*sm.DescribeEndpointOutput), args.Error(1)
}

// DescribeEndpointConfig mocks the DescribeEndpointConfig API call.
func (m *MockSageMakerClient) DescribeEndpointConfig(ctx context.Context, params *sm.DescribeEndpointConfigInput, optFns ...func(*sm.Options)) (*sm.DescribeEndpointConfigOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*sm.DescribeEndpointConfigOutput), args.Error(1)
}
