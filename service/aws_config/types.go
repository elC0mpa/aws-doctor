package awsconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type service struct{}

// Service is the interface for AWS configuration service.
type Service interface {
	GetAWSCfg(ctx context.Context, region string, profile string) (aws.Config, error)
}
