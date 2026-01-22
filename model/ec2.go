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

// SnapshotCategory indicates whether a snapshot is orphaned or stale
type SnapshotCategory string

const (
	// SnapshotCategoryOrphaned - source volume deleted, safe to delete (high confidence)
	SnapshotCategoryOrphaned SnapshotCategory = "orphaned"
	// SnapshotCategoryStale - volume exists but snapshot is old, needs review (low confidence)
	SnapshotCategoryStale SnapshotCategory = "stale"
)

// SnapshotWasteInfo contains information about potentially orphaned EBS snapshots
type SnapshotWasteInfo struct {
	SnapshotId          string
	VolumeId            string           // Source volume ID (may no longer exist)
	VolumeExists        bool             // Whether the source volume still exists
	UsedByAMI           bool             // Whether snapshot is used by an AMI
	AMIId               string           // AMI ID if used
	SizeGB              int32            // Snapshot size in GB
	StartTime           time.Time        // When snapshot was created
	DaysSinceCreate     int              // Days since creation
	Description         string
	Category            SnapshotCategory // "orphaned" or "stale"
	Reason              string           // Human-readable reason (e.g., "Volume Deleted", "Old Backup")
	MaxPotentialSavings float64          // Max monthly savings (actual may be lower due to incremental storage)
}
