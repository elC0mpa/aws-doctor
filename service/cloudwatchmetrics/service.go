// Package cloudwatchmetrics provides a service for querying CloudWatch metrics.
package cloudwatchmetrics

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// NewService creates a new CloudWatch metrics service.
func NewService(awsconfig aws.Config) Service {
	client := cloudwatch.NewFromConfig(awsconfig)

	return &service{
		client: client,
	}
}

// RDSHasZeroConnectionsInPeriod checks if an RDS instance had zero DatabaseConnections over the given number of days.
func (s *service) RDSHasZeroConnectionsInPeriod(ctx context.Context, dbInstanceID string, days int) (bool, error) {
	now := time.Now()
	startTime := now.AddDate(0, 0, -days)

	output, err := s.client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/RDS"),
		MetricName: aws.String("DatabaseConnections"),
		Dimensions: []cwtypes.Dimension{
			{
				Name:  aws.String("DBInstanceIdentifier"),
				Value: aws.String(dbInstanceID),
			},
		},
		StartTime:  &startTime,
		EndTime:    &now,
		Period:     aws.Int32(86400),
		Statistics: []cwtypes.Statistic{cwtypes.StatisticSum},
	})
	if err != nil {
		return false, fmt.Errorf("failed to get CloudWatch metrics for %s: %w", dbInstanceID, err)
	}

	for _, dp := range output.Datapoints {
		if dp.Sum != nil && *dp.Sum > 0 {
			return false, nil
		}
	}

	return true, nil
}
