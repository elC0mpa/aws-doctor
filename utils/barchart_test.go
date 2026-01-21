package utils

import (
	"strings"
	"testing"

	"github.com/elC0mpa/aws-doctor/model"
)

func TestGetBarLabel(t *testing.T) {
	tests := []struct {
		name        string
		date        string
		monthlyCost model.CostInfo
		wantPrefix  string // We check prefix since format includes the month name
		wantSuffix  string // And suffix includes amount and unit
	}{
		{
			name: "valid_january_date",
			date: "2024-01-01",
			monthlyCost: model.CostInfo{
				CostGroup: model.CostGroup{
					"Total": {Amount: 123.45, Unit: "USD"},
				},
			},
			wantPrefix: "Jan:",
			wantSuffix: "123.45 USD",
		},
		{
			name: "valid_december_date",
			date: "2024-12-15",
			monthlyCost: model.CostInfo{
				CostGroup: model.CostGroup{
					"Total": {Amount: 999.99, Unit: "USD"},
				},
			},
			wantPrefix: "Dec:",
			wantSuffix: "999.99 USD",
		},
		{
			name: "valid_july_date",
			date: "2024-07-31",
			monthlyCost: model.CostInfo{
				CostGroup: model.CostGroup{
					"Total": {Amount: 0.00, Unit: "USD"},
				},
			},
			wantPrefix: "Jul:",
			wantSuffix: "0.00 USD",
		},
		{
			name: "invalid_date_format_returns_raw",
			date: "invalid-date",
			monthlyCost: model.CostInfo{
				CostGroup: model.CostGroup{
					"Total": {Amount: 50.00, Unit: "USD"},
				},
			},
			wantPrefix: "invalid-date:",
			wantSuffix: "50.00 USD",
		},
		{
			name: "empty_date_returns_raw",
			date: "",
			monthlyCost: model.CostInfo{
				CostGroup: model.CostGroup{
					"Total": {Amount: 25.00, Unit: "USD"},
				},
			},
			wantPrefix: ":",
			wantSuffix: "25.00 USD",
		},
		{
			name: "wrong_date_format_returns_raw",
			date: "01/15/2024",
			monthlyCost: model.CostInfo{
				CostGroup: model.CostGroup{
					"Total": {Amount: 75.50, Unit: "USD"},
				},
			},
			wantPrefix: "01/15/2024:",
			wantSuffix: "75.50 USD",
		},
		{
			name: "large_amount",
			date: "2024-06-01",
			monthlyCost: model.CostInfo{
				CostGroup: model.CostGroup{
					"Total": {Amount: 123456.78, Unit: "USD"},
				},
			},
			wantPrefix: "Jun:",
			wantSuffix: "123456.78 USD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getBarLabel(tt.date, tt.monthlyCost)

			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("getBarLabel() = %q, want prefix %q", got, tt.wantPrefix)
			}

			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("getBarLabel() = %q, want suffix %q", got, tt.wantSuffix)
			}
		})
	}
}

func TestAssignRankedColors(t *testing.T) {
	tests := []struct {
		name     string
		costs    []model.CostInfo
		wantLen  int
		validate func(t *testing.T, colors []string, costs []model.CostInfo)
	}{
		{
			name:    "empty_costs",
			costs:   []model.CostInfo{},
			wantLen: 0,
			validate: func(t *testing.T, colors []string, costs []model.CostInfo) {
				// No validation needed for empty
			},
		},
		{
			name: "single_cost",
			costs: []model.CostInfo{
				{CostGroup: model.CostGroup{"Total": {Amount: 100.0, Unit: "USD"}}},
			},
			wantLen: 1,
			validate: func(t *testing.T, colors []string, costs []model.CostInfo) {
				// Single item should get rank 1 color (highest = red)
				if colors[0] != ColorRank1 {
					t.Errorf("Single item should get ColorRank1, got %s", colors[0])
				}
			},
		},
		{
			name: "two_costs_descending",
			costs: []model.CostInfo{
				{CostGroup: model.CostGroup{"Total": {Amount: 200.0, Unit: "USD"}}},
				{CostGroup: model.CostGroup{"Total": {Amount: 100.0, Unit: "USD"}}},
			},
			wantLen: 2,
			validate: func(t *testing.T, colors []string, costs []model.CostInfo) {
				// First (higher) should be rank 1, second (lower) should be rank 2
				if colors[0] != ColorRank1 {
					t.Errorf("Higher cost should get ColorRank1, got %s", colors[0])
				}
				if colors[1] != ColorRank2 {
					t.Errorf("Lower cost should get ColorRank2, got %s", colors[1])
				}
			},
		},
		{
			name: "two_costs_ascending",
			costs: []model.CostInfo{
				{CostGroup: model.CostGroup{"Total": {Amount: 100.0, Unit: "USD"}}},
				{CostGroup: model.CostGroup{"Total": {Amount: 200.0, Unit: "USD"}}},
			},
			wantLen: 2,
			validate: func(t *testing.T, colors []string, costs []model.CostInfo) {
				// Second (higher) should be rank 1, first (lower) should be rank 2
				if colors[0] != ColorRank2 {
					t.Errorf("Lower cost at index 0 should get ColorRank2, got %s", colors[0])
				}
				if colors[1] != ColorRank1 {
					t.Errorf("Higher cost at index 1 should get ColorRank1, got %s", colors[1])
				}
			},
		},
		{
			name: "six_costs_all_different",
			costs: []model.CostInfo{
				{CostGroup: model.CostGroup{"Total": {Amount: 300.0, Unit: "USD"}}}, // rank 3
				{CostGroup: model.CostGroup{"Total": {Amount: 600.0, Unit: "USD"}}}, // rank 1 (highest)
				{CostGroup: model.CostGroup{"Total": {Amount: 100.0, Unit: "USD"}}}, // rank 6
				{CostGroup: model.CostGroup{"Total": {Amount: 400.0, Unit: "USD"}}}, // rank 2
				{CostGroup: model.CostGroup{"Total": {Amount: 200.0, Unit: "USD"}}}, // rank 5
				{CostGroup: model.CostGroup{"Total": {Amount: 250.0, Unit: "USD"}}}, // rank 4
			},
			wantLen: 6,
			validate: func(t *testing.T, colors []string, costs []model.CostInfo) {
				expectedRanks := []string{ColorRank3, ColorRank1, ColorRank6, ColorRank2, ColorRank5, ColorRank4}
				for i, expected := range expectedRanks {
					if colors[i] != expected {
						t.Errorf("Index %d: got %s, want %s (amount: %.2f)",
							i, colors[i], expected, costs[i].CostGroup["Total"].Amount)
					}
				}
			},
		},
		{
			name: "equal_costs",
			costs: []model.CostInfo{
				{CostGroup: model.CostGroup{"Total": {Amount: 100.0, Unit: "USD"}}},
				{CostGroup: model.CostGroup{"Total": {Amount: 100.0, Unit: "USD"}}},
				{CostGroup: model.CostGroup{"Total": {Amount: 100.0, Unit: "USD"}}},
			},
			wantLen: 3,
			validate: func(t *testing.T, colors []string, costs []model.CostInfo) {
				// All colors should be assigned (order depends on sort stability)
				for i, c := range colors {
					if c == "" {
						t.Errorf("Index %d should have a color assigned", i)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := assignRankedColors(tt.costs)

			if len(got) != tt.wantLen {
				t.Errorf("assignRankedColors() returned %d colors, want %d", len(got), tt.wantLen)
				return
			}

			tt.validate(t, got, tt.costs)
		})
	}
}

func TestAssignRankedColors_MoreThanSixItems(t *testing.T) {
	// Test with more items than available colors
	costs := make([]model.CostInfo, 8)
	for i := 0; i < 8; i++ {
		costs[i] = model.CostInfo{
			CostGroup: model.CostGroup{
				"Total": {Amount: float64(i * 100), Unit: "USD"},
			},
		}
	}

	colors := assignRankedColors(costs)

	if len(colors) != 8 {
		t.Errorf("Expected 8 colors, got %d", len(colors))
	}

	// Count non-empty colors (should be 6, as that's the palette size)
	nonEmpty := 0
	for _, c := range colors {
		if c != "" {
			nonEmpty++
		}
	}

	if nonEmpty != 6 {
		t.Errorf("Expected 6 non-empty colors (palette size), got %d", nonEmpty)
	}
}

func BenchmarkAssignRankedColors(b *testing.B) {
	costs := make([]model.CostInfo, 6)
	for i := 0; i < 6; i++ {
		costs[i] = model.CostInfo{
			CostGroup: model.CostGroup{
				"Total": {Amount: float64((i + 1) * 100), Unit: "USD"},
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		assignRankedColors(costs)
	}
}
