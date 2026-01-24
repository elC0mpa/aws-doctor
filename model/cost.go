package model

import "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"

// CostInfo holds cost information for a specific period.
type CostInfo struct {
	types.DateInterval
	CostGroup CostGroup
}

// CostGroup maps service names to their respective costs.
type CostGroup map[string]struct {
	Amount float64
	Unit   string
}

// ServiceCost represents cost information for a specific service.
type ServiceCost struct {
	Name   string
	Amount float64
	Unit   string
}
