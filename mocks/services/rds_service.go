package services

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockRDSService is a mock of RDS Service
type MockRDSService struct {
	mock.Mock
}

// GetRDSWaste mocks the GetRDSWaste method.
func (m *MockRDSService) GetRDSWaste(ctx context.Context) ([]model.RDSInstanceWasteInfo, []model.RDSSnapshotWasteInfo, []model.RDSIdleInstanceInfo, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, nil, nil, args.Error(3)
	}

	return args.Get(0).([]model.RDSInstanceWasteInfo), args.Get(1).([]model.RDSSnapshotWasteInfo), args.Get(2).([]model.RDSIdleInstanceInfo), args.Error(3)
}
