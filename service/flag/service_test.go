package flag

import (
	"testing"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
)

func TestGetParsedFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    model.Flags
		wantErr bool
	}{
		{
			name: "parses update and region",
			args: []string{"-update", "-region", "us-east-1"},
			want: model.Flags{
				Region: "us-east-1",
				Output: "table",
				Update: true,
			},
		},
		{
			name: "parses all supported flags",
			args: []string{"-region", "eu-west-1", "-profile", "dev", "-trend", "-waste", "-output", "json", "-version"},
			want: model.Flags{
				Region:  "eu-west-1",
				Profile: "dev",
				Trend:   true,
				Waste:   true,
				Output:  "json",
				Version: true,
			},
		},
		{
			name: "uses defaults when no args provided",
			args: []string{},
			want: model.Flags{
				Output: "table",
			},
		},
		{
			name:    "returns error on unknown flag",
			args:    []string{"-unknown-flag"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService()
			parsedFlags, err := svc.GetParsedFlags(tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, parsedFlags)
		})
	}
}
