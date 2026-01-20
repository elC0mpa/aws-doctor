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
	Months      []MonthCostJSON `json:"months"`
}

// MonthCostJSON represents cost data for a single month
type MonthCostJSON struct {
	Start string  `json:"start"`
	End   string  `json:"end"`
	Total float64 `json:"total"`
	Unit  string  `json:"unit"`
}

// WasteReportJSON represents the JSON output for waste detection
type WasteReportJSON struct {
	AccountID           string                 `json:"account_id"`
	GeneratedAt         string                 `json:"generated_at"`
	HasWaste            bool                   `json:"has_waste"`
	UnusedElasticIPs    []ElasticIPJSON        `json:"unused_elastic_ips"`
	UnusedEBSVolumes    []EBSVolumeJSON        `json:"unused_ebs_volumes"`
	StoppedVolumes      []EBSVolumeJSON        `json:"stopped_instance_volumes"`
	StoppedInstances    []StoppedInstanceJSON  `json:"stopped_instances"`
	ReservedInstances   []ReservedInstanceJSON `json:"reserved_instances"`
	UnusedLoadBalancers []LoadBalancerJSON     `json:"unused_load_balancers"`
}

// ElasticIPJSON represents an unused Elastic IP
type ElasticIPJSON struct {
	PublicIP     string `json:"public_ip"`
	AllocationID string `json:"allocation_id"`
}

// EBSVolumeJSON represents an EBS volume
type EBSVolumeJSON struct {
	VolumeID string `json:"volume_id"`
	Size     int32  `json:"size_gib"`
	Status   string `json:"status"`
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
	Name string `json:"name"`
	ARN  string `json:"arn"`
	Type string `json:"type"`
}
