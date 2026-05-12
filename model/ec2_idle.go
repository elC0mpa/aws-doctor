package model

// EC2IdleInstanceInfo represents a running EC2 instance whose average CPU utilization and average
// daily network throughput across the lookback window were both below the configured thresholds.
type EC2IdleInstanceInfo struct {
	InstanceID           string
	InstanceType         string
	Name                 string
	CPUUtilizationAvg    float64
	NetworkBytesPerDay   float64
	DaysChecked          int
	EstimatedMonthlyCost float64
}
