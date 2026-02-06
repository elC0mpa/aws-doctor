package services

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockS3Service is a mock of S3 Service
type MockS3Service struct {
	mock.Mock
}

// GetBucketsWithoutLifecyclePolicies mocks the GetBucketsWithoutLifecyclePolicies method.
func (m *MockS3Service) GetBucketsWithoutLifecyclePolicies(ctx context.Context) ([]model.S3BucketWasteInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]model.S3BucketWasteInfo), args.Error(1)
}
