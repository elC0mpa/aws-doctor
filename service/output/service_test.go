package output

import (
	"fmt"
	"testing"
)

func TestNewRenderer(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected interface{}
	}{
		{
			name:     "JSON format returns jsonRenderer",
			format:   "json",
			expected: &jsonRenderer{},
		},
		{
			name:     "CSV format returns csvRenderer",
			format:   "csv",
			expected: &csvRenderer{},
		},
		{
			name:     "Table format returns tableRenderer",
			format:   "table",
			expected: &tableRenderer{},
		},
		{
			name:     "Empty format returns tableRenderer default",
			format:   "",
			expected: &tableRenderer{},
		},
		{
			name:     "Unknown format returns tableRenderer default",
			format:   "unknown",
			expected: &tableRenderer{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewRenderer(tt.format)
			if renderer == nil {
				t.Fatalf("Expected a renderer, got nil")
			}
			
			// Compare types
			expectedType := fmt.Sprintf("%T", tt.expected)
			actualType := fmt.Sprintf("%T", renderer)
			if actualType != expectedType {
				t.Errorf("Expected type %s, got %s", expectedType, actualType)
			}
		})
	}
}
