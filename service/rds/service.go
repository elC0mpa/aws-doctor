// Package rds provides a service for detecting RDS waste.
package rds

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/pricing"
	"golang.org/x/sync/errgroup"
)

const idleDaysThreshold = 7

// NewService creates a new RDS service.
func NewService(awsconfig aws.Config, cwService cloudwatchMetricsService) Service {
	client := rds.NewFromConfig(awsconfig)

	return &service{
		client:    client,
		cwService: cwService,
	}
}

// GetRDSWaste returns stopped RDS instances, old manual snapshots (>30 days), and idle instances.
func (s *service) GetRDSWaste(ctx context.Context) ([]model.RDSInstanceWasteInfo, []model.RDSSnapshotWasteInfo, []model.RDSIdleInstanceInfo, error) {
	g, ctx := errgroup.WithContext(ctx)

	var (
		stopped   []model.RDSInstanceWasteInfo
		idle      []model.RDSIdleInstanceInfo
		snapshots []model.RDSSnapshotWasteInfo
	)

	g.Go(func() error {
		var err error

		stopped, idle, err = s.getInstanceWaste(ctx)

		return err
	})

	g.Go(func() error {
		var err error

		snapshots, err = s.getOldManualSnapshots(ctx)

		return err
	})

	if err := g.Wait(); err != nil {
		return nil, nil, nil, err
	}

	return stopped, snapshots, idle, nil
}

func (s *service) getInstanceWaste(ctx context.Context) ([]model.RDSInstanceWasteInfo, []model.RDSIdleInstanceInfo, error) {
	var stopped []model.RDSInstanceWasteInfo

	var idle []model.RDSIdleInstanceInfo

	paginator := rds.NewDescribeDBInstancesPaginator(s.client, &rds.DescribeDBInstancesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to describe RDS instances: %w", err)
		}

		for _, db := range output.DBInstances {
			status := aws.ToString(db.DBInstanceStatus)
			allocatedGB := aws.ToInt32(db.AllocatedStorage)
			multiAZ := aws.ToBool(db.MultiAZ)
			instanceClass := aws.ToString(db.DBInstanceClass)
			instanceID := aws.ToString(db.DBInstanceIdentifier)

			if status == "stopped" {
				stopped = append(stopped, model.RDSInstanceWasteInfo{
					DBInstanceID:         instanceID,
					DBInstanceClass:      instanceClass,
					Engine:               aws.ToString(db.Engine),
					Status:               status,
					MultiAZ:              multiAZ,
					AllocatedStorage:     allocatedGB,
					EstimatedMonthlyCost: pricing.CalculateRDSInstanceMonthlyCost(allocatedGB, multiAZ),
				})
			}

			if status == "available" {
				isIdle, err := s.cwService.RDSHasZeroConnectionsInPeriod(ctx, instanceID, idleDaysThreshold)
				if err != nil {
					continue
				}

				if isIdle {
					idle = append(idle, model.RDSIdleInstanceInfo{
						DBInstanceID:         instanceID,
						DBInstanceClass:      instanceClass,
						Engine:               aws.ToString(db.Engine),
						MultiAZ:              multiAZ,
						AllocatedStorage:     allocatedGB,
						DaysChecked:          idleDaysThreshold,
						EstimatedMonthlyCost: pricing.CalculateRDSIdleInstanceMonthlyCost(instanceClass, allocatedGB, multiAZ),
					})
				}
			}
		}
	}

	return stopped, idle, nil
}

func (s *service) getOldManualSnapshots(ctx context.Context) ([]model.RDSSnapshotWasteInfo, error) {
	var result []model.RDSSnapshotWasteInfo

	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	paginator := rds.NewDescribeDBSnapshotsPaginator(s.client, &rds.DescribeDBSnapshotsInput{
		SnapshotType: aws.String("manual"),
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe RDS snapshots: %w", err)
		}

		for _, snap := range output.DBSnapshots {
			if snap.SnapshotCreateTime != nil && snap.SnapshotCreateTime.Before(thirtyDaysAgo) {
				daysSince := int(math.Floor(now.Sub(*snap.SnapshotCreateTime).Hours() / 24))
				allocatedGB := aws.ToInt32(snap.AllocatedStorage)

				result = append(result, model.RDSSnapshotWasteInfo{
					DBSnapshotID:         aws.ToString(snap.DBSnapshotIdentifier),
					DBInstanceID:         aws.ToString(snap.DBInstanceIdentifier),
					Engine:               aws.ToString(snap.Engine),
					AllocatedStorage:     allocatedGB,
					SnapshotCreateTime:   *snap.SnapshotCreateTime,
					DaysSinceCreate:      daysSince,
					EstimatedMonthlyCost: pricing.CalculateRDSSnapshotMonthlyCost(allocatedGB),
				})
			}
		}
	}

	return result, nil
}
