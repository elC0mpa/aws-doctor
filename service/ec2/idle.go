package ec2

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"golang.org/x/sync/errgroup"
)

// idleScanConcurrency caps the number of concurrent CloudWatch lookups when scanning running
// instances. Keeping this small avoids hammering the CloudWatch GetMetricStatistics throttle
// when an account has hundreds of running instances.
const idleScanConcurrency = 10

// GetIdleInstances returns running EC2 instances whose average CPU utilization and average daily
// network throughput have both been below the supplied thresholds over the lookback window.
//
// CloudWatch failures on individual instances are skipped so a single broken instance does not
// hide other waste, matching the behavior of the RDS idle check.
func (s *service) GetIdleInstances(ctx context.Context, idleDays int, cpuThresholdPercent float64, networkBytesPerDayThreshold float64) ([]model.EC2IdleInstanceInfo, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running"},
			},
		},
	}

	var instances []types.Instance

	paginator := ec2.NewDescribeInstancesPaginator(s.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe instances: %w", err)
		}

		for _, reservation := range page.Reservations {
			instances = append(instances, reservation.Instances...)
		}
	}

	return s.evaluateIdleInstances(ctx, instances, idleDays, cpuThresholdPercent, networkBytesPerDayThreshold), nil
}

func (s *service) evaluateIdleInstances(ctx context.Context, instances []types.Instance, idleDays int, cpuThresholdPercent float64, networkBytesPerDayThreshold float64) []model.EC2IdleInstanceInfo {
	// Each goroutine writes into its own slot so the result aggregation needs no mutex.
	results := make([]*model.EC2IdleInstanceInfo, len(instances))

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(idleScanConcurrency)

	for i, instance := range instances {
		i, instance := i, instance

		instanceID := aws.ToString(instance.InstanceId)
		if instanceID == "" {
			continue
		}

		g.Go(func() error {
			cpuAvg, networkBytesPerDay, err := s.cwService.EC2InstanceIdleStats(ctx, instanceID, idleDays)
			if err != nil {
				return nil
			}

			if cpuAvg >= cpuThresholdPercent || networkBytesPerDay >= networkBytesPerDayThreshold {
				return nil
			}

			instanceType := string(instance.InstanceType)
			results[i] = &model.EC2IdleInstanceInfo{
				InstanceID:           instanceID,
				InstanceType:         instanceType,
				Name:                 nameTag(instance.Tags),
				CPUUtilizationAvg:    cpuAvg,
				NetworkBytesPerDay:   networkBytesPerDay,
				DaysChecked:          idleDays,
				EstimatedMonthlyCost: s.pricingService.CalculateEC2InstanceMonthlyCost(instanceType),
			}

			return nil
		})
	}

	_ = g.Wait()

	return flattenIdleResults(results)
}

func flattenIdleResults(results []*model.EC2IdleInstanceInfo) []model.EC2IdleInstanceInfo {
	var idle []model.EC2IdleInstanceInfo

	for _, r := range results {
		if r != nil {
			idle = append(idle, *r)
		}
	}

	return idle
}

// nameTag returns the value of the Name tag if present. Matching is case-insensitive so that
// instances tagged with "name" or "NAME" are still picked up.
func nameTag(tags []types.Tag) string {
	for _, t := range tags {
		if strings.EqualFold(aws.ToString(t.Key), "Name") {
			return aws.ToString(t.Value)
		}
	}

	return ""
}
