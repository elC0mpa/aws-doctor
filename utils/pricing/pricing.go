// Package pricing provides AWS pricing constants for cost estimation.
// These are approximate us-east-1 prices and may vary by region.
package pricing

const (
	// EIPCostPerMonth is the cost of an unassociated Elastic IP (~$0.005/hour * 730 hours)
	EIPCostPerMonth = 3.65

	// EBSgp2CostPerGBMonth is the cost of gp2 EBS storage per GB per month
	EBSgp2CostPerGBMonth = 0.10

	// ALBCostPerMonth is the base cost of an ALB/NLB (~$0.0225/hour * 730 hours)
	ALBCostPerMonth = 16.43

	// CLBCostPerMonth is the base cost of a CLB (~$0.025/hour * 730 hours)
	CLBCostPerMonth = 18.25
)
