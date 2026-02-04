// Package awsconfig provides a service for loading AWS configuration.
package awsconfig

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// loadSharedConfigProfile is a variable to allow mocking in tests.
var loadSharedConfigProfile = config.LoadSharedConfigProfile

// NewService creates a new AWS configuration service.
func NewService() Service {
	return &service{}
}

func (s *service) GetAWSCfg(ctx context.Context, region, profile string) (aws.Config, error) {
	// Proactively check if the profile requires MFA and force manual handling if so.
	// This avoids the issue where LoadDefaultConfig returns a config that fails later
	// (SignatureDoesNotMatch) because it didn't correctly use the source profile's credentials for signing.
	if profile != "" {
		sharedCfg, err := loadSharedConfigProfile(ctx, profile)
		if err == nil && sharedCfg.RoleARN != "" && sharedCfg.MFASerial != "" {
			return s.loadConfigWithManualMFA(ctx, region, profile)
		}
	}

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

	// Force credential retrieval to ensure any authentication challenges (like MFA)
	// are handled before returning the config.
	if cfg.Credentials != nil {
		if _, err := cfg.Credentials.Retrieve(ctx); err != nil {
			return aws.Config{}, fmt.Errorf("failed to retrieve credentials: %w", err)
		}
	}

	return cfg, nil
}

// loadConfigWithManualMFA manually constructs the configuration for a profile with MFA
// when LoadDefaultConfig fails to apply the token provider.
func (s *service) loadConfigWithManualMFA(ctx context.Context, region, profile string) (aws.Config, error) {
	// 1. Load the shared config to get RoleARN and MFASerial
	sharedCfg, err := loadSharedConfigProfile(ctx, profile)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load shared config profile: %w", err)
	}

	if sharedCfg.RoleARN == "" || sharedCfg.MFASerial == "" {
		// Should not happen if the error was "AssumeRoleTokenProvider...", but good to check
		return aws.Config{}, fmt.Errorf("profile %s missing role_arn or mfa_serial", profile)
	}

	// 2. Load the base config (to get credentials for the source profile)
	sourceProfile := sharedCfg.SourceProfileName
	if sourceProfile == "" {
		sourceProfile = "default"
	}

	baseOpts := []func(*config.LoadOptions) error{
		config.WithSharedConfigProfile(sourceProfile),
	}

	// Ensure the base config has a region for the STS client
	// The STS client requires a region to resolve the endpoint for AssumeRole.
	stsRegion := region
	if stsRegion == "" {
		stsRegion = sharedCfg.Region
	}

	if stsRegion == "" {
		stsRegion = "us-east-1" // Default fallback for STS if no region is found
	}

	baseOpts = append(baseOpts, config.WithRegion(stsRegion))

	baseCfg, err := config.LoadDefaultConfig(ctx, baseOpts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load source profile config: %w", err)
	}

	// 3. Create STS client from base config
	stsClient := sts.NewFromConfig(baseCfg)

	// 4. Create AssumeRoleProvider with the MFA token provider
	provider := stscreds.NewAssumeRoleProvider(stsClient, sharedCfg.RoleARN, func(o *stscreds.AssumeRoleOptions) {
		o.SerialNumber = aws.String(sharedCfg.MFASerial)
		o.TokenProvider = stscreds.StdinTokenProvider
	})

	// 5. Create the final config
	// We start with a default config that respects the *target* region and other defaults,
	// but strictly overrides credentials.
	finalOpts := []func(*config.LoadOptions) error{
		config.WithCredentialsProvider(aws.NewCredentialsCache(provider)),
	}
	if region != "" {
		finalOpts = append(finalOpts, config.WithRegion(region))
	} else if sharedCfg.Region != "" {
		finalOpts = append(finalOpts, config.WithRegion(sharedCfg.Region))
	}

	finalCfg, err := config.LoadDefaultConfig(ctx, finalOpts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load final config with mfa: %w", err)
	}

	// Force credential retrieval to trigger MFA prompt immediately.
	// This ensures the prompt happens before any UI spinners are started.
	if finalCfg.Credentials != nil {
		if _, err := finalCfg.Credentials.Retrieve(ctx); err != nil {
			return aws.Config{}, fmt.Errorf("failed to retrieve credentials (MFA might have failed): %w", err)
		}
	}

	return finalCfg, nil
}
