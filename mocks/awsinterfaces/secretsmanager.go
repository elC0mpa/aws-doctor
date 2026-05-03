package awsinterfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/mock"
)

// MockSecretsManagerClient is a mock of the Secrets Manager ClientAPI.
type MockSecretsManagerClient struct {
	mock.Mock
}

// ListSecrets mocks the ListSecrets API call.
func (m *MockSecretsManagerClient) ListSecrets(ctx context.Context, params *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*secretsmanager.ListSecretsOutput), args.Error(1)
}
