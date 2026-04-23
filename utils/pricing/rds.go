package pricing

import "strings"

const (
	// RDSStorageCostPerGBMonth is the default cost of RDS gp2 storage per GB per month.
	RDSStorageCostPerGBMonth = 0.115
	// RDSSnapshotCostPerGBMonth is the default cost of RDS snapshot storage per GB per month.
	RDSSnapshotCostPerGBMonth = 0.095
)

// rdsInstancePricing maps RDS instance classes to approximate monthly compute costs (us-east-1,
// Single-AZ, On-Demand MySQL). Used as a fallback when the AWS Pricing API lookup is unavailable
// or returns no entry for the instance class.
//
//nolint:gochecknoglobals // pricing lookup table
var rdsInstancePricing = map[string]float64{
	// T3 family
	"db.t3.micro":  12.41,
	"db.t3.small":  24.82,
	"db.t3.medium": 49.64,
	"db.t3.large":  99.28,
	"db.t3.xlarge": 198.56,
	// T4g family (Graviton)
	"db.t4g.micro":  11.52,
	"db.t4g.small":  23.04,
	"db.t4g.medium": 46.08,
	"db.t4g.large":  92.16,
	"db.t4g.xlarge": 184.32,
	// M5 family
	"db.m5.large":    124.10,
	"db.m5.xlarge":   248.20,
	"db.m5.2xlarge":  496.40,
	"db.m5.4xlarge":  992.80,
	"db.m5.8xlarge":  1985.60,
	"db.m5.12xlarge": 2978.40,
	"db.m5.16xlarge": 3971.20,
	"db.m5.24xlarge": 5956.80,
	// M6g family (Graviton)
	"db.m6g.large":    118.26,
	"db.m6g.xlarge":   236.52,
	"db.m6g.2xlarge":  473.04,
	"db.m6g.4xlarge":  946.08,
	"db.m6g.8xlarge":  1892.16,
	"db.m6g.12xlarge": 2838.24,
	"db.m6g.16xlarge": 3784.32,
	// M7g family (Graviton3)
	"db.m7g.large":    127.02,
	"db.m7g.xlarge":   254.04,
	"db.m7g.2xlarge":  508.08,
	"db.m7g.4xlarge":  1016.16,
	"db.m7g.8xlarge":  2032.32,
	"db.m7g.12xlarge": 3048.48,
	"db.m7g.16xlarge": 4064.64,
	// R5 family (memory optimized)
	"db.r5.large":    175.20,
	"db.r5.xlarge":   350.40,
	"db.r5.2xlarge":  700.80,
	"db.r5.4xlarge":  1401.60,
	"db.r5.8xlarge":  2803.20,
	"db.r5.12xlarge": 4204.80,
	"db.r5.16xlarge": 5606.40,
	"db.r5.24xlarge": 8409.60,
	// R6g family (Graviton, memory optimized)
	"db.r6g.large":    166.44,
	"db.r6g.xlarge":   332.88,
	"db.r6g.2xlarge":  665.76,
	"db.r6g.4xlarge":  1331.52,
	"db.r6g.8xlarge":  2663.04,
	"db.r6g.12xlarge": 3994.56,
	"db.r6g.16xlarge": 5326.08,
	// R7g family (Graviton3, memory optimized)
	"db.r7g.large":    178.85,
	"db.r7g.xlarge":   357.70,
	"db.r7g.2xlarge":  715.40,
	"db.r7g.4xlarge":  1430.80,
	"db.r7g.8xlarge":  2861.60,
	"db.r7g.12xlarge": 4292.40,
	"db.r7g.16xlarge": 5723.20,
}

// rdsStoragePerGBMonth returns the runtime-loaded RDS gp2 storage rate, or the default constant.
func rdsStoragePerGBMonth() float64 {
	if v, ok := lookupMonthly(priceKey(categoryRDSStorage, ""), 0); ok {
		return v
	}

	return RDSStorageCostPerGBMonth
}

// rdsSnapshotPerGBMonth returns the runtime-loaded RDS snapshot storage rate, or the default.
func rdsSnapshotPerGBMonth() float64 {
	if v, ok := lookupMonthly(priceKey(categoryRDSSnapshot, ""), 0); ok {
		return v
	}

	return RDSSnapshotCostPerGBMonth
}

// CalculateRDSInstanceMonthlyCost calculates the estimated monthly storage cost for a stopped RDS instance.
func CalculateRDSInstanceMonthlyCost(allocatedGB int32, multiAZ bool) float64 {
	cost := float64(allocatedGB) * rdsStoragePerGBMonth()
	if multiAZ {
		cost *= 2
	}

	return cost
}

// CalculateRDSSnapshotMonthlyCost calculates the estimated monthly storage cost for an RDS snapshot.
func CalculateRDSSnapshotMonthlyCost(allocatedGB int32) float64 {
	return float64(allocatedGB) * rdsSnapshotPerGBMonth()
}

// CalculateRDSIdleInstanceMonthlyCost calculates the estimated monthly cost for an idle RDS instance (compute + storage).
func CalculateRDSIdleInstanceMonthlyCost(instanceClass string, allocatedGB int32, multiAZ bool) float64 {
	computeCost := RDSInstanceComputeCost(instanceClass)
	storageCost := float64(allocatedGB) * rdsStoragePerGBMonth()

	total := computeCost + storageCost
	if multiAZ {
		total *= 2
	}

	return total
}

// RDSInstanceComputeCost returns the monthly compute cost for an RDS instance class, or 0 if
// unknown. The runtime cache is checked first (trying both the full class and the class with the
// "db." prefix stripped, since the Pricing API uses either in different contexts), and the
// returned hourly rate is multiplied by hoursPerMonth.
func RDSInstanceComputeCost(instanceClass string) float64 {
	if v, ok := lookupMonthly(priceKey(categoryRDSInstance, instanceClass), hoursPerMonth); ok {
		return v
	}

	if trimmed := strings.TrimPrefix(instanceClass, "db."); trimmed != instanceClass {
		if v, ok := lookupMonthly(priceKey(categoryRDSInstance, trimmed), hoursPerMonth); ok {
			return v
		}
	}

	return rdsInstancePricing[instanceClass]
}
