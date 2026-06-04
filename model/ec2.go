package model

import "time"

// ElasticIPInfo holds information about unused and used Elastic IPs.
type ElasticIPInfo struct {
	UnusedElasticIPAddresses []string
	UsedElasticIPAddresses   []AttachedIPInfo
}

// AttachedIPInfo holds information about an IP address attached to a resource.
type AttachedIPInfo struct {
	IPAddress    string
	AllocationID string
	ResourceType string
}

// RiExpirationInfo holds information about Reserved Instance expirations.
type RiExpirationInfo struct {
	ReservedInstanceID string
	InstanceType       string
	ExpirationDate     time.Time
	DaysUntilExpiry    int
	State              string
	Status             string
}

// AMIWasteInfo contains information about potentially unused AMIs
type AMIWasteInfo struct {
	ImageID            string
	Name               string
	Description        string
	CreationDate       time.Time
	DaysSinceCreate    int
	IsPublic           bool
	SnapshotIDs        []string // Associated EBS snapshots
	SnapshotSizeGB     int64    // Total size of associated snapshots
	UsedByInstances    int      // Number of instances using this AMI
	MaxPotentialSaving float64  // Max potential monthly savings (snapshot storage cost)
	SafetyWarning      string   // Warning about potential ASG/Launch Template usage
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
	SnapshotID          string
	VolumeID            string    // Source volume ID (may no longer exist)
	VolumeExists        bool      // Whether the source volume still exists
	UsedByAMI           bool      // Whether snapshot is used by an AMI
	AMIID               string    // AMI ID if used
	SizeGB              int32     // Snapshot size in GB
	StartTime           time.Time // When snapshot was created
	DaysSinceCreate     int       // Days since creation
	Description         string
	Category            SnapshotCategory // "orphaned" or "stale"
	Reason              string           // Human-readable reason (e.g., "Volume Deleted", "Old Backup")
	MaxPotentialSavings float64          // Max monthly savings (actual may be lower due to incremental storage)
}

// PublicIPv4Summary reports the total count of in-use public IPv4 addresses
// across EC2 instances, Elastic IPs, NAT Gateways, and ENIs, together with
// the estimated hourly and monthly cost at the AWS per-IP rate.
// Since February 2024 AWS charges $0.005/hr for every public IPv4 in use.
type PublicIPv4Summary struct {
	// TotalCount is the total number of public IPv4 addresses detected.
	TotalCount int
	// EC2InstanceIPs is the count from EC2 instances with a public IP.
	EC2InstanceIPs int
	// ElasticIPs is the count from allocated Elastic IP addresses (in use).
	ElasticIPs int
	// NatGatewayIPs is the count from NAT Gateway public IPs.
	NatGatewayIPs int
	// OtherENIIPs is the count from other ENIs (LBs, RDS, etc).
	OtherENIIPs int
	// HourlyCostUSD is estimated cost at $0.005 per IP per hour.
	HourlyCostUSD float64
	// MonthlyCostUSD is estimated cost per 730-hour month.
	MonthlyCostUSD float64
}

// PublicIPv4RatePerHour is the AWS per-IP hourly charge (since Feb 2024).
const PublicIPv4RatePerHour = 0.005

// HoursPerMonth is the standard 730-hour billing month used by AWS for monthly cost estimates.
const HoursPerMonth = 730.0

// KeyPairWasteInfo contains information about unused EC2 key pairs
type KeyPairWasteInfo struct {
	KeyName         string
	KeyPairID       string
	CreateTime      time.Time
	DaysSinceCreate int
}
