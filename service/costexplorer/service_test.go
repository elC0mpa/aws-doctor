package awscostexplorer

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

const costsAggregation = "UnblendedCost"

func TestGetFirstDayOfMonth(t *testing.T) {
	s := &service{}

	tests := []struct {
		name  string
		input time.Time
		want  time.Time
	}{
		{
			name:  "mid_month",
			input: time.Date(2024, 1, 15, 10, 30, 45, 123, time.UTC),
			want:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "first_day_of_month",
			input: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			want:  time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "last_day_of_month",
			input: time.Date(2024, 1, 31, 23, 59, 59, 999, time.UTC),
			want:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "february_leap_year",
			input: time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
			want:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "february_non_leap_year",
			input: time.Date(2023, 2, 28, 12, 0, 0, 0, time.UTC),
			want:  time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "december",
			input: time.Date(2024, 12, 25, 10, 0, 0, 0, time.UTC),
			want:  time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "january_new_year",
			input: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			want:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "preserves_timezone",
			input: time.Date(2024, 6, 15, 10, 0, 0, 0, time.FixedZone("EST", -5*3600)),
			want:  time.Date(2024, 6, 1, 0, 0, 0, 0, time.FixedZone("EST", -5*3600)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.getFirstDayOfMonth(tt.input)
			if !got.Equal(tt.want) {
				t.Errorf("getFirstDayOfMonth(%v) = %v, want %v", tt.input, got, tt.want)
			}

			// Verify it's always day 1
			if got.Day() != 1 {
				t.Errorf("getFirstDayOfMonth(%v) returned day %d, want 1", tt.input, got.Day())
			}

			// Verify time is zeroed
			if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
				t.Errorf("getFirstDayOfMonth(%v) time not zeroed: %v", tt.input, got)
			}
		})
	}
}

func TestGetLastDayOfMonth(t *testing.T) {
	s := &service{}

	tests := []struct {
		name  string
		input time.Time
		want  time.Time
	}{
		{
			name:  "january_31_days",
			input: time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			want:  time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "april_30_days",
			input: time.Date(2024, 4, 10, 0, 0, 0, 0, time.UTC),
			want:  time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "february_leap_year_29_days",
			input: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			want:  time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "february_non_leap_year_28_days",
			input: time.Date(2023, 2, 15, 0, 0, 0, 0, time.UTC),
			want:  time.Date(2023, 2, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "december_year_boundary",
			input: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			want:  time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "preserves_timezone",
			input: time.Date(2024, 6, 15, 10, 0, 0, 0, time.FixedZone("PST", -8*3600)),
			want:  time.Date(2024, 6, 30, 0, 0, 0, 0, time.FixedZone("PST", -8*3600)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.getLastDayOfMonth(tt.input)
			if !got.Equal(tt.want) {
				t.Errorf("getLastDayOfMonth(%v) = %v, want %v", tt.input, got, tt.want)
			}

			// Verify month is same as input
			if got.Month() != tt.input.Month() {
				t.Errorf("getLastDayOfMonth(%v) changed month from %v to %v",
					tt.input, tt.input.Month(), got.Month())
			}
		})
	}
}

func TestFilterGroups(t *testing.T) {
	s := &service{}

	tests := []struct {
		name   string
		groups []types.Group
		want   map[string]float64 // service name -> expected amount
	}{
		{
			name:   "empty_groups",
			groups: []types.Group{},
			want:   map[string]float64{},
		},
		{
			name: "single_group_with_cost",
			groups: []types.Group{
				{
					Keys: []string{"Amazon EC2"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {
							Amount: aws.String("100.50"),
							Unit:   aws.String("USD"),
						},
					},
				},
			},
			want: map[string]float64{
				"Amazon EC2": 100.50,
			},
		},
		{
			name: "multiple_groups",
			groups: []types.Group{
				{
					Keys: []string{"Amazon EC2"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {Amount: aws.String("100.00"), Unit: aws.String("USD")},
					},
				},
				{
					Keys: []string{"Amazon S3"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {Amount: aws.String("50.25"), Unit: aws.String("USD")},
					},
				},
				{
					Keys: []string{"AWS Lambda"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {Amount: aws.String("25.75"), Unit: aws.String("USD")},
					},
				},
			},
			want: map[string]float64{
				"Amazon EC2": 100.00,
				"Amazon S3":  50.25,
				"AWS Lambda": 25.75,
			},
		},
		{
			name: "filters_zero_cost",
			groups: []types.Group{
				{
					Keys: []string{"Amazon EC2"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {Amount: aws.String("100.00"), Unit: aws.String("USD")},
					},
				},
				{
					Keys: []string{"Free Service"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {Amount: aws.String("0"), Unit: aws.String("USD")},
					},
				},
				{
					Keys: []string{"Another Free"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {Amount: aws.String("0.00"), Unit: aws.String("USD")},
					},
				},
			},
			want: map[string]float64{
				"Amazon EC2": 100.00,
			},
		},
		{
			name: "filters_nil_amount",
			groups: []types.Group{
				{
					Keys: []string{"Amazon EC2"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {Amount: aws.String("100.00"), Unit: aws.String("USD")},
					},
				},
				{
					Keys: []string{"Nil Amount Service"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {Amount: nil, Unit: aws.String("USD")},
					},
				},
			},
			want: map[string]float64{
				"Amazon EC2": 100.00,
			},
		},
		{
			name: "filters_invalid_amount",
			groups: []types.Group{
				{
					Keys: []string{"Amazon EC2"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {Amount: aws.String("100.00"), Unit: aws.String("USD")},
					},
				},
				{
					Keys: []string{"Invalid Amount"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {Amount: aws.String("not-a-number"), Unit: aws.String("USD")},
					},
				},
			},
			want: map[string]float64{
				"Amazon EC2": 100.00,
			},
		},
		{
			name: "filters_missing_metric",
			groups: []types.Group{
				{
					Keys: []string{"Amazon EC2"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {Amount: aws.String("100.00"), Unit: aws.String("USD")},
					},
				},
				{
					Keys:    []string{"Missing Metric"},
					Metrics: map[string]types.MetricValue{},
				},
			},
			want: map[string]float64{
				"Amazon EC2": 100.00,
			},
		},
		{
			name: "handles_very_small_amounts",
			groups: []types.Group{
				{
					Keys: []string{"Micro Service"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {Amount: aws.String("0.0001"), Unit: aws.String("USD")},
					},
				},
			},
			want: map[string]float64{
				"Micro Service": 0.0001,
			},
		},
		{
			name: "handles_large_amounts",
			groups: []types.Group{
				{
					Keys: []string{"Expensive Service"},
					Metrics: map[string]types.MetricValue{
						costsAggregation: {Amount: aws.String("999999.99"), Unit: aws.String("USD")},
					},
				},
			},
			want: map[string]float64{
				"Expensive Service": 999999.99,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.filterGroups(tt.groups, costsAggregation)

			// Check that we got the expected number of results
			if len(got) != len(tt.want) {
				t.Errorf("filterGroups() returned %d groups, want %d", len(got), len(tt.want))
			}

			// Check each expected service
			for serviceName, wantAmount := range tt.want {
				gotCost, ok := got[serviceName]
				if !ok {
					t.Errorf("filterGroups() missing service %q", serviceName)
					continue
				}

				// Compare floats with small tolerance
				if diff := gotCost.Amount - wantAmount; diff > 0.0001 || diff < -0.0001 {
					t.Errorf("filterGroups() service %q amount = %v, want %v",
						serviceName, gotCost.Amount, wantAmount)
				}

				// Verify unit is preserved
				if gotCost.Unit != "USD" {
					t.Errorf("filterGroups() service %q unit = %q, want USD",
						serviceName, gotCost.Unit)
				}
			}

			// Check no unexpected services
			for serviceName := range got {
				if _, ok := tt.want[serviceName]; !ok {
					t.Errorf("filterGroups() returned unexpected service %q", serviceName)
				}
			}
		})
	}
}

func BenchmarkFilterGroups(b *testing.B) {
	s := &service{}

	// Create a realistic set of groups
	groups := make([]types.Group, 50)
	for i := range 50 {
		amount := "0.00"
		if i%3 != 0 { // 2/3 have non-zero cost
			amount = "100.50"
		}
		groups[i] = types.Group{
			Keys: []string{"Service " + string(rune('A'+i%26))},
			Metrics: map[string]types.MetricValue{
				costsAggregation: {
					Amount: aws.String(amount),
					Unit:   aws.String("USD"),
				},
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.filterGroups(groups, costsAggregation)
	}
}

func TestDateHelpers_Consistency(t *testing.T) {
	s := &service{}

	// Test that for any date, first day is always <= input <= last day
	testDates := []time.Time{
		time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
		time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), // leap year
		time.Date(2023, 2, 28, 0, 0, 0, 0, time.UTC), // non-leap year
		time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	for _, date := range testDates {
		first := s.getFirstDayOfMonth(date)
		last := s.getLastDayOfMonth(date)

		// First should be before or equal to the input
		if first.After(date) {
			t.Errorf("getFirstDayOfMonth(%v) = %v is after input", date, first)
		}

		// Last should be after or equal to the input
		if last.Before(date) && last.Day() < date.Day() {
			t.Errorf("getLastDayOfMonth(%v) = %v day is before input day", date, last)
		}

		// First and last should be in the same month
		if first.Month() != last.Month() {
			t.Errorf("First (%v) and last (%v) are in different months for input %v",
				first, last, date)
		}

		// First should be before last
		if !first.Before(last) && first.Day() != last.Day() {
			t.Errorf("First (%v) is not before last (%v) for input %v",
				first, last, date)
		}
	}
}
