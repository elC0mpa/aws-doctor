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

// GetS3Waste mocks the GetS3Waste method.
func (m *MockS3Service) GetS3Waste(ctx context.Context) ([]model.S3BucketWasteInfo, []model.S3MultipartUploadWasteInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil && args.Get(1) == nil {
		return nil, nil, args.Error(2)
	}

	return args.Get(0).([]model.S3BucketWasteInfo), args.Get(1).([]model.S3MultipartUploadWasteInfo), args.Error(2)
}