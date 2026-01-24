package utils

import (
	"testing"
)

func TestParseCostString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{
			name:  "standard_format_with_usd",
			input: "123.45 USD",
			want:  123.45,
		},
		{
			name:  "integer_amount",
			input: "100 USD",
			want:  100.0,
		},
		{
			name:  "zero_amount",
			input: "0 USD",
			want:  0.0,
		},
		{
			name:  "zero_with_decimals",
			input: "0.00 USD",
			want:  0.0,
		},
		{
			name:  "large_amount",
			input: "999999.99 USD",
			want:  999999.99,
		},
		{
			name:  "small_amount",
			input: "0.0001 USD",
			want:  0.0001,
		},
		{
			name:  "negative_amount",
			input: "-50.25 USD",
			want:  -50.25,
		},
		{
			name:  "amount_only_no_currency",
			input: "42.50",
			want:  42.50,
		},
		{
			name:  "empty_string",
			input: "",
			want:  0.0,
		},
		{
			name:  "invalid_non_numeric",
			input: "not-a-number USD",
			want:  0.0,
		},
		{
			name:  "whitespace_only",
			input: "   ",
			want:  0.0,
		},
		{
			name:  "multiple_spaces",
			input: "123.45  USD  extra",
			want:  123.45,
		},
		{
			name:  "scientific_notation",
			input: "1.5e2 USD",
			want:  150.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseCostString(tt.input)
			if got != tt.want {
				t.Errorf("ParseCostString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func BenchmarkParseCostString(b *testing.B) {
	inputs := []string{
		"123.45 USD",
		"0.00 USD",
		"999999.99 USD",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			ParseCostString(input)
		}
	}
}
