package model

import "time"

// CloudWatchLogsWasteInfo represents information about a CloudWatch Log Group without a retention policy.
type CloudWatchLogsWasteInfo struct {
	LogGroupName         string
	CreationTime         time.Time
	StoredBytes          int64
	EstimatedMonthlyCost float64 // Estimated monthly storage cost
}
