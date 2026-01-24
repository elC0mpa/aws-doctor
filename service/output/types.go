package output

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
)

// Format represents the output format type
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// service is the internal implementation
type service struct {
	format Format
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
