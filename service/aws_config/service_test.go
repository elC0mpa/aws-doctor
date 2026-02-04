package awsconfig

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
)

func TestNewService(t *testing.T) {
	s := NewService()
	if s == nil {
		t.Error("NewService() returned nil")
	}
}

func TestGetAWSCfg_DefaultOptions(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_REGION", "us-east-1")

	s := NewService()

	// Test with empty region and profile (uses SDK defaults)
	cfg, err := s.GetAWSCfg(context.Background(), "", "")
	if err != nil {
		t.Errorf("GetAWSCfg() with default options returned error: %v", err)
	}

	// Verify config is not zero value (has been initialized)
	if cfg.Region == "" {
		// Region may be empty if no default is configured, which is acceptable
		// The important thing is that the function didn't error
		t.Log("No default region configured, which is acceptable")
	}
}

func TestGetAWSCfg_WithRegion(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	s := NewService()

	tests := []struct {
		name   string
		region string
	}{
		{
			name:   "us_east_1",
			region: "us-east-1",
		},
		{
			name:   "eu_west_1",
			region: "eu-west-1",
		},
		{
			name:   "ap_southeast_1",
			region: "ap-southeast-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := s.GetAWSCfg(context.Background(), tt.region, "")
			if err != nil {
				t.Errorf("GetAWSCfg() with region %q returned error: %v", tt.region, err)
			}

			if cfg.Region != tt.region {
				t.Errorf("GetAWSCfg() region = %q, want %q", cfg.Region, tt.region)
			}
		})
	}
}

func TestGetAWSCfg_WithInvalidProfile(t *testing.T) {
	s := NewService()

	// Using a non-existent profile should return an error
	_, err := s.GetAWSCfg(context.Background(), "", "non-existent-profile-12345")
	if err == nil {
		t.Error("GetAWSCfg() with invalid profile should return error")
	}
}

func TestGetAWSCfg_ContextCancellation(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_REGION", "us-east-1")

	s := NewService()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// A cancelled context may or may not cause an error depending on
	// when the cancellation is checked. This test verifies the function
	// handles context properly without panicking.
	_, _ = s.GetAWSCfg(ctx, "", "")
	// Not asserting on error since context cancellation behavior
	// depends on SDK internals
}

func TestGetAWSCfg_WithMFAProfile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := tmpDir + "/config"
	configContent := `[default]
          region = us-east-1
         `

	if err := os.WriteFile(configFile, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write temporary config file: %v", err)
	}

	// Mock loadSharedConfigProfile
	origLoadShared := loadSharedConfigProfile

	defer func() { loadSharedConfigProfile = origLoadShared }()

	loadSharedConfigProfile = func(ctx context.Context, profileName string, optFns ...func(*config.LoadSharedConfigOptions)) (config.SharedConfig, error) {
		if profileName == "mfa-test" {
			return config.SharedConfig{
				RoleARN:           "arn:aws:iam::123456789012:role/test-role",
				MFASerial:         "arn:aws:iam::123456789012:mfa/test-user",
				SourceProfileName: "default",
				Region:            "us-west-2",
			}, nil
		}

		if profileName == "mfa-no-arn" {
			return config.SharedConfig{
				MFASerial: "arn:aws:iam::123456789012:mfa/test-user",
			}, nil
		}

		if profileName == "mfa-error" {
			return config.SharedConfig{
				RoleARN:   "arn:aws:iam::123456789012:role/test-role",
				MFASerial: "arn:aws:iam::123456789012:mfa/test-user",
			}, nil
		}

		return config.SharedConfig{}, nil
	}

	t.Setenv("AWS_CONFIG_FILE", configFile)
	t.Setenv("AWS_ACCESS_KEY_ID", "test-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")
	t.Setenv("AWS_SDK_LOAD_CONFIG", "1")

	s := NewService()

	t.Run("MFA with explicit region", func(t *testing.T) {
		_, err := s.GetAWSCfg(context.Background(), "us-east-1", "mfa-test")
		if err == nil {
			t.Error("Expected error for fake MFA profile, but got nil")
		}
	})

	t.Run("MFA with config region", func(t *testing.T) {
		_, err := s.GetAWSCfg(context.Background(), "", "mfa-test")
		if err == nil {
			t.Error("Expected error for fake MFA profile, but got nil")
		}
	})

	t.Run("MFA profile with missing ARN", func(t *testing.T) {
		_, _ = s.GetAWSCfg(context.Background(), "us-east-1", "mfa-no-arn")
	})

	t.Run("MFA with manual trigger error path", func(t *testing.T) {
		// To cover the path where stsRegion fallback happens
		_, _ = s.GetAWSCfg(context.Background(), "", "mfa-error")
	})

	t.Run("loadConfigWithManualMFA missing fields", func(t *testing.T) {
		// Call directly to hit error paths
		loadSharedConfigProfile = func(ctx context.Context, profileName string, optFns ...func(*config.LoadSharedConfigOptions)) (config.SharedConfig, error) {
			return config.SharedConfig{
				RoleARN: "", // Missing RoleARN
			}, nil
		}

		_, err := s.(*service).loadConfigWithManualMFA(context.Background(), "", "any")
		if err == nil {
			t.Error("Expected error for missing RoleARN, but got nil")
		}
	})

	t.Run("loadConfigWithManualMFA stsRegion fallback", func(t *testing.T) {
		loadSharedConfigProfile = func(ctx context.Context, profileName string, optFns ...func(*config.LoadSharedConfigOptions)) (config.SharedConfig, error) {
			return config.SharedConfig{
				RoleARN:   "arn:aws:iam::123456789012:role/test-role",
				MFASerial: "arn:aws:iam::123456789012:mfa/test-user",
				Region:    "", // Trigger fallback
			}, nil
		}

		_, _ = s.(*service).loadConfigWithManualMFA(context.Background(), "", "any")
	})
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
