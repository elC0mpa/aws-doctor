package model

// EC2IdleInstanceInfo represents a running EC2 instance whose average CPU utilization and average
// daily network throughput across the lookback window were both below the configured thresholds.
type EC2IdleInstanceInfo struct {
	InstanceID           string  `json:"instance_id"`
	InstanceType         string  `json:"instance_type"`
	Name                 string  `json:"name,omitempty"`
	CPUUtilizationAvg    float64 `json:"cpu_utilization_avg_percent"`
	NetworkBytesPerDay   float64 `json:"network_bytes_per_day_avg"`
	DaysChecked          int     `json:"days_checked"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost"`
}
