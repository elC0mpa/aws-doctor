package mocks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/mock"
)

// MockSTSService is a mock implementation of the STS service interface.
// MockSTSService is a mock implementation of the STS service.
//
//nolint:revive
type MockSTSService struct {
	mock.Mock
}

// GetCallerIdentity mocks the GetCallerIdentity method.
func (m *MockSTSService) GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*sts.GetCallerIdentityOutput), args.Error(1)
}
