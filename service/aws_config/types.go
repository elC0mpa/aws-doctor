package awsconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type service struct{}

// ConfigService is the interface for AWS configuration service.
type ConfigService interface {
	LoadDefaultConfig(ctx context.Context, region string, profile string) (aws.Config, error)
}
