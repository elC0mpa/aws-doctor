package model

// NatGatewayWasteInfo contains information about idle NAT Gateways
type NatGatewayWasteInfo struct {
	NatGatewayID          string
	VPCID                 string
	SubnetID              string
	State                 string
	BytesOutToDestination float64
}
