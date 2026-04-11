package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEqual(t *testing.T) {
	tests := []struct {
		name         string
		latestTag    string
		localVersion string
		expected     bool
	}{
		{
			name:         "exactly equal",
			latestTag:    "v1.2.3",
			localVersion: "v1.2.3",
			expected:     true,
		},
		{
			name:         "tag has v prefix",
			latestTag:    "v1.2.3",
			localVersion: "1.2.3",
			expected:     true,
		},
		{
			name:         "not equal",
			latestTag:    "v1.2.4",
			localVersion: "1.2.3",
			expected:     false,
		},
		{
			name:         "empty latest tag",
			latestTag:    "",
			localVersion: "1.2.3",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEqual(tt.latestTag, tt.localVersion)
			assert.Equal(t, tt.expected, result)
		})
	}
}
