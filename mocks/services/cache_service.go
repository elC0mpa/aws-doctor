package services

import (
	"github.com/elC0mpa/aws-doctor/service/cache"
	"github.com/stretchr/testify/mock"
)

// MockCacheService is a mock implementation of the cache Service interface.
type MockCacheService struct {
	mock.Mock
}

// Get mocks the Get method.
func (m *MockCacheService) Get(key cache.Key, target interface{}, contexts ...string) (bool, error) {
	args := m.Called(key, target, contexts)

	return args.Bool(0), args.Error(1)
}

// Set mocks the Set method.
func (m *MockCacheService) Set(key cache.Key, value interface{}, contexts ...string) error {
	args := m.Called(key, value, contexts)

	return args.Error(0)
}
