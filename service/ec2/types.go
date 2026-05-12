package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elC0mpa/aws-doctor/model"
)

// ClientAPI is the interface for the AWS EC2 client methods used by the service.
type ClientAPI interface {
	DescribeAddresses(ctx context.Context, params *ec2.DescribeAddressesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeAddressesOutput, error)
	DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error)
	DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeReservedInstances(ctx context.Context, params *ec2.DescribeReservedInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeReservedInstancesOutput, error)
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error)
	DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error)
}

type service struct {
	client         ClientAPI
	cwService      cloudwatchMetricsService
	pricingService pricingService
}

// pricingService is a local interface for the pricing dependency.
type pricingService interface {
	CalculateEBSMonthlyCost(sizeGiB int32, volumeType types.VolumeType) float64
	CalculateEBSSnapshotMonthlyCost(sizeGB int64) float64
	CalculateEC2InstanceMonthlyCost(instanceType string) float64
}

// cloudwatchMetricsService is a local interface for the CloudWatch metrics dependency.
type cloudwatchMetricsService interface {
	EC2InstanceIdleStats(ctx context.Context, instanceID string, days int) (cpuAvgPercent, networkBytesPerDay float64, err error)
}

// Service is the interface for AWS EC2 service.
type Service interface {
	GetElasticIPAddressesInfo(ctx context.Context) (*model.ElasticIPInfo, error)
	GetUnusedElasticIPAddressesInfo(ctx context.Context) ([]types.Address, error)
	GetUnusedEBSVolumes(ctx context.Context) ([]types.Volume, error)
	GetStoppedInstancesInfo(ctx context.Context) ([]types.Instance, []types.Volume, error)
	GetReservedInstanceExpiringOrExpired30DaysWaste(ctx context.Context) ([]model.RiExpirationInfo, error)
	GetUnusedAMIs(ctx context.Context, staleDays int) ([]model.AMIWasteInfo, error)
	GetOrphanedSnapshots(ctx context.Context, staleDays int) ([]model.SnapshotWasteInfo, error)
	GetUnusedKeyPairs(ctx context.Context) ([]model.KeyPairWasteInfo, error)
	GetIdleInstances(ctx context.Context, idleDays int, cpuThresholdPercent float64, networkBytesPerDayThreshold float64) ([]model.EC2IdleInstanceInfo, error)
}
