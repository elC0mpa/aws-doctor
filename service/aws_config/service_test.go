package awsconfig

import (
	"context"
	"testing"
)

func TestNewService(t *testing.T) {
	s := NewService()
	if s == nil {
		t.Error("NewService() returned nil")
	}
}

func TestGetAWSCfg_DefaultOptions(t *testing.T) {
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

func TestConfigService_Interface(t *testing.T) {
	// Verify that *service implements ConfigService interface
	var _ ConfigService = NewService()
}
