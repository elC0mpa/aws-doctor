package model

import "time"

// RDSInstanceWasteInfo represents information about a stopped RDS instance.
type RDSInstanceWasteInfo struct {
	DBInstanceID         string
	DBInstanceClass      string
	Engine               string
	Status               string
	MultiAZ              bool
	AllocatedStorage     int32
	EstimatedMonthlyCost float64
}

// RDSIdleInstanceInfo represents an RDS instance with zero connections over a recent period.
type RDSIdleInstanceInfo struct {
	DBInstanceID         string
	DBInstanceClass      string
	Engine               string
	MultiAZ              bool
	AllocatedStorage     int32
	DaysChecked          int
	EstimatedMonthlyCost float64
}

// RDSSnapshotWasteInfo represents information about an old manual RDS snapshot.
type RDSSnapshotWasteInfo struct {
	DBSnapshotID         string
	DBInstanceID         string
	Engine               string
	AllocatedStorage     int32
	SnapshotCreateTime   time.Time
	DaysSinceCreate      int
	EstimatedMonthlyCost float64
}
