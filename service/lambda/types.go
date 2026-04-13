package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/elC0mpa/aws-doctor/model"
)

// ClientAPI is the interface for the AWS Lambda client methods used by the service.
type ClientAPI interface {
	ListFunctions(ctx context.Context, params *awslambda.ListFunctionsInput, optFns ...func(*awslambda.Options)) (*awslambda.ListFunctionsOutput, error)
}

// LogsClientAPI is the interface for the CloudWatch Logs client methods used by the service.
type LogsClientAPI interface {
	FilterLogEvents(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error)
}

type service struct {
	lambdaClient ClientAPI
	logsClient   LogsClientAPI
}

// Service defines the interface for AWS Lambda waste detection.
type Service interface {
	GetOverProvisionedFunctions(ctx context.Context, memoryThresholdPercent int) ([]model.LambdaOverProvisionedInfo, error)
}
