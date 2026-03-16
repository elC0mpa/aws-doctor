// Package cloudwatchlogs provides a service for interacting with AWS CloudWatch Logs.
package cloudwatchlogs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/elC0mpa/aws-doctor/model"
)

// NewService creates a new CloudWatch Logs service.
func NewService(awsconfig aws.Config) Service {
	client := cloudwatchlogs.NewFromConfig(awsconfig)

	return &service{
		client: client,
	}
}

// CloudWatch Logs storage pricing: ~$0.03 per GB per month
const cloudwatchStorageCostPerGBMonth = 0.03

// GetCloudWatchLogsWaste returns a list of CloudWatch Log Groups without a retention policy.
func (s *service) GetCloudWatchLogsWaste(ctx context.Context) ([]model.CloudWatchLogsWasteInfo, error) {
	var wasteLogGroups []model.CloudWatchLogsWasteInfo

	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(s.client, &cloudwatchlogs.DescribeLogGroupsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe log groups: %w", err)
		}

		for _, logGroup := range output.LogGroups {
			if logGroup.RetentionInDays == nil {
				storedBytes := aws.ToInt64(logGroup.StoredBytes)
				storedGB := float64(storedBytes) / (1024 * 1024 * 1024)
				wasteLogGroups = append(wasteLogGroups, model.CloudWatchLogsWasteInfo{
					LogGroupName:         *logGroup.LogGroupName,
					CreationTime:         time.Unix(0, *logGroup.CreationTime*int64(time.Millisecond)),
					StoredBytes:          storedBytes,
					EstimatedMonthlyCost: storedGB * cloudwatchStorageCostPerGBMonth,
				})
			}
		}
	}

	return wasteLogGroups, nil
}
