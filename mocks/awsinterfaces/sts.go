package awsinterfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/mock"
)

// MockSTSClient is a mock of STSClientAPI
type MockSTSClient struct {
	mock.Mock
}

// GetCallerIdentity mocks the GetCallerIdentity API call.
func (m *MockSTSClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*sts.GetCallerIdentityOutput), args.Error(1)
}
