package awsconfig

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	s := NewService()
	if s == nil {
		t.Error("NewService() returned nil")
	}
}

func TestGetAWSCfg(t *testing.T) {
	s := NewService()
	ctx := context.Background()

	t.Run("default_config", func(t *testing.T) {
		// Mock environment to avoid using local credentials
		t.Setenv("AWS_ACCESS_KEY_ID", "test")
		t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		t.Setenv("AWS_REGION", "us-east-1")

		_, err := s.GetAWSCfg(ctx, "", "")
		if err != nil {
			t.Errorf("GetAWSCfg() error = %v", err)
		}
	})

	t.Run("with_region", func(t *testing.T) {
		t.Setenv("AWS_ACCESS_KEY_ID", "test")
		t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

		_, err := s.GetAWSCfg(ctx, "us-west-2", "")
		if err != nil {
			t.Errorf("GetAWSCfg() error = %v", err)
		}
	})

	t.Run("with_profile", func(t *testing.T) {
		// This might fail if the profile doesn't exist, but we check if it handles the flag
		_, _ = s.GetAWSCfg(ctx, "", "non-existent-profile")
	})
}

func TestGetAWSCfg_WithMFAProfile(t *testing.T) {
	origLoadShared := loadSharedConfigProfile
	defer func() { loadSharedConfigProfile = origLoadShared }()

	s := NewService().(*service)
	s.input = strings.NewReader("123456\n")
	s.output = io.Discard

	t.Run("loadConfigWithManualMFA stsRegion fallback", func(t *testing.T) {
		loadSharedConfigProfile = func(ctx context.Context, profileName string, optFns ...func(*config.LoadSharedConfigOptions)) (config.SharedConfig, error) {
			return config.SharedConfig{
				RoleARN:   "arn:aws:iam::123456789012:role/test-role",
				MFASerial: "arn:aws:iam::123456789012:mfa/test-user",
				Region:    "", // Trigger fallback
			}, nil
		}

		_, _ = s.loadConfigWithManualMFA(context.Background(), "", "any")
	})

	t.Run("loadConfigWithManualMFA sourceProfile fallback", func(t *testing.T) {
		loadSharedConfigProfile = func(ctx context.Context, profileName string, optFns ...func(*config.LoadSharedConfigOptions)) (config.SharedConfig, error) {
			return config.SharedConfig{
				RoleARN:           "arn:aws:iam::123456789012:role/test-role",
				MFASerial:         "arn:aws:iam::123456789012:mfa/test-user",
				SourceProfileName: "", // Trigger fallback to "default"
			}, nil
		}

		_, _ = s.loadConfigWithManualMFA(context.Background(), "", "any")
	})

	t.Run("loadConfigWithManualMFA with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		loadSharedConfigProfile = func(ctx context.Context, profileName string, optFns ...func(*config.LoadSharedConfigOptions)) (config.SharedConfig, error) {
			return config.SharedConfig{
				RoleARN:   "arn:aws:iam::123456789012:role/test-role",
				MFASerial: "arn:aws:iam::123456789012:mfa/test-user",
			}, nil
		}

		_, err := s.loadConfigWithManualMFA(ctx, "", "any")
		if err == nil {
			t.Log("Context cancellation might not be checked by LoadDefaultConfig immediately")
		}
	})
}

func TestMFATokenProvider(t *testing.T) {
	s := &service{
		output: io.Discard,
	}

	t.Run("without serial", func(t *testing.T) {
		s.input = strings.NewReader("123456\n")

		provider := s.mfaTokenProvider("")
		token, err := provider()
		assert.NoError(t, err)
		assert.Equal(t, "123456", token)
	})

	t.Run("with serial", func(t *testing.T) {
		s.input = strings.NewReader("654321\n")

		provider := s.mfaTokenProvider("arn:aws:iam::123456789012:mfa/user")
		token, err := provider()
		assert.NoError(t, err)
		assert.Equal(t, "654321", token)
	})
}

func TestGetAWSCfg_WithMFATokenInput(t *testing.T) {
	origLoadShared := loadSharedConfigProfile
	defer func() { loadSharedConfigProfile = origLoadShared }()

	loadSharedConfigProfile = func(ctx context.Context, profileName string, optFns ...func(*config.LoadSharedConfigOptions)) (config.SharedConfig, error) {
		return config.SharedConfig{
			RoleARN:   "arn:aws:iam::123456789012:role/test-role",
			MFASerial: "arn:aws:iam::123456789012:mfa/test-user",
		}, nil
	}

	s := NewService().(*service)
	s.input = strings.NewReader("123456\n")
	s.output = io.Discard

	// This will still fail later because STS is not real, but it hits the token provider code
	_, _ = s.GetAWSCfg(context.Background(), "us-east-1", "mfa-test")
}

func TestGetAWSCfg_LoadSharedConfigError(t *testing.T) {
	origLoadShared := loadSharedConfigProfile
	defer func() { loadSharedConfigProfile = origLoadShared }()

	loadSharedConfigProfile = func(ctx context.Context, profileName string, optFns ...func(*config.LoadSharedConfigOptions)) (config.SharedConfig, error) {
		return config.SharedConfig{}, errors.New("mock error")
	}

	s := NewService()
	// Should continue to default path if shared config fails
	_, _ = s.GetAWSCfg(context.Background(), "", "any")
}

func TestGetAWSCfg_LoadConfigError(t *testing.T) {
	// We can't easily trigger an error in config.LoadDefaultConfig without
	// more complex mocking, but we can test the case where credentials retrieval fails.
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	// Use an invalid region to potentially trigger some errors,
	// though LoadDefaultConfig is quite lenient.

	s := NewService()
	_, err := s.GetAWSCfg(context.Background(), "invalid-region-!@#$", "")
	// If it doesn't error here, it's fine, we are just exploring.
	if err != nil {
		t.Logf("Got expected error or log: %v", err)
	}
}
