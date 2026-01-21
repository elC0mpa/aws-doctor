package utils

import (
	"testing"
	"time"
)

func TestParseTransitionDate(t *testing.T) {
	tests := []struct {
		name      string
		reason    string
		wantTime  time.Time
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "valid_stopped_reason_UTC",
			reason:   "User initiated (2024-01-15 10:30:45 UTC)",
			wantTime: time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "valid_stopped_reason_GMT",
			reason:   "User initiated (2023-12-25 23:59:59 GMT)",
			wantTime: time.Date(2023, 12, 25, 23, 59, 59, 0, time.FixedZone("GMT", 0)),
			wantErr:  false,
		},
		{
			name:     "valid_stopped_reason_with_different_prefix",
			reason:   "Client.UserInitiatedShutdown: User initiated shutdown (2024-06-01 00:00:00 UTC)",
			wantTime: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "valid_with_nested_parentheses_takes_first",
			reason:   "Some reason (2024-03-15 12:00:00 UTC) (extra info)",
			wantTime: time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:      "empty_string",
			reason:    "",
			wantErr:   true,
			errSubstr: "no date found",
		},
		{
			name:      "no_parentheses",
			reason:    "User initiated shutdown",
			wantErr:   true,
			errSubstr: "no date found",
		},
		{
			name:      "empty_parentheses",
			reason:    "User initiated ()",
			wantErr:   true,
			errSubstr: "no date found",
		},
		{
			name:      "invalid_date_format_in_parentheses",
			reason:    "User initiated (invalid-date)",
			wantErr:   true,
			errSubstr: "parsing time",
		},
		{
			name:      "wrong_date_format",
			reason:    "User initiated (01/15/2024 10:30:45)",
			wantErr:   true,
			errSubstr: "parsing time",
		},
		{
			name:      "missing_timezone",
			reason:    "User initiated (2024-01-15 10:30:45)",
			wantErr:   true,
			errSubstr: "parsing time",
		},
		{
			name:      "date_only_no_time",
			reason:    "User initiated (2024-01-15)",
			wantErr:   true,
			errSubstr: "parsing time",
		},
		{
			name:     "boundary_date_start_of_year",
			reason:   "User initiated (2024-01-01 00:00:00 UTC)",
			wantTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "boundary_date_end_of_year",
			reason:   "User initiated (2024-12-31 23:59:59 UTC)",
			wantTime: time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "leap_year_date",
			reason:   "User initiated (2024-02-29 12:00:00 UTC)",
			wantTime: time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTransitionDate(tt.reason)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseTransitionDate() expected error containing %q, got nil", tt.errSubstr)
					return
				}
				if tt.errSubstr != "" && !contains(err.Error(), tt.errSubstr) {
					t.Errorf("ParseTransitionDate() error = %v, want error containing %q", err, tt.errSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTransitionDate() unexpected error = %v", err)
				return
			}

			if !got.Equal(tt.wantTime) {
				t.Errorf("ParseTransitionDate() = %v, want %v", got, tt.wantTime)
			}
		})
	}
}

// contains checks if substr is in s (simple helper to avoid importing strings)
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParseTransitionDate_RealWorldExamples(t *testing.T) {
	// These are actual AWS state transition reason formats
	realWorldCases := []struct {
		name   string
		reason string
	}{
		{
			name:   "user_initiated_shutdown",
			reason: "User initiated (2024-01-15 10:30:45 UTC)",
		},
		{
			name:   "client_user_initiated_shutdown",
			reason: "Client.UserInitiatedShutdown: User initiated shutdown (2024-01-15 10:30:45 UTC)",
		},
		{
			name:   "client_instance_initiated_shutdown",
			reason: "Client.InstanceInitiatedShutdown: Instance initiated shutdown (2024-01-15 10:30:45 UTC)",
		},
		{
			name:   "server_spot_instance_termination",
			reason: "Server.SpotInstanceTermination: Spot Instance termination (2024-01-15 10:30:45 UTC)",
		},
	}

	for _, tt := range realWorldCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTransitionDate(tt.reason)
			if err != nil {
				t.Errorf("ParseTransitionDate() failed for real-world case %q: %v", tt.reason, err)
				return
			}

			// Verify we got a valid time (not zero)
			if got.IsZero() {
				t.Errorf("ParseTransitionDate() returned zero time for %q", tt.reason)
			}

			// Verify the year is reasonable (2020-2030)
			if got.Year() < 2020 || got.Year() > 2030 {
				t.Errorf("ParseTransitionDate() returned unexpected year %d for %q", got.Year(), tt.reason)
			}
		})
	}
}

func BenchmarkParseTransitionDate(b *testing.B) {
	reason := "User initiated (2024-01-15 10:30:45 UTC)"
	for i := 0; i < b.N; i++ {
		_, _ = ParseTransitionDate(reason)
	}
}
