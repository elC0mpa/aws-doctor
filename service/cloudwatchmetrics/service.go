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
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
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

// NATGatewayBytesOut returns the total bytes out to destination for a NAT Gateway over the given number of days.
func (s *service) NATGatewayBytesOut(ctx context.Context, natGatewayID string, days int) (float64, error) {
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

// ExtractLoadBalancerID extracts the CloudWatch dimension value from a load balancer ARN.
func ExtractLoadBalancerID(arn string) (string, error) {
	parts := strings.SplitN(arn, ":loadbalancer/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid load balancer ARN: %s", arn)
	}

	return parts[1], nil
}

// ELBHasZeroRequestsInPeriod checks if a load balancer had zero requests/connections over the given number of days.
func (s *service) ELBHasZeroRequestsInPeriod(ctx context.Context, loadBalancerArn string, lbType elbtypes.LoadBalancerTypeEnum, days int) (bool, error) {
	now := time.Now()
	startTime := now.AddDate(0, 0, -days)

	lbID, err := ExtractLoadBalancerID(loadBalancerArn)
	if err != nil {
		return false, err
	}

	var namespace, metricName string

	switch lbType {
	case elbtypes.LoadBalancerTypeEnumApplication:
		namespace = "AWS/ApplicationELB"
		metricName = "RequestCount"
	case elbtypes.LoadBalancerTypeEnumNetwork:
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

// EC2InstanceIdleStats returns the average CPU utilization (percent) and the average daily
// network throughput (NetworkIn + NetworkOut bytes per day) for the given EC2 instance over
// the lookback window. Callers compare these against their own thresholds to decide whether
// the instance is idle.
func (s *service) EC2InstanceIdleStats(ctx context.Context, instanceID string, days int) (float64, float64, error) {
	now := time.Now()
	startTime := now.AddDate(0, 0, -days)

	output, err := s.client.GetMetricData(ctx, &cloudwatch.GetMetricDataInput{
		MetricDataQueries: []cwtypes.MetricDataQuery{
			{
				Id: aws.String("cpu"),
				MetricStat: &cwtypes.MetricStat{
					Metric: &cwtypes.Metric{
						Namespace:  aws.String("AWS/EC2"),
						MetricName: aws.String("CPUUtilization"),
						Dimensions: []cwtypes.Dimension{
							{Name: aws.String("InstanceId"), Value: aws.String(instanceID)},
						},
					},
					Period: aws.Int32(metricPeriodSeconds),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("net_in"),
				MetricStat: &cwtypes.MetricStat{
					Metric: &cwtypes.Metric{
						Namespace:  aws.String("AWS/EC2"),
						MetricName: aws.String("NetworkIn"),
						Dimensions: []cwtypes.Dimension{
							{Name: aws.String("InstanceId"), Value: aws.String(instanceID)},
						},
					},
					Period: aws.Int32(metricPeriodSeconds),
					Stat:   aws.String("Sum"),
				},
			},
			{
				Id: aws.String("net_out"),
				MetricStat: &cwtypes.MetricStat{
					Metric: &cwtypes.Metric{
						Namespace:  aws.String("AWS/EC2"),
						MetricName: aws.String("NetworkOut"),
						Dimensions: []cwtypes.Dimension{
							{Name: aws.String("InstanceId"), Value: aws.String(instanceID)},
						},
					},
					Period: aws.Int32(metricPeriodSeconds),
					Stat:   aws.String("Sum"),
				},
			},
		},
		StartTime: &startTime,
		EndTime:   &now,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get CloudWatch metric data for %s: %w", instanceID, err)
	}

	var (
		cpuAvgPercent     float64
		networkBytesTotal float64
	)

	for _, result := range output.MetricDataResults {
		switch aws.ToString(result.Id) {
		case "cpu":
			var total float64
			for _, val := range result.Values {
				total += val
			}
			if n := len(result.Values); n > 0 {
				cpuAvgPercent = total / float64(n)
			}
		case "net_in", "net_out":
			for _, val := range result.Values {
				networkBytesTotal += val
			}
		}
	}

	if days <= 0 {
		return cpuAvgPercent, networkBytesTotal, nil
	}

	return cpuAvgPercent, networkBytesTotal / float64(days), nil
}

func averageOfAverages(datapoints []cwtypes.Datapoint) float64 {
	if len(datapoints) == 0 {
		return 0
	}

	var (
		total float64
		count int
	)

	for _, dp := range datapoints {
		if dp.Average != nil {
			total += *dp.Average
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / float64(count)
}

// SageMakerVariantInvocations returns the total Invocations for a SageMaker endpoint production
// variant over the given number of days. The CloudWatch metric is published per (EndpointName,
// VariantName) pair, so callers that need an endpoint-wide total must sum across variants.
func (s *service) SageMakerVariantInvocations(ctx context.Context, endpointName, variantName string, days int) (float64, error) {
	now := time.Now()
	startTime := now.AddDate(0, 0, -days)

	output, err := s.client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/SageMaker"),
		MetricName: aws.String("Invocations"),
		Dimensions: []cwtypes.Dimension{
			{Name: aws.String("EndpointName"), Value: aws.String(endpointName)},
			{Name: aws.String("VariantName"), Value: aws.String(variantName)},
		},
		StartTime:  &startTime,
		EndTime:    &now,
		Period:     aws.Int32(metricPeriodSeconds),
		Statistics: []cwtypes.Statistic{cwtypes.StatisticSum},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get SageMaker invocations for %s/%s: %w", endpointName, variantName, err)
	}

	var total float64

	for _, dp := range output.Datapoints {
		if dp.Sum != nil {
			total += *dp.Sum
		}
	}

	return total, nil
}
