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
	"golang.org/x/sync/errgroup"
)

// NewService creates a new RDS service.
func NewService(awsconfig aws.Config, cwService cloudwatchMetricsService, pricingSvc pricingService) Service {
	client := rds.NewFromConfig(awsconfig)

	return &service{
		client:         client,
		cwService:      cwService,
		pricingService: pricingSvc,
	}
}

func (s *service) Name() string {
	return "rds"
}

func (s *service) Analyze(ctx context.Context, flags model.Flags) (model.ScopeResult, error) {
	start := time.Now()
	input := model.RenderWasteInput{}

	var errs []error

	unusedInstances, unusedSnapshots, idleInstances, err := s.GetRDSWaste(ctx, flags.RDSIdleDays, flags.RDSSnapshotDays)
	if err != nil {
		errs = append(errs, err)
	} else {
		input.RDSInstances = unusedInstances
		input.RDSSnapshots = unusedSnapshots
		input.RDSIdleInstances = idleInstances
	}

	var finalErr error
	if len(errs) > 0 {
		finalErr = fmt.Errorf("rds analyze errors: %v", errs)
	}

	return model.ScopeResult{
		Scope:    s.Name(),
		Input:    input,
		Duration: time.Since(start),
		Err:      finalErr,
	}, nil
}

// GetRDSWaste returns stopped RDS instances, old manual snapshots, and idle instances.
func (s *service) GetRDSWaste(ctx context.Context, idleDays int, snapshotDays int) ([]model.RDSInstanceWasteInfo, []model.RDSSnapshotWasteInfo, []model.RDSIdleInstanceInfo, error) {
	g, ctx := errgroup.WithContext(ctx)

	var (
		stopped   []model.RDSInstanceWasteInfo
		idle      []model.RDSIdleInstanceInfo
		snapshots []model.RDSSnapshotWasteInfo
	)

	g.Go(func() error {
		var err error

		stopped, idle, err = s.getInstanceWaste(ctx, idleDays)

		return err
	})

	g.Go(func() error {
		var err error

		snapshots, err = s.getOldManualSnapshots(ctx, snapshotDays)

		return err
	})

	if err := g.Wait(); err != nil {
		return nil, nil, nil, err
	}

	return stopped, snapshots, idle, nil
}

func (s *service) getInstanceWaste(ctx context.Context, idleDays int) ([]model.RDSInstanceWasteInfo, []model.RDSIdleInstanceInfo, error) {
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
					EstimatedMonthlyCost: s.pricingService.CalculateRDSInstanceMonthlyCost(allocatedGB, multiAZ),
				})
			}

			if status == "available" {
				isIdle, err := s.cwService.RDSHasZeroConnectionsInPeriod(ctx, instanceID, idleDays)
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
						DaysChecked:          idleDays,
						EstimatedMonthlyCost: s.pricingService.CalculateRDSIdleInstanceMonthlyCost(instanceClass, allocatedGB, multiAZ),
					})
				}
			}
		}
	}

	return stopped, idle, nil
}

func (s *service) getOldManualSnapshots(ctx context.Context, snapshotDays int) ([]model.RDSSnapshotWasteInfo, error) {
	var result []model.RDSSnapshotWasteInfo

	now := time.Now()
	thresholdDate := now.AddDate(0, 0, -snapshotDays)

	paginator := rds.NewDescribeDBSnapshotsPaginator(s.client, &rds.DescribeDBSnapshotsInput{
		SnapshotType: aws.String("manual"),
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe RDS snapshots: %w", err)
		}

		for _, snap := range output.DBSnapshots {
			if snap.SnapshotCreateTime != nil && snap.SnapshotCreateTime.Before(thresholdDate) {
				daysSince := int(math.Floor(now.Sub(*snap.SnapshotCreateTime).Hours() / 24))
				allocatedGB := aws.ToInt32(snap.AllocatedStorage)

				result = append(result, model.RDSSnapshotWasteInfo{
					DBSnapshotID:         aws.ToString(snap.DBSnapshotIdentifier),
					DBInstanceID:         aws.ToString(snap.DBInstanceIdentifier),
					Engine:               aws.ToString(snap.Engine),
					AllocatedStorage:     allocatedGB,
					SnapshotCreateTime:   *snap.SnapshotCreateTime,
					DaysSinceCreate:      daysSince,
					EstimatedMonthlyCost: s.pricingService.CalculateRDSSnapshotMonthlyCost(allocatedGB),
				})
			}
		}
	}

	return result, nil
}
