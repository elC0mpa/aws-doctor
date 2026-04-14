// Package pricing provides AWS pricing constants for cost estimation.
// These are approximate us-east-1 prices and may vary by region.
package pricing

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

const (
	// EIPCostPerMonth is the cost of an unassociated Elastic IP (~$0.005/hour * 730 hours)
	EIPCostPerMonth = 3.65

	// EBSgp2CostPerGBMonth is the cost of gp2 EBS storage per GB per month.
	EBSgp2CostPerGBMonth = 0.10
	// EBSgp3CostPerGBMonth is the cost of gp3 EBS storage per GB per month.
	EBSgp3CostPerGBMonth = 0.08
	// EBSio1CostPerGBMonth is the cost of io1 EBS storage per GB per month.
	EBSio1CostPerGBMonth = 0.125
	// EBSio2CostPerGBMonth is the cost of io2 EBS storage per GB per month.
	EBSio2CostPerGBMonth = 0.125
	// EBSst1CostPerGBMonth is the cost of st1 EBS storage per GB per month.
	EBSst1CostPerGBMonth = 0.045
	// EBSsc1CostPerGBMonth is the cost of sc1 EBS storage per GB per month.
	EBSsc1CostPerGBMonth = 0.015

	// CloudWatchLogsCostPerGBMonth is the cost of CloudWatch Logs storage per GB per month.
	CloudWatchLogsCostPerGBMonth = 0.03

	// NatGatewayCostPerMonth is the cost of a NAT Gateway (~$0.045/hour * 730 hours in us-east-1).
	NatGatewayCostPerMonth = 32.85

	// ALBCostPerMonth is the base cost of an ALB/NLB (~$0.0225/hour * 730 hours)
	ALBCostPerMonth = 16.43

	// CLBCostPerMonth is the base cost of a CLB (~$0.025/hour * 730 hours)
	CLBCostPerMonth = 18.25
)

// EBSCostPerGBMonth returns the per-GB monthly cost for a given EBS volume type.
func EBSCostPerGBMonth(volumeType types.VolumeType) float64 {
	switch volumeType {
	case types.VolumeTypeGp3:
		return EBSgp3CostPerGBMonth
	case types.VolumeTypeIo1:
		return EBSio1CostPerGBMonth
	case types.VolumeTypeIo2:
		return EBSio2CostPerGBMonth
	case types.VolumeTypeSt1:
		return EBSst1CostPerGBMonth
	case types.VolumeTypeSc1:
		return EBSsc1CostPerGBMonth
	default:
		return EBSgp2CostPerGBMonth
	}
}

// CalculateEBSMonthlyCost calculates the estimated monthly cost for an EBS volume.
func CalculateEBSMonthlyCost(sizeGiB int32, volumeType types.VolumeType) float64 {
	return float64(sizeGiB) * EBSCostPerGBMonth(volumeType)
}

// CalculateEIPMonthlyCost returns the estimated monthly cost for an unassociated Elastic IP.
func CalculateEIPMonthlyCost() float64 {
	return EIPCostPerMonth
}

// CalculateLoadBalancerMonthlyCost calculates the estimated monthly cost for a load balancer.
func CalculateLoadBalancerMonthlyCost(lbType elbtypes.LoadBalancerTypeEnum) float64 {
	if lbType == "classic" {
		return CLBCostPerMonth
	}

	return ALBCostPerMonth
}

// CalculateCloudWatchLogsMonthlyCost calculates the estimated monthly storage cost for CloudWatch Logs.
func CalculateCloudWatchLogsMonthlyCost(storedBytes int64) float64 {
	storedGB := float64(storedBytes) / (1024 * 1024 * 1024)
	return storedGB * CloudWatchLogsCostPerGBMonth
}

// CalculateNATGatewayMonthlyCost returns the estimated monthly cost for a NAT Gateway.
func CalculateNATGatewayMonthlyCost() float64 {
	return NatGatewayCostPerMonth
}
