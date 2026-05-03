package services

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockSecretsManagerService is a mock of the Secrets Manager service.
type MockSecretsManagerService struct {
	mock.Mock
}

// GetUnusedSecrets mocks the GetUnusedSecrets method.
func (m *MockSecretsManagerService) GetUnusedSecrets(ctx context.Context, idleDays int) ([]model.UnusedSecretInfo, error) {
	args := m.Called(ctx, idleDays)
	return args.Get(0).([]model.UnusedSecretInfo), args.Error(1)
}
