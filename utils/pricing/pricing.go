// Package pricing provides AWS pricing estimates. At startup, Load fetches region-aware rates
// from the AWS Pricing API and caches them; the Calculate* helpers prefer cached rates and fall
// back to the hardcoded us-east-1 defaults below when an entry is missing.
package pricing

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// hoursPerMonth converts Pricing API hourly rates into an approximate monthly cost. 730 is the
// same figure AWS uses in Cost Explorer examples (365.25 * 24 / 12).
const hoursPerMonth = 730.0

const (
	// EIPCostPerMonth is the default cost of an unassociated Elastic IP (~$0.005/hour * 730 hours).
	EIPCostPerMonth = 3.65

	// EBSgp2CostPerGBMonth is the default cost of gp2 EBS storage per GB per month.
	EBSgp2CostPerGBMonth = 0.10
	// EBSgp3CostPerGBMonth is the default cost of gp3 EBS storage per GB per month.
	EBSgp3CostPerGBMonth = 0.08
	// EBSio1CostPerGBMonth is the default cost of io1 EBS storage per GB per month.
	EBSio1CostPerGBMonth = 0.125
	// EBSio2CostPerGBMonth is the default cost of io2 EBS storage per GB per month.
	EBSio2CostPerGBMonth = 0.125
	// EBSst1CostPerGBMonth is the default cost of st1 EBS storage per GB per month.
	EBSst1CostPerGBMonth = 0.045
	// EBSsc1CostPerGBMonth is the default cost of sc1 EBS storage per GB per month.
	EBSsc1CostPerGBMonth = 0.015

	// CloudWatchLogsCostPerGBMonth is the default cost of CloudWatch Logs storage per GB per month.
	CloudWatchLogsCostPerGBMonth = 0.03

	// NatGatewayCostPerMonth is the default cost of a NAT Gateway (~$0.045/hour * 730 hours in us-east-1).
	NatGatewayCostPerMonth = 32.85

	// ALBCostPerMonth is the default base cost of an ALB/NLB (~$0.0225/hour * 730 hours).
	ALBCostPerMonth = 16.43

	// CLBCostPerMonth is the default base cost of a CLB (~$0.025/hour * 730 hours).
	CLBCostPerMonth = 18.25
)

// ebsSpec maps each supported EBS volume type to its Pricing API variant name and default rate.
//
//nolint:gochecknoglobals // lookup table
var ebsSpec = map[types.VolumeType]struct {
	variant     string
	defaultRate float64
}{
	types.VolumeTypeGp2: {"gp2", EBSgp2CostPerGBMonth},
	types.VolumeTypeGp3: {"gp3", EBSgp3CostPerGBMonth},
	types.VolumeTypeIo1: {"io1", EBSio1CostPerGBMonth},
	types.VolumeTypeIo2: {"io2", EBSio2CostPerGBMonth},
	types.VolumeTypeSt1: {"st1", EBSst1CostPerGBMonth},
	types.VolumeTypeSc1: {"sc1", EBSsc1CostPerGBMonth},
}

// EBSCostPerGBMonth returns the per-GB monthly cost for a given EBS volume type. Unknown types
// fall back to gp2 pricing, matching the prior behavior.
func EBSCostPerGBMonth(volumeType types.VolumeType) float64 {
	spec, ok := ebsSpec[volumeType]
	if !ok {
		spec = ebsSpec[types.VolumeTypeGp2]
	}

	if v, ok := lookupMonthly(priceKey(categoryEBS, spec.variant), 0); ok {
		return v
	}

	return spec.defaultRate
}

// CalculateEBSMonthlyCost calculates the estimated monthly cost for an EBS volume.
func CalculateEBSMonthlyCost(sizeGiB int32, volumeType types.VolumeType) float64 {
	return float64(sizeGiB) * EBSCostPerGBMonth(volumeType)
}

// CalculateEIPMonthlyCost returns the estimated monthly cost for an unassociated Elastic IP.
func CalculateEIPMonthlyCost() float64 {
	if v, ok := lookupMonthly(priceKey(categoryEIP, ""), hoursPerMonth); ok {
		return v
	}

	return EIPCostPerMonth
}

// CalculateLoadBalancerMonthlyCost calculates the estimated monthly cost for a load balancer.
func CalculateLoadBalancerMonthlyCost(lbType elbtypes.LoadBalancerTypeEnum) float64 {
	if lbType == "classic" {
		if v, ok := lookupMonthly(priceKey(categoryLBClassic, ""), hoursPerMonth); ok {
			return v
		}

		return CLBCostPerMonth
	}

	if v, ok := lookupMonthly(priceKey(categoryLBApp, string(lbType)), hoursPerMonth); ok {
		return v
	}

	return ALBCostPerMonth
}

// CalculateCloudWatchLogsMonthlyCost calculates the estimated monthly storage cost for CloudWatch Logs.
func CalculateCloudWatchLogsMonthlyCost(storedBytes int64) float64 {
	storedGB := float64(storedBytes) / (1024 * 1024 * 1024)

	rate := CloudWatchLogsCostPerGBMonth
	if v, ok := lookupMonthly(priceKey(categoryCWLogs, ""), 0); ok {
		rate = v
	}

	return storedGB * rate
}

// CalculateNATGatewayMonthlyCost returns the estimated monthly cost for a NAT Gateway.
func CalculateNATGatewayMonthlyCost() float64 {
	if v, ok := lookupMonthly(priceKey(categoryNAT, ""), hoursPerMonth); ok {
		return v
	}

	return NatGatewayCostPerMonth
}
