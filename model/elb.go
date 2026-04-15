package model

// ELBIdleInfo represents a load balancer with zero connections over a period.
type ELBIdleInfo struct {
	Name                 string  `json:"name"`
	ARN                  string  `json:"arn"`
	Type                 string  `json:"type"`
	DaysChecked          int     `json:"days_checked"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost"`
}
