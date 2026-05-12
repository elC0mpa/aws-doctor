package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elC0mpa/aws-doctor/model"
)

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

	var idle []model.EC2IdleInstanceInfo

	paginator := ec2.NewDescribeInstancesPaginator(s.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe instances: %w", err)
		}

		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				instanceID := aws.ToString(instance.InstanceId)
				if instanceID == "" {
					continue
				}

				cpuAvg, networkBytesPerDay, err := s.cwService.EC2InstanceIdleStats(ctx, instanceID, idleDays)
				if err != nil {
					continue
				}

				if cpuAvg >= cpuThresholdPercent || networkBytesPerDay >= networkBytesPerDayThreshold {
					continue
				}

				instanceType := string(instance.InstanceType)

				idle = append(idle, model.EC2IdleInstanceInfo{
					InstanceID:           instanceID,
					InstanceType:         instanceType,
					Name:                 nameTag(instance.Tags),
					CPUUtilizationAvg:    cpuAvg,
					NetworkBytesPerDay:   networkBytesPerDay,
					DaysChecked:          idleDays,
					EstimatedMonthlyCost: s.pricingService.CalculateEC2InstanceMonthlyCost(instanceType),
				})
			}
		}
	}

	return idle, nil
}

func nameTag(tags []types.Tag) string {
	for _, t := range tags {
		if aws.ToString(t.Key) == "Name" {
			return aws.ToString(t.Value)
		}
	}

	return ""
}
