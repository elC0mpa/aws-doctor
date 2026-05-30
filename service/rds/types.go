package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/analyzer"
)

// ClientAPI is the interface for the AWS RDS client methods used by the service.
type ClientAPI interface {
	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
	DescribeDBSnapshots(ctx context.Context, params *rds.DescribeDBSnapshotsInput, optFns ...func(*rds.Options)) (*rds.DescribeDBSnapshotsOutput, error)
}

type service struct {
	client         ClientAPI
	cwService      cloudwatchMetricsService
	pricingService pricingService
}

// cloudwatchMetricsService is a local interface for the CloudWatch metrics dependency.
type cloudwatchMetricsService interface {
	RDSHasZeroConnectionsInPeriod(ctx context.Context, dbInstanceID string, days int) (bool, error)
}

// pricingService is a local interface for the pricing dependency.
type pricingService interface {
	CalculateRDSInstanceMonthlyCost(allocatedGB int32, multiAZ bool) float64
	CalculateRDSSnapshotMonthlyCost(allocatedGB int32) float64
	CalculateRDSIdleInstanceMonthlyCost(instanceClass string, allocatedGB int32, multiAZ bool) float64
}

// Service is the interface for AWS RDS service.
type Service interface {
	analyzer.WasteAnalyzer
	GetRDSWaste(ctx context.Context, idleDays int, snapshotDays int) ([]model.RDSInstanceWasteInfo, []model.RDSSnapshotWasteInfo, []model.RDSIdleInstanceInfo, error)
}
