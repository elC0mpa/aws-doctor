package lambda

import (
	"context"
	"time"

	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/elC0mpa/aws-doctor/model"
)

// ClientAPI is the interface for the AWS Lambda client methods used by the service.
type ClientAPI interface {
	ListFunctions(ctx context.Context, params *awslambda.ListFunctionsInput, optFns ...func(*awslambda.Options)) (*awslambda.ListFunctionsOutput, error)
}

// cloudWatchLogsService defines the interface for the CloudWatch Logs dependency.
type cloudWatchLogsService interface {
	GetMaxMemoryUsed(ctx context.Context, logGroupName string, startTime, endTime time.Time) (int32, error)
}

type service struct {
	lambdaClient ClientAPI
	logsService  cloudWatchLogsService
}

// Service defines the interface for AWS Lambda waste detection.
type Service interface {
	GetOverProvisionedFunctions(ctx context.Context, memoryThresholdPercent int) ([]model.LambdaOverProvisionedInfo, error)
}
