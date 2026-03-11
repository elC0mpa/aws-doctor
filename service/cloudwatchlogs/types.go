package cloudwatchlogs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/elC0mpa/aws-doctor/model"
)

// ClientAPI is the interface for the AWS CloudWatch Logs client methods used by the service.
type ClientAPI interface {
	DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
}

// Service is the interface for AWS CloudWatch Logs service.
type Service interface {
	GetCloudWatchLogsWaste(ctx context.Context) ([]model.CloudWatchLogsWasteInfo, error)
}

type service struct {
	client ClientAPI
}
