package awsconfig

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type service struct {
	input  io.Reader
	output io.Writer
}

// Service is the interface for AWS configuration service.
type Service interface {
	GetAWSCfg(ctx context.Context, region string, profile string) (aws.Config, error)
}
