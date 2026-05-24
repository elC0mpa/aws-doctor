package awsinterfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/stretchr/testify/mock"
)

// MockIAMClientAPI is a mock for IAMClientAPI
type MockIAMClientAPI struct {
	mock.Mock
}

// ListUsers provides a mock function with given fields: ctx, params, optFns
func (m *MockIAMClientAPI) ListUsers(ctx context.Context, params *iam.ListUsersInput, optFns ...func(*iam.Options)) (*iam.ListUsersOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*iam.ListUsersOutput), args.Error(1)
	}

	return nil, args.Error(1)
}

// ListAccessKeys provides a mock function with given fields: ctx, params, optFns
func (m *MockIAMClientAPI) ListAccessKeys(ctx context.Context, params *iam.ListAccessKeysInput, optFns ...func(*iam.Options)) (*iam.ListAccessKeysOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*iam.ListAccessKeysOutput), args.Error(1)
	}

	return nil, args.Error(1)
}

// GetAccessKeyLastUsed provides a mock function with given fields: ctx, params, optFns
func (m *MockIAMClientAPI) GetAccessKeyLastUsed(ctx context.Context, params *iam.GetAccessKeyLastUsedInput, optFns ...func(*iam.Options)) (*iam.GetAccessKeyLastUsedOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*iam.GetAccessKeyLastUsedOutput), args.Error(1)
	}

	return nil, args.Error(1)
}

// GetAccountSummary provides a mock function with given fields: ctx, params, optFns
func (m *MockIAMClientAPI) GetAccountSummary(ctx context.Context, params *iam.GetAccountSummaryInput, optFns ...func(*iam.Options)) (*iam.GetAccountSummaryOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*iam.GetAccountSummaryOutput), args.Error(1)
	}

	return nil, args.Error(1)
}
