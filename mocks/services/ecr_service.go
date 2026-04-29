package services

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockECRService is a mock of ECR Service
type MockECRService struct {
	mock.Mock
}

// GetECRWaste mocks the GetECRWaste method.
func (m *MockECRService) GetECRWaste(ctx context.Context) ([]model.ECRNoLifecyclePolicyInfo, []model.ECREmptyRepositoryInfo, []model.ECRUntaggedImageInfo, error) {
	args := m.Called(ctx)

	noPolicy, _ := args.Get(0).([]model.ECRNoLifecyclePolicyInfo)
	empty, _ := args.Get(1).([]model.ECREmptyRepositoryInfo)
	untagged, _ := args.Get(2).([]model.ECRUntaggedImageInfo)

	return noPolicy, empty, untagged, args.Error(3)
}
