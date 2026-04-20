// Package cloudwatchlogs provides a service for interacting with AWS CloudWatch Logs.
package cloudwatchlogs

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
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

// ListExistingLogGroups returns the set of log group names matching the given name prefix. This
// is used to pre-filter before calling GetLambdaMaxMemoryUsedBatch, since a StartQuery fails the
// entire request when any named log group is missing.
func (s *service) ListExistingLogGroups(ctx context.Context, prefix string) (map[string]struct{}, error) {
	existing := make(map[string]struct{})

	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(s.client, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe log groups: %w", err)
		}

		for _, lg := range output.LogGroups {
			if lg.LogGroupName != nil {
				existing[*lg.LogGroupName] = struct{}{}
			}
		}
	}

	return existing, nil
}

// GetLambdaMaxMemoryUsedBatch runs a single CloudWatch Logs Insights query against multiple log
// groups and returns a map of log group name to the maximum memory used (in MB) within the given
// time range. Log groups with no REPORT entries are omitted from the result.
func (s *service) GetLambdaMaxMemoryUsedBatch(ctx context.Context, logGroupNames []string, startTime, endTime time.Time) (map[string]int32, error) {
	if len(logGroupNames) == 0 {
		return map[string]int32{}, nil
	}

	queryString := `filter @type = "REPORT" | stats max(@maxMemoryUsed / 1048576) as maxMemMB by @log`

	startOutput, err := s.client.StartQuery(ctx, &cloudwatchlogs.StartQueryInput{
		LogGroupNames: logGroupNames,
		StartTime:     aws.Int64(startTime.Unix()),
		EndTime:       aws.Int64(endTime.Unix()),
		QueryString:   aws.String(queryString),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start query: %w", err)
	}

	for {
		results, err := s.client.GetQueryResults(ctx, &cloudwatchlogs.GetQueryResultsInput{
			QueryId: startOutput.QueryId,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get query results: %w", err)
		}

		if results.Status == cwlogstypes.QueryStatusComplete {
			return parseMaxMemMBByGroup(results.Results), nil
		}

		if results.Status == cwlogstypes.QueryStatusFailed || results.Status == cwlogstypes.QueryStatusCancelled || results.Status == cwlogstypes.QueryStatusTimeout {
			return nil, fmt.Errorf("query %s", results.Status)
		}

		time.Sleep(queryPollInterval)
	}
}

// parseMaxMemMBByGroup extracts per-log-group maxMemMB values from CW Logs Insights results. The
// @log field is returned as "accountId:logGroupName"; this strips the account prefix.
func parseMaxMemMBByGroup(results [][]cwlogstypes.ResultField) map[string]int32 {
	out := make(map[string]int32, len(results))

	for _, row := range results {
		var (
			logGroup string
			maxMemMB int32
			hasValue bool
		)

		for _, field := range row {
			switch aws.ToString(field.Field) {
			case "@log":
				logValue := aws.ToString(field.Value)
				if idx := strings.Index(logValue, ":"); idx >= 0 {
					logGroup = logValue[idx+1:]
				} else {
					logGroup = logValue
				}
			case "maxMemMB":
				val, err := strconv.ParseFloat(aws.ToString(field.Value), 64)
				if err == nil {
					maxMemMB = int32(math.Ceil(val))
					hasValue = true
				}
			}
		}

		if logGroup != "" && hasValue {
			out[logGroup] = maxMemMB
		}
	}

	return out
}
