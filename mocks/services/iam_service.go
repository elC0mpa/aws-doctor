package services

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockIAMService is a mock for IAM Service
type MockIAMService struct {
	mock.Mock
}

// GetIAMWaste provides a mock function with given fields: ctx, idleDays
func (m *MockIAMService) GetIAMWaste(ctx context.Context, idleDays int) ([]model.IAMUserWasteInfo, []model.IAMRootUserWasteInfo, error) {
	args := m.Called(ctx, idleDays)

	var users []model.IAMUserWasteInfo
	if args.Get(0) != nil {
		users = args.Get(0).([]model.IAMUserWasteInfo)
	}

	var root []model.IAMRootUserWasteInfo
	if args.Get(1) != nil {
		root = args.Get(1).([]model.IAMRootUserWasteInfo)
	}

	return users, root, args.Error(2)
}
