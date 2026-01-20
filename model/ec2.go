package model

import "time"

type ElasticIpInfo struct {
	UnusedElasticIpAddresses []string
	UsedElasticIpAddresses   []AttachedIpInfo
}

type AttachedIpInfo struct {
	IpAddress    string
	AllocationId string
	ResourceType string
}

type RiExpirationInfo struct {
	ReservedInstanceId string
	InstanceType       string
	ExpirationDate     time.Time
	DaysUntilExpiry    int
	State              string
	Status             string
}

// AMIWasteInfo contains information about potentially unused AMIs
type AMIWasteInfo struct {
	ImageId         string
	Name            string
	Description     string
	CreationDate    time.Time
	DaysSinceCreate int
	IsPublic        bool
	SnapshotIds     []string  // Associated EBS snapshots
	SnapshotSizeGB  int64     // Total size of associated snapshots
	UsedByInstances int       // Number of instances using this AMI
	EstimatedCost   float64   // Monthly storage cost of associated snapshots
}
