package renderers

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockRenderer is a mock implementation of output.Renderer
type MockRenderer struct {
	mock.Mock
}

// DrawCostTable mocks DrawCostTable
func (m *MockRenderer) DrawCostTable(accountID, lastTotalCost, currentTotalCost string, lastMonth, currentMonth *model.CostInfo, costsAggregation string) {
	m.Called(accountID, lastTotalCost, currentTotalCost, lastMonth, currentMonth, costsAggregation)
}

// OutputCostComparisonJSON mocks OutputCostComparisonJSON
func (m *MockRenderer) OutputCostComparisonJSON(accountID string, lastTotalCost, currentTotalCost float64, lastMonth, currentMonth *model.CostInfo) error {
	args := m.Called(accountID, lastTotalCost, currentTotalCost, lastMonth, currentMonth)
	return args.Error(0)
}

// DrawTrendChart mocks DrawTrendChart
func (m *MockRenderer) DrawTrendChart(accountID string, costInfo []model.CostInfo) {
	m.Called(accountID, costInfo)
}

// OutputTrendJSON mocks OutputTrendJSON
func (m *MockRenderer) OutputTrendJSON(accountID string, costInfo []model.CostInfo) error {
	args := m.Called(accountID, costInfo)
	return args.Error(0)
}

// DrawWasteTable mocks DrawWasteTable
func (m *MockRenderer) DrawWasteTable(accountID string, elasticIPs []types.Address, unusedVolumes []types.Volume, stoppedVolumes []types.Volume, ris []model.RiExpirationInfo, stoppedInstances []types.Instance, loadBalancers []elbtypes.LoadBalancer, unusedAMIs []model.AMIWasteInfo, orphanedSnapshots []model.SnapshotWasteInfo, unusedKeyPairs []model.KeyPairWasteInfo) {
	m.Called(accountID, elasticIPs, unusedVolumes, stoppedVolumes, ris, stoppedInstances, loadBalancers, unusedAMIs, orphanedSnapshots, unusedKeyPairs)
}

// OutputWasteJSON mocks OutputWasteJSON
func (m *MockRenderer) OutputWasteJSON(accountID string, elasticIPs []types.Address, unusedVolumes []types.Volume, stoppedVolumes []types.Volume, ris []model.RiExpirationInfo, stoppedInstances []types.Instance, loadBalancers []elbtypes.LoadBalancer, unusedAMIs []model.AMIWasteInfo, orphanedSnapshots []model.SnapshotWasteInfo, unusedKeyPairs []model.KeyPairWasteInfo) error {
	args := m.Called(accountID, elasticIPs, unusedVolumes, stoppedVolumes, ris, stoppedInstances, loadBalancers, unusedAMIs, orphanedSnapshots, unusedKeyPairs)
	return args.Error(0)
}

// StopSpinner mocks StopSpinner
func (m *MockRenderer) StopSpinner() {
	m.Called()
}
