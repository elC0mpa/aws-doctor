package model

// ELBIdleInfo represents a load balancer with zero connections over a period.
type ELBIdleInfo struct {
	LoadBalancerName     string
	LoadBalancerArn      string
	Type                 string // "application" or "network"
	DaysChecked          int
	EstimatedMonthlyCost float64
}
