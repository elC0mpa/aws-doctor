package model

// HostedZoneWasteInfo contains information about potentially unused Route 53 hosted zones
type HostedZoneWasteInfo struct {
	HostedZoneId   string
	Name           string
	RecordSetCount int64   // Number of record sets (NS and SOA are always present)
	IsPrivate      bool
	Comment        string
	MonthlyCost    float64 // $0.50/month per hosted zone
	IsEmpty        bool    // True if only NS and SOA records exist
}
