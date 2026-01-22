package output

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils"
)

// NewService creates a new output service with the specified format
func NewService(format string) Service {
	f := FormatTable
	if format == "json" {
		f = FormatJSON
	}
	return &service{format: f}
}

func (s *service) RenderCostComparison(accountID, lastTotalCost, currentTotalCost string, lastMonth, currentMonth *model.CostInfo) error {
	if s.format == FormatJSON {
		return utils.OutputCostComparisonJSON(
			accountID,
			utils.ParseCostString(lastTotalCost),
			utils.ParseCostString(currentTotalCost),
			lastMonth,
			currentMonth,
		)
	}

	utils.DrawCostTable(accountID, lastTotalCost, currentTotalCost, lastMonth, currentMonth, "UnblendedCost")
	return nil
}

func (s *service) RenderTrend(accountID string, costInfo []model.CostInfo) error {
	if s.format == FormatJSON {
		return utils.OutputTrendJSON(accountID, costInfo)
	}

	utils.DrawTrendChart(accountID, costInfo)
	return nil
}

func (s *service) RenderWaste(accountID string, elasticIPs []types.Address, unusedVolumes []types.Volume, stoppedVolumes []types.Volume, ris []model.RiExpirationInfo, stoppedInstances []types.Instance, loadBalancers []elbtypes.LoadBalancer, unusedAMIs []model.AMIWasteInfo, orphanedSnapshots []model.SnapshotWasteInfo) error {
	if s.format == FormatJSON {
		return utils.OutputWasteJSON(accountID, elasticIPs, unusedVolumes, stoppedVolumes, ris, stoppedInstances, loadBalancers, unusedAMIs, orphanedSnapshots)
	}

	utils.DrawWasteTable(accountID, elasticIPs, unusedVolumes, stoppedVolumes, ris, stoppedInstances, loadBalancers, unusedAMIs, orphanedSnapshots)
	return nil
}

func (s *service) StopSpinner() {
	utils.StopSpinner()
}
