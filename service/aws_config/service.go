// Package awsconfig provides a service for loading AWS configuration.
package awsconfig

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
)

// NewService creates a new AWS configuration service.
func NewService() Service {
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

	// Provide MFA token provider for profiles that use assume role with MFA.
	// This prompts the user to enter their MFA code when required.
	opts = append(opts, config.WithAssumeRoleCredentialOptions(func(options *stscreds.AssumeRoleOptions) {
		options.TokenProvider = stscreds.StdinTokenProvider
	}))

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to load AWS config: %w", err)
	}

	return cfg, nil
}
