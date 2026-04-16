// Package cloudwatchlogs provides a service for interacting with AWS CloudWatch Logs.
package cloudwatchlogs

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwlogstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/pricing"
)

// NewService creates a new CloudWatch Logs service.
func NewService(awsconfig aws.Config) Service {
	client := cloudwatchlogs.NewFromConfig(awsconfig)

	return &service{
		client: client,
	}
}

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
				wasteLogGroups = append(wasteLogGroups, model.CloudWatchLogsWasteInfo{
					LogGroupName:         *logGroup.LogGroupName,
					CreationTime:         time.Unix(0, *logGroup.CreationTime*int64(time.Millisecond)),
					StoredBytes:          storedBytes,
					EstimatedMonthlyCost: pricing.CalculateCloudWatchLogsMonthlyCost(storedBytes),
				})
			}
		}
	}

	return wasteLogGroups, nil
}

const queryPollInterval = 500 * time.Millisecond

// GetLambdaMaxMemoryUsed runs a CloudWatch Logs Insights query to find the maximum memory used
// by a Lambda function within the given time range. Returns the value in MB.
func (s *service) GetLambdaMaxMemoryUsed(ctx context.Context, logGroupName string, startTime, endTime time.Time) (int32, error) {
	queryString := `filter @type = "REPORT" | stats max(@maxMemoryUsed / 1048576) as maxMemMB`

	startOutput, err := s.client.StartQuery(ctx, &cloudwatchlogs.StartQueryInput{
		LogGroupName: aws.String(logGroupName),
		StartTime:    aws.Int64(startTime.Unix()),
		EndTime:      aws.Int64(endTime.Unix()),
		QueryString:  aws.String(queryString),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to start query for %s: %w", logGroupName, err)
	}

	for {
		results, err := s.client.GetQueryResults(ctx, &cloudwatchlogs.GetQueryResultsInput{
			QueryId: startOutput.QueryId,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to get query results for %s: %w", logGroupName, err)
		}

		if results.Status == cwlogstypes.QueryStatusComplete {
			return parseMaxMemMB(results.Results), nil
		}

		if results.Status == cwlogstypes.QueryStatusFailed || results.Status == cwlogstypes.QueryStatusCancelled || results.Status == cwlogstypes.QueryStatusTimeout {
			return 0, fmt.Errorf("query %s for %s", results.Status, logGroupName)
		}

		time.Sleep(queryPollInterval)
	}
}

// parseMaxMemMB extracts the maxMemMB value from CW Logs Insights query results.
func parseMaxMemMB(results [][]cwlogstypes.ResultField) int32 {
	for _, row := range results {
		for _, field := range row {
			if aws.ToString(field.Field) == "maxMemMB" {
				val, err := strconv.ParseFloat(aws.ToString(field.Value), 64)
				if err != nil {
					return 0
				}

				return int32(math.Ceil(val))
			}
		}
	}

	return 0
}
