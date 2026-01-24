package mocks

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockOutputService is a mock implementation of the output service interface.
type MockOutputService struct {
	mock.Mock
}

func (m *MockOutputService) RenderCostComparison(accountID, lastTotalCost, currentTotalCost string, lastMonth, currentMonth *model.CostInfo) error {
	args := m.Called(accountID, lastTotalCost, currentTotalCost, lastMonth, currentMonth)
	return args.Error(0)
}

func (m *MockOutputService) RenderTrend(accountID string, costInfo []model.CostInfo) error {
	args := m.Called(accountID, costInfo)
	return args.Error(0)
}

func (m *MockOutputService) RenderWaste(accountID string, elasticIPs []types.Address, unusedVolumes []types.Volume, stoppedVolumes []types.Volume, ris []model.RiExpirationInfo, stoppedInstances []types.Instance, loadBalancers []elbtypes.LoadBalancer, unusedAMIs []model.AMIWasteInfo, orphanedSnapshots []model.SnapshotWasteInfo) error {
	args := m.Called(accountID, elasticIPs, unusedVolumes, stoppedVolumes, ris, stoppedInstances, loadBalancers, unusedAMIs, orphanedSnapshots)
	return args.Error(0)
}

func (m *MockOutputService) StopSpinner() {
	m.Called()
}
