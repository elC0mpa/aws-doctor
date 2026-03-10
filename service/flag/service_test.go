package flag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetParsedFlags(t *testing.T) {
	svc := NewService()

	tests := []struct {
		name       string
		args       []string
		wantWaste  bool
		wantChecks []string
		wantUpdate bool
		wantRegion string
		wantErr    bool
	}{
		{
			name:       "just update and region",
			args:       []string{"-update", "-region", "us-east-1"},
			wantUpdate: true,
			wantRegion: "us-east-1",
			wantWaste:  false,
			wantChecks: nil,
		},
		{
			name:       "just waste flag",
			args:       []string{"--waste"},
			wantWaste:  true,
			wantChecks: nil,
		},
		{
			name:       "waste with single check",
			args:       []string{"--waste", "ec2"},
			wantWaste:  true,
			wantChecks: []string{"ec2"},
		},
		{
			name:       "waste with multiple checks",
			args:       []string{"--waste", "ec2,s3,elb"},
			wantWaste:  true,
			wantChecks: []string{"ec2", "s3", "elb"},
		},
		{
			name:       "waste with checks and other flags",
			args:       []string{"--waste", "s3", "--region", "us-west-2", "--profile", "dev"},
			wantWaste:  true,
			wantChecks: []string{"s3"},
			wantRegion: "us-west-2",
		},
		{
			name:       "waste flag followed by another flag should not be treated as a check",
			args:       []string{"--waste", "--region", "us-east-1"},
			wantWaste:  true,
			wantChecks: nil,
			wantRegion: "us-east-1",
		},
		{
			name:       "short waste flag with checks",
			args:       []string{"-waste", "ec2,s3"},
			wantWaste:  true,
			wantChecks: []string{"ec2", "s3"},
		},
		{
			name:    "invalid flag",
			args:    []string{"--invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, err := svc.GetParsedFlags(tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantWaste, flags.Waste)
			assert.Equal(t, tt.wantChecks, flags.WasteChecks)
			assert.Equal(t, tt.wantUpdate, flags.Update)
			assert.Equal(t, tt.wantRegion, flags.Region)
		})
	}
}
