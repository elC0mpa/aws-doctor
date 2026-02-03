package output

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils"
)

// Format represents the output format type
type Format string

// FormatTable represents the table output format.
const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// Renderer defines the interface for drawing tables and charts
type Renderer interface {
	DrawCostTable(accountID, lastTotalCost, currentTotalCost string, lastMonth, currentMonth *model.CostInfo, costsAggregation string)
	OutputCostComparisonJSON(accountID string, lastTotalCost, currentTotalCost float64, lastMonth, currentMonth *model.CostInfo) error
	DrawTrendChart(accountID string, costInfo []model.CostInfo)
	OutputTrendJSON(accountID string, costInfo []model.CostInfo) error
	DrawWasteTable(accountID string, elasticIPs []types.Address, unusedVolumes []types.Volume, stoppedVolumes []types.Volume, ris []model.RiExpirationInfo, stoppedInstances []types.Instance, loadBalancers []elbtypes.LoadBalancer, unusedAMIs []model.AMIWasteInfo, orphanedSnapshots []model.SnapshotWasteInfo)
	OutputWasteJSON(accountID string, elasticIPs []types.Address, unusedVolumes []types.Volume, stoppedVolumes []types.Volume, ris []model.RiExpirationInfo, stoppedInstances []types.Instance, loadBalancers []elbtypes.LoadBalancer, unusedAMIs []model.AMIWasteInfo, orphanedSnapshots []model.SnapshotWasteInfo) error
	StopSpinner()
}

type realRenderer struct{}

func (r *realRenderer) DrawCostTable(accountID, lastTotalCost, currentTotalCost string, lastMonth, currentMonth *model.CostInfo, costsAggregation string) {
	utils.DrawCostTable(accountID, lastTotalCost, currentTotalCost, lastMonth, currentMonth, costsAggregation)
}

func (r *realRenderer) OutputCostComparisonJSON(accountID string, lastTotalCost, currentTotalCost float64, lastMonth, currentMonth *model.CostInfo) error {
	return utils.OutputCostComparisonJSON(accountID, lastTotalCost, currentTotalCost, lastMonth, currentMonth)
}

func (r *realRenderer) DrawTrendChart(accountID string, costInfo []model.CostInfo) {
	utils.DrawTrendChart(accountID, costInfo)
}

func (r *realRenderer) OutputTrendJSON(accountID string, costInfo []model.CostInfo) error {
	return utils.OutputTrendJSON(accountID, costInfo)
}

func (r *realRenderer) DrawWasteTable(accountID string, elasticIPs []types.Address, unusedVolumes []types.Volume, stoppedVolumes []types.Volume, ris []model.RiExpirationInfo, stoppedInstances []types.Instance, loadBalancers []elbtypes.LoadBalancer, unusedAMIs []model.AMIWasteInfo, orphanedSnapshots []model.SnapshotWasteInfo) {
	utils.DrawWasteTable(accountID, elasticIPs, unusedVolumes, stoppedVolumes, ris, stoppedInstances, loadBalancers, unusedAMIs, orphanedSnapshots)
}

func (r *realRenderer) OutputWasteJSON(accountID string, elasticIPs []types.Address, unusedVolumes []types.Volume, stoppedVolumes []types.Volume, ris []model.RiExpirationInfo, stoppedInstances []types.Instance, loadBalancers []elbtypes.LoadBalancer, unusedAMIs []model.AMIWasteInfo, orphanedSnapshots []model.SnapshotWasteInfo) error {
	return utils.OutputWasteJSON(accountID, elasticIPs, unusedVolumes, stoppedVolumes, ris, stoppedInstances, loadBalancers, unusedAMIs, orphanedSnapshots)
}

func (r *realRenderer) StopSpinner() {
	utils.StopSpinner()
}

// service is the internal implementation
type service struct {
	format   Format
	renderer Renderer
}

// Service defines the interface for output operations
type Service interface {
	// RenderCostComparison outputs cost comparison data in the configured format
	RenderCostComparison(accountID, lastTotalCost, currentTotalCost string, lastMonth, currentMonth *model.CostInfo) error

	// RenderTrend outputs trend data in the configured format
	RenderTrend(accountID string, costInfo []model.CostInfo) error

	// RenderWaste outputs waste report data in the configured format
	RenderWaste(accountID string, elasticIPs []types.Address, unusedVolumes []types.Volume, stoppedVolumes []types.Volume, ris []model.RiExpirationInfo, stoppedInstances []types.Instance, loadBalancers []elbtypes.LoadBalancer, unusedAMIs []model.AMIWasteInfo, orphanedSnapshots []model.SnapshotWasteInfo) error

	// StopSpinner stops the loading spinner before rendering output
	StopSpinner()
}
