package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/cloudwatchmetrics"
)

// ClientAPI is the interface for the AWS RDS client methods used by the service.
type ClientAPI interface {
	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
	DescribeDBSnapshots(ctx context.Context, params *rds.DescribeDBSnapshotsInput, optFns ...func(*rds.Options)) (*rds.DescribeDBSnapshotsOutput, error)
}

// Service is the interface for AWS RDS service.
type Service interface {
	GetRDSWaste(ctx context.Context) ([]model.RDSInstanceWasteInfo, []model.RDSSnapshotWasteInfo, []model.RDSIdleInstanceInfo, error)
}

type service struct {
	client    ClientAPI
	cwService cloudwatchmetrics.Service
}
