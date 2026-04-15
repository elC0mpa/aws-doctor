package model

// NATGatewayWasteInfo contains information about idle NAT Gateways
type NATGatewayWasteInfo struct {
	NATGatewayID          string
	VPCID                 string
	SubnetID              string
	State                 string
	BytesOutToDestination float64
	EstimatedMonthlyCost  float64
	DaysSinceCreate       int
}
