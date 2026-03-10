package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		val      string
		expected bool
	}{
		{
			name:     "exact match",
			slice:    []string{"ec2", "s3", "elb"},
			val:      "s3",
			expected: true,
		},
		{
			name:     "case insensitive match",
			slice:    []string{"EC2", "S3", "ELB"},
			val:      "s3",
			expected: true,
		},
		{
			name:     "match with whitespace",
			slice:    []string{" ec2 ", "s3"},
			val:      "ec2",
			expected: true,
		},
		{
			name:     "no match",
			slice:    []string{"ec2", "s3"},
			val:      "lambda",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			val:      "ec2",
			expected: false,
		},
		{
			name:     "substring not matched",
			slice:    []string{"ec2-instance"},
			val:      "ec2",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsIgnoreCase(tt.slice, tt.val)
			assert.Equal(t, tt.expected, result)
		})
	}
}
