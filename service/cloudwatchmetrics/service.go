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

const (
	// metricPeriodSeconds is the period for CloudWatch metric queries (1 day in seconds).
	// Using daily periods ensures we capture all metrics even if they are reported less frequently.
	metricPeriodSeconds = 86400
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
		Period:     aws.Int32(metricPeriodSeconds),
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

// NatGatewayBytesOut returns the total bytes out to destination for a NAT Gateway over the given number of days.
func (s *service) NatGatewayBytesOut(ctx context.Context, natGatewayID string, days int) (float64, error) {
	now := time.Now()
	startTime := now.AddDate(0, 0, -days)

	output, err := s.client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/NATGateway"),
		MetricName: aws.String("BytesOutToDestination"),
		Dimensions: []cwtypes.Dimension{
			{
				Name:  aws.String("NatGatewayId"),
				Value: aws.String(natGatewayID),
			},
		},
		StartTime:  &startTime,
		EndTime:    &now,
		Period:     aws.Int32(metricPeriodSeconds),
		Statistics: []cwtypes.Statistic{cwtypes.StatisticSum},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get CloudWatch metrics for NAT Gateway %s: %w", natGatewayID, err)
	}

	var totalBytes float64

	for _, dp := range output.Datapoints {
		if dp.Sum != nil {
			totalBytes += *dp.Sum
		}
	}

	return totalBytes, nil
}
