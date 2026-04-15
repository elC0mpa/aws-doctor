package model

// LambdaOverProvisionedInfo represents a Lambda function using less than 10% of allocated memory.
type LambdaOverProvisionedInfo struct {
	FunctionName        string
	Runtime             string
	ConfiguredMemoryMB  int32
	MaxMemoryUsedMB     int32
	MemoryUtilization   float64
	RecommendedMemoryMB int32
}
