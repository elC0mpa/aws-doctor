package awsinterfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/stretchr/testify/mock"
)

// MockECRClient is a mock of ECR ClientAPI
type MockECRClient struct {
	mock.Mock
}

// DescribeRepositories mocks the DescribeRepositories method.
func (m *MockECRClient) DescribeRepositories(ctx context.Context, params *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ecr.DescribeRepositoriesOutput), args.Error(1)
}

// DescribeImages mocks the DescribeImages method.
func (m *MockECRClient) DescribeImages(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ecr.DescribeImagesOutput), args.Error(1)
}

// GetLifecyclePolicy mocks the GetLifecyclePolicy method.
func (m *MockECRClient) GetLifecyclePolicy(ctx context.Context, params *ecr.GetLifecyclePolicyInput, optFns ...func(*ecr.Options)) (*ecr.GetLifecyclePolicyOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ecr.GetLifecyclePolicyOutput), args.Error(1)
}
