package pricing

// sagemakerInstancePricing maps SageMaker real-time inference instance types to approximate
// monthly costs (us-east-1 on-demand hourly × 730h). These rates are used as a fallback when the
// Pricing API does not return a match, and match the same source as the RDS fallback table
// (AWS public pricing pages).
//
//nolint:gochecknoglobals // pricing lookup table
var sagemakerInstancePricing = map[string]float64{
	// CPU / general purpose
	"ml.t2.medium":   46.72,
	"ml.t2.large":    93.44,
	"ml.t2.xlarge":   186.15,
	"ml.t2.2xlarge":  372.30,
	"ml.t3.medium":   41.61,
	"ml.t3.large":    83.22,
	"ml.t3.xlarge":   166.44,
	"ml.t3.2xlarge":  332.88,
	"ml.m4.xlarge":   204.40,
	"ml.m4.2xlarge":  408.80,
	"ml.m4.4xlarge":  817.60,
	"ml.m5.large":    101.91,
	"ml.m5.xlarge":   203.82,
	"ml.m5.2xlarge":  407.64,
	"ml.m5.4xlarge":  815.28,
	"ml.m5.12xlarge": 2445.84,
	"ml.m5.24xlarge": 4891.68,
	"ml.c4.xlarge":   176.66,
	"ml.c4.2xlarge":  353.32,
	"ml.c4.4xlarge":  706.64,
	"ml.c5.large":    87.60,
	"ml.c5.xlarge":   175.20,
	"ml.c5.2xlarge":  350.40,
	"ml.c5.4xlarge":  700.80,
	"ml.c5.9xlarge":  1576.80,
	// GPU
	"ml.p2.xlarge":    1144.80,
	"ml.p2.8xlarge":   9154.80,
	"ml.p3.2xlarge":   3183.84,
	"ml.p3.8xlarge":   12735.36,
	"ml.p3.16xlarge":  25470.72,
	"ml.g4dn.xlarge":  538.72,
	"ml.g4dn.2xlarge": 770.42,
	"ml.g4dn.4xlarge": 1232.88,
	"ml.g4dn.8xlarge": 2227.68,
	"ml.g5.xlarge":    802.82,
	"ml.g5.2xlarge":   966.48,
	"ml.g5.4xlarge":   1295.20,
	"ml.g5.8xlarge":   1950.64,
	"ml.g5.12xlarge":  5058.96,
	"ml.g5.24xlarge":  8049.28,
	// Inferentia
	"ml.inf1.xlarge":   218.40,
	"ml.inf1.2xlarge":  347.48,
	"ml.inf1.6xlarge":  1128.32,
	"ml.inf1.24xlarge": 4497.92,
}

// CalculateSageMakerEndpointMonthlyCost returns the approximate monthly on-demand cost for a
// SageMaker real-time inference endpoint given the sum of (instance type, instance count) pairs
// across its production variants. Unknown instance types contribute 0 to the total so the caller
// at least sees a partial cost estimate.
func CalculateSageMakerEndpointMonthlyCost(variants []SageMakerVariantCost) float64 {
	var total float64

	for _, v := range variants {
		total += sagemakerInstancePricing[v.InstanceType] * float64(v.InstanceCount)
	}

	return total
}

// SageMakerVariantCost is one (instance type, count) pair within an endpoint's production
// variants, used for cost estimation.
type SageMakerVariantCost struct {
	InstanceType  string
	InstanceCount int32
}
