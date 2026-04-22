package model

// IdleSageMakerEndpointInfo represents a real-time inference endpoint that served zero
// invocations across all production variants over a recent lookback period.
type IdleSageMakerEndpointInfo struct {
	EndpointName         string             `json:"endpoint_name"`
	EndpointARN          string             `json:"endpoint_arn"`
	Status               string             `json:"status"`
	Variants             []SageMakerVariant `json:"variants"`
	DaysChecked          int                `json:"days_checked"`
	EstimatedMonthlyCost float64            `json:"estimated_monthly_cost"`
}

// SageMakerVariant is a single production variant of a SageMaker endpoint.
type SageMakerVariant struct {
	VariantName   string `json:"variant_name"`
	InstanceType  string `json:"instance_type"`
	InstanceCount int32  `json:"instance_count"`
}
