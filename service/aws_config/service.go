// Package awsconfig provides a service for loading AWS configuration.
package awsconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// NewService creates a new AWS configuration service.
func NewService() *service {
	return &service{}
}

func (s *service) GetAWSCfg(ctx context.Context, region, profile string) (aws.Config, error) {
	var opts []func(*config.LoadOptions) error

	// Only set region if explicitly provided; otherwise use SDK defaults
	// (AWS_REGION, AWS_DEFAULT_REGION env vars, or ~/.aws/config)
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	// Only set profile if explicitly provided
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	return config.LoadDefaultConfig(ctx, opts...)
}
