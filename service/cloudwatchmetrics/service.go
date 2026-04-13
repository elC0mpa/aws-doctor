// Package cloudwatchmetrics provides a service for querying CloudWatch metrics.
package cloudwatchmetrics

import (
	"context"
	"fmt"
	"strings"
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

// ELBHasZeroRequestsInPeriod checks if a load balancer had zero requests/connections over the given number of days.
func (s *service) ELBHasZeroRequestsInPeriod(ctx context.Context, loadBalancerArn string, lbType string, days int) (bool, error) {
	now := time.Now()
	startTime := now.AddDate(0, 0, -days)

	// Extract the LB ID from the ARN (the part after "loadbalancer/")
	parts := strings.SplitN(loadBalancerArn, ":loadbalancer/", 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid load balancer ARN: %s", loadBalancerArn)
	}

	lbID := parts[1]

	var namespace, metricName string

	switch lbType {
	case "application":
		namespace = "AWS/ApplicationELB"
		metricName = "RequestCount"
	case "network":
		namespace = "AWS/NetworkELB"
		metricName = "ActiveFlowCount"
	default:
		return false, fmt.Errorf("unsupported load balancer type: %s", lbType)
	}

	output, err := s.client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(namespace),
		MetricName: aws.String(metricName),
		Dimensions: []cwtypes.Dimension{
			{
				Name:  aws.String("LoadBalancer"),
				Value: aws.String(lbID),
			},
		},
		StartTime:  &startTime,
		EndTime:    &now,
		Period:     aws.Int32(metricPeriodSeconds),
		Statistics: []cwtypes.Statistic{cwtypes.StatisticSum},
	})
	if err != nil {
		return false, fmt.Errorf("failed to get CloudWatch metrics for %s: %w", loadBalancerArn, err)
	}

	for _, dp := range output.Datapoints {
		if dp.Sum != nil && *dp.Sum > 0 {
			return false, nil
		}
	}

	return true, nil
}
