package model

// CostComparisonJSON represents the JSON output for cost comparison
type CostComparisonJSON struct {
	AccountID        string                   `json:"account_id"`
	GeneratedAt      string                   `json:"generated_at"`
	CurrentMonth     CostPeriodJSON           `json:"current_month"`
	LastMonth        CostPeriodJSON           `json:"last_month"`
	ServiceBreakdown []ServiceCostCompareJSON `json:"service_breakdown"`
}

// CostPeriodJSON represents cost data for a time period
type CostPeriodJSON struct {
	Start string  `json:"start"`
	End   string  `json:"end"`
	Total float64 `json:"total"`
	Unit  string  `json:"unit"`
}

// ServiceCostCompareJSON represents cost comparison for a single service
type ServiceCostCompareJSON struct {
	Service     string  `json:"service"`
	CurrentCost float64 `json:"current_cost"`
	LastCost    float64 `json:"last_cost"`
	Difference  float64 `json:"difference"`
	Unit        string  `json:"unit"`
}

// TrendJSON represents the JSON output for trend analysis
type TrendJSON struct {
	AccountID   string          `json:"account_id"`
	GeneratedAt string          `json:"generated_at"`
	Services    []string        `json:"services,omitempty"`
	Months      []MonthCostJSON `json:"months"`
}

// MonthCostJSON represents cost data for a single month
type MonthCostJSON struct {
	Start string  `json:"start"`
	End   string  `json:"end"`
	Total float64 `json:"total"`
	Unit  string  `json:"unit"`
}

// WasteSummaryJSON represents a single category in the waste summary.
type WasteSummaryJSON struct {
	Category             string  `json:"category"`
	Count                int     `json:"count"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost"`
}

// WasteReportJSON represents the JSON output for waste detection
type WasteReportJSON struct {
	AccountID                 string                   `json:"account_id"`
	GeneratedAt               string                   `json:"generated_at"`
	HasWaste                  bool                     `json:"has_waste"`
	TotalEstimatedMonthlyCost float64                  `json:"total_estimated_monthly_cost"`
	Summary                   []WasteSummaryJSON       `json:"summary"`
	UnusedElasticIPs          []ElasticIPJSON          `json:"unused_elastic_ips"`
	UnusedEBSVolumes          []EBSVolumeJSON          `json:"unused_ebs_volumes"`
	StoppedVolumes            []EBSVolumeJSON          `json:"stopped_instance_volumes"`
	StoppedInstances          []StoppedInstanceJSON    `json:"stopped_instances"`
	ReservedInstances         []ReservedInstanceJSON   `json:"reserved_instances"`
	UnusedLoadBalancers       []LoadBalancerJSON       `json:"unused_load_balancers"`
	UnusedAMIs                []AMIJSON                `json:"unused_amis"`
	OrphanedSnapshots         []SnapshotJSON           `json:"orphaned_snapshots"`
	StaleSnapshots            []SnapshotJSON           `json:"stale_snapshots"`
	UnusedKeyPairs            []KeyPairJSON            `json:"unused_key_pairs"`
	S3Buckets                 []S3BucketJSON           `json:"s3_buckets_without_lifecycle"`
	S3MultipartUploads        []S3MultipartJSON        `json:"s3_buckets_with_incomplete_multipart_uploads"`
	CloudWatchLogGroups       []CloudWatchLogGroupJSON `json:"cloudwatch_log_groups_without_retention_policy"`
	StoppedRDSInstances       []RDSInstanceJSON        `json:"stopped_rds_instances"`
	OldRDSSnapshots           []RDSSnapshotJSON        `json:"old_rds_snapshots"`
	IdleRDSInstances          []RDSIdleInstanceJSON    `json:"idle_rds_instances"`
	IdleNATGateways           []NATGatewayJSON         `json:"idle_nat_gateways"`
	IdleLoadBalancers         []ELBIdleJSON            `json:"idle_load_balancers"`
}

// RDSInstanceJSON represents a stopped RDS instance.
type RDSInstanceJSON struct {
	DBInstanceID         string  `json:"db_instance_id"`
	DBInstanceClass      string  `json:"db_instance_class"`
	Engine               string  `json:"engine"`
	Status               string  `json:"status"`
	MultiAZ              bool    `json:"multi_az"`
	AllocatedStorage     int32   `json:"allocated_storage_gb"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost"`
}

// RDSSnapshotJSON represents an old manual RDS snapshot.
type RDSSnapshotJSON struct {
	DBSnapshotID         string  `json:"db_snapshot_id"`
	DBInstanceID         string  `json:"db_instance_id"`
	Engine               string  `json:"engine"`
	AllocatedStorage     int32   `json:"allocated_storage_gb"`
	SnapshotCreateTime   string  `json:"snapshot_create_time"`
	DaysSinceCreate      int     `json:"days_since_create"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost"`
}

// RDSIdleInstanceJSON represents an idle RDS instance with no connections.
type RDSIdleInstanceJSON struct {
	DBInstanceID         string  `json:"db_instance_id"`
	DBInstanceClass      string  `json:"db_instance_class"`
	Engine               string  `json:"engine"`
	MultiAZ              bool    `json:"multi_az"`
	AllocatedStorage     int32   `json:"allocated_storage_gb"`
	DaysChecked          int     `json:"days_checked"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost"`
}

// ELBIdleJSON represents an idle load balancer with zero connections.
type ELBIdleJSON struct {
	Name                 string  `json:"name"`
	ARN                  string  `json:"arn"`
	Type                 string  `json:"type"`
	DaysChecked          int     `json:"days_checked"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost"`
}

// CloudWatchLogGroupJSON represents a CloudWatch Log Group without a retention policy
type CloudWatchLogGroupJSON struct {
	LogGroupName         string  `json:"log_group_name"`
	CreationTime         string  `json:"creation_time"`
	StoredBytes          int64   `json:"stored_bytes"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost"`
}

// S3BucketJSON represents an S3 bucket without lifecycle policy
type S3BucketJSON struct {
	BucketName   string `json:"bucket_name"`
	CreationDate string `json:"creation_date"`
}

// S3MultipartJSON represents an S3 bucket with incomplete multipart uploads
type S3MultipartJSON struct {
	BucketName  string `json:"bucket_name"`
	UploadCount int    `json:"upload_count"`
}

// ElasticIPJSON represents an unused Elastic IP
type ElasticIPJSON struct {
	PublicIP             string  `json:"public_ip"`
	AllocationID         string  `json:"allocation_id"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost"`
}

// EBSVolumeJSON represents an EBS volume
type EBSVolumeJSON struct {
	VolumeID             string  `json:"volume_id"`
	Size                 int32   `json:"size_gib"`
	Status               string  `json:"status"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost"`
}

// StoppedInstanceJSON represents a stopped EC2 instance
type StoppedInstanceJSON struct {
	InstanceID string `json:"instance_id"`
	StoppedAt  string `json:"stopped_at,omitempty"`
	DaysAgo    int    `json:"days_ago,omitempty"`
}

// ReservedInstanceJSON represents a reserved instance
type ReservedInstanceJSON struct {
	ReservedInstanceID string `json:"reserved_instance_id"`
	InstanceType       string `json:"instance_type"`
	ExpirationDate     string `json:"expiration_date"`
	DaysUntilExpiry    int    `json:"days_until_expiry"`
	State              string `json:"state"`
	Status             string `json:"status"`
}

// LoadBalancerJSON represents an unused load balancer
type LoadBalancerJSON struct {
	Name                 string  `json:"name"`
	ARN                  string  `json:"arn"`
	Type                 string  `json:"type"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost"`
}

// AMIJSON represents an unused AMI
type AMIJSON struct {
	ImageID            string   `json:"image_id"`
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	CreationDate       string   `json:"creation_date"`
	DaysSinceCreate    int      `json:"days_since_create"`
	IsPublic           bool     `json:"is_public"`
	SnapshotIDs        []string `json:"snapshot_ids"`
	SnapshotSizeGB     int64    `json:"snapshot_size_gb"`
	MaxPotentialSaving float64  `json:"max_potential_saving_monthly"`
	SafetyWarning      string   `json:"safety_warning"`
}

// SnapshotJSON represents an orphaned or stale EBS snapshot
type SnapshotJSON struct {
	SnapshotID          string  `json:"snapshot_id"`
	VolumeID            string  `json:"volume_id,omitempty"`
	VolumeExists        bool    `json:"volume_exists"`
	UsedByAMI           bool    `json:"used_by_ami"`
	AMIID               string  `json:"ami_id,omitempty"`
	SizeGB              int32   `json:"size_gb"`
	StartTime           string  `json:"start_time"`
	DaysSinceCreate     int     `json:"days_since_create"`
	Description         string  `json:"description,omitempty"`
	Category            string  `json:"category"`              // "orphaned" or "stale"
	Reason              string  `json:"reason"`                // Human-readable reason
	MaxPotentialSavings float64 `json:"max_potential_savings"` // Actual savings may be lower due to incremental storage
}

// KeyPairJSON represents an unused EC2 key pair
type KeyPairJSON struct {
	KeyName         string `json:"key_name"`
	KeyPairID       string `json:"key_pair_id"`
	CreationDate    string `json:"creation_date"`
	DaysSinceCreate int    `json:"days_since_create"`
}

// NATGatewayJSON represents an idle NAT Gateway
type NATGatewayJSON struct {
	NATGatewayID          string  `json:"nat_gateway_id"`
	VPCID                 string  `json:"vpc_id"`
	SubnetID              string  `json:"subnet_id"`
	State                 string  `json:"state"`
	BytesOutToDestination float64 `json:"bytes_out_to_destination"`
	EstimatedMonthlyCost  float64 `json:"estimated_monthly_cost"`
	DaysSinceCreate       int     `json:"days_since_create"`
}
