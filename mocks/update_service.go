package mocks

import (
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
