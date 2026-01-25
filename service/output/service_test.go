package output

import (
	"testing"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name           string
		inputFormat    string
		expectedFormat Format
	}{
		{
			name:           "json format",
			inputFormat:    "json",
			expectedFormat: FormatJSON,
		},
		{
			name:           "table format explicit",
			inputFormat:    "table",
			expectedFormat: FormatTable,
		},
		{
			name:           "empty string defaults to table",
			inputFormat:    "",
			expectedFormat: FormatTable,
		},
		{
			name:           "unknown format defaults to table",
			inputFormat:    "unknown",
			expectedFormat: FormatTable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.inputFormat)

			// Type assert to access internal format field
			s, ok := svc.(*service)
			if !ok {
				t.Fatal("NewService did not return *service type")
			}

			if s.format != tt.expectedFormat {
				t.Errorf("expected format %q, got %q", tt.expectedFormat, s.format)
			}
		})
	}
}

func TestFormatConstants(t *testing.T) {
	if FormatTable != "table" {
		t.Errorf("FormatTable should be 'table', got %q", FormatTable)
	}

	if FormatJSON != "json" {
		t.Errorf("FormatJSON should be 'json', got %q", FormatJSON)
	}
}
