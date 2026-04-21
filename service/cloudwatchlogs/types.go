package cloudwatchlogs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/elC0mpa/aws-doctor/model"
)

// ClientAPI is the interface for the AWS CloudWatch Logs client methods used by the service.
type ClientAPI interface {
	DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	StartQuery(ctx context.Context, params *cloudwatchlogs.StartQueryInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.StartQueryOutput, error)
	GetQueryResults(ctx context.Context, params *cloudwatchlogs.GetQueryResultsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetQueryResultsOutput, error)
}

// Service is the interface for AWS CloudWatch Logs service.
type Service interface {
	GetCloudWatchLogsWaste(ctx context.Context) ([]model.CloudWatchLogsWasteInfo, error)
	GetLambdaMaxMemoryUsedBatch(ctx context.Context, logGroupNames []string, startTime, endTime time.Time) (map[string]int32, error)
	ListExistingLogGroups(ctx context.Context, prefix string) (map[string]struct{}, error)
}

type service struct {
	client ClientAPI
}
