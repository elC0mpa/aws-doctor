package services

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockUpdateService is a mock implementation of the update Service interface.
type MockUpdateService struct {
	mock.Mock
}

// Update mocks the Update method.
func (m *MockUpdateService) Update() error {
	args := m.Called()
	return args.Error(0)
}

// CheckForUpdate mocks the CheckForUpdate method.
func (m *MockUpdateService) CheckForUpdate(ctx context.Context) (*string, error) {
	args := m.Called(ctx)

	var s *string
	if args.Get(0) != nil {
		s = args.Get(0).(*string)
	}

	return s, args.Error(1)
}
