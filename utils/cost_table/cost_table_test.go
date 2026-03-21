package costtable

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elC0mpa/aws-doctor/model"
)

func TestPopulateFirstRow(t *testing.T) {
	tests := []struct {
		name             string
		lastTotalCost    string
		currentTotalCost string
		wantIncrease     bool // true if current > last (red), false if decrease (green)
	}{
		{
			name:             "costs_increased",
			lastTotalCost:    "100.00 USD",
			currentTotalCost: "150.00 USD",
			wantIncrease:     true,
		},
		{
			name:             "costs_decreased",
			lastTotalCost:    "150.00 USD",
			currentTotalCost: "100.00 USD",
			wantIncrease:     false,
		},
		{
			name:             "costs_unchanged",
			lastTotalCost:    "100.00 USD",
			currentTotalCost: "100.00 USD",
			wantIncrease:     false, // Equal is treated as not increased
		},
		{
			name:             "large_increase",
			lastTotalCost:    "1000.00 USD",
			currentTotalCost: "5000.00 USD",
			wantIncrease:     true,
		},
		{
			name:             "small_decimal_difference",
			lastTotalCost:    "100.00 USD",
			currentTotalCost: "100.01 USD",
			wantIncrease:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := populateFirstRow(tt.lastTotalCost, tt.currentTotalCost)

			// Check row has 4 columns
			if len(row) != 4 {
				t.Errorf("populateFirstRow() returned %d columns, want 4", len(row))
				return
			}

			// Row should have values in all columns
			for i, val := range row {
				if val == nil || val == "" {
					t.Errorf("populateFirstRow() column %d is empty", i)
				}
			}
		})
	}
}

func TestPopulateRow(t *testing.T) {
	tests := []struct {
		name              string
		lastMonthGroups   model.CostInfo
		currentMonthGroup model.ServiceCost
		wantIncrease      bool
	}{
		{
			name: "service_cost_increased",
			lastMonthGroups: model.CostInfo{
				CostGroup: model.CostGroup{
					"Amazon EC2": {Amount: 100.0, Unit: "USD"},
				},
			},
			currentMonthGroup: model.ServiceCost{
				Name:   "Amazon EC2",
				Amount: 150.0,
				Unit:   "USD",
			},
			wantIncrease: true,
		},
		{
			name: "service_cost_decreased",
			lastMonthGroups: model.CostInfo{
				CostGroup: model.CostGroup{
					"Amazon EC2": {Amount: 150.0, Unit: "USD"},
				},
			},
			currentMonthGroup: model.ServiceCost{
				Name:   "Amazon EC2",
				Amount: 100.0,
				Unit:   "USD",
			},
			wantIncrease: false,
		},
		{
			name: "new_service_not_in_last_month",
			lastMonthGroups: model.CostInfo{
				CostGroup: model.CostGroup{},
			},
			currentMonthGroup: model.ServiceCost{
				Name:   "New Service",
				Amount: 50.0,
				Unit:   "USD",
			},
			wantIncrease: true, // 50 > 0
		},
		{
			name: "service_cost_unchanged",
			lastMonthGroups: model.CostInfo{
				CostGroup: model.CostGroup{
					"Amazon S3": {Amount: 75.0, Unit: "USD"},
				},
			},
			currentMonthGroup: model.ServiceCost{
				Name:   "Amazon S3",
				Amount: 75.0,
				Unit:   "USD",
			},
			wantIncrease: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := populateRow(tt.lastMonthGroups, tt.currentMonthGroup)

			// Check row has 4 columns
			if len(row) != 4 {
				t.Errorf("populateRow() returned %d columns, want 4", len(row))
				return
			}

			// Row should have values in all columns
			for i, val := range row {
				if val == nil || val == "" {
					t.Errorf("populateRow() column %d is empty", i)
				}
			}
		})
	}
}

// captureTableOutput captures stdout during function execution
func captureTableOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer

	_, _ = io.Copy(&buf, r)

	return buf.String()
}

func TestDrawCostTable(t *testing.T) {
	lastMonthGroups := &model.CostInfo{
		CostGroup: model.CostGroup{
			"Amazon EC2": {Amount: 100.0, Unit: "USD"},
			"Amazon S3":  {Amount: 50.0, Unit: "USD"},
		},
	}
	lastMonthGroups.Start = aws.String("2024-01-01")
	lastMonthGroups.End = aws.String("2024-01-31")

	currentMonthGroups := &model.CostInfo{
		CostGroup: model.CostGroup{
			"Amazon EC2": {Amount: 120.0, Unit: "USD"},
			"Amazon S3":  {Amount: 45.0, Unit: "USD"},
		},
	}
	currentMonthGroups.Start = aws.String("2024-02-01")
	currentMonthGroups.End = aws.String("2024-02-29")

	output := captureTableOutput(func() {
		DrawCostTable(model.RenderCostComparisonInput{
			AccountID:        "123456789012",
			LastTotalCost:    "150.00 USD",
			CurrentTotalCost: "165.00 USD",
			LastMonth:        lastMonthGroups,
			CurrentMonth:     currentMonthGroups,
		})
	})

	// Verify output contains expected elements
	if !strings.Contains(output, "AWS COST DIAGNOSIS") {
		t.Error("DrawCostTable() output missing header")
	}

	if !strings.Contains(output, "123456789012") {
		t.Error("DrawCostTable() output missing account ID")
	}

	// Verify table structure is present (tables use box-drawing characters)
	if len(output) < 200 {
		t.Errorf("DrawCostTable() output seems too short: %d chars", len(output))
	}

	// Verify service names appear in output
	if !strings.Contains(output, "EC2") {
		t.Error("DrawCostTable() output missing EC2 service")
	}
}

func TestDrawCostTable_CostsIncreased(t *testing.T) {
	lastMonthGroups := &model.CostInfo{
		CostGroup: model.CostGroup{
			"Amazon EC2": {Amount: 100.0, Unit: "USD"},
		},
	}
	lastMonthGroups.Start = aws.String("2024-01-01")
	lastMonthGroups.End = aws.String("2024-01-31")

	currentMonthGroups := &model.CostInfo{
		CostGroup: model.CostGroup{
			"Amazon EC2": {Amount: 200.0, Unit: "USD"},
		},
	}
	currentMonthGroups.Start = aws.String("2024-02-01")
	currentMonthGroups.End = aws.String("2024-02-29")

	output := captureTableOutput(func() {
		DrawCostTable(model.RenderCostComparisonInput{
			AccountID:        "123456789012",
			LastTotalCost:    "100.00 USD",
			CurrentTotalCost: "200.00 USD",
			LastMonth:        lastMonthGroups,
			CurrentMonth:     currentMonthGroups,
		})
	})

	// Should have output (table with red colors for increases)
	if len(output) == 0 {
		t.Error("DrawCostTable() with increased costs produced no output")
	}
}

func TestDrawCostTable_CostsDecreased(t *testing.T) {
	lastMonthGroups := &model.CostInfo{
		CostGroup: model.CostGroup{
			"Amazon EC2": {Amount: 200.0, Unit: "USD"},
		},
	}
	lastMonthGroups.Start = aws.String("2024-01-01")
	lastMonthGroups.End = aws.String("2024-01-31")

	currentMonthGroups := &model.CostInfo{
		CostGroup: model.CostGroup{
			"Amazon EC2": {Amount: 100.0, Unit: "USD"},
		},
	}
	currentMonthGroups.Start = aws.String("2024-02-01")
	currentMonthGroups.End = aws.String("2024-02-29")

	output := captureTableOutput(func() {
		DrawCostTable(model.RenderCostComparisonInput{
			AccountID:        "123456789012",
			LastTotalCost:    "200.00 USD",
			CurrentTotalCost: "100.00 USD",
			LastMonth:        lastMonthGroups,
			CurrentMonth:     currentMonthGroups,
		})
	})

	// Should have output (table with green colors for decreases)
	if len(output) == 0 {
		t.Error("DrawCostTable() with decreased costs produced no output")
	}
}

func TestDrawCostTable_EmptyServices(t *testing.T) {
	lastMonthGroups := &model.CostInfo{
		CostGroup: model.CostGroup{},
	}
	lastMonthGroups.Start = aws.String("2024-01-01")
	lastMonthGroups.End = aws.String("2024-01-31")

	currentMonthGroups := &model.CostInfo{
		CostGroup: model.CostGroup{},
	}
	currentMonthGroups.Start = aws.String("2024-02-01")
	currentMonthGroups.End = aws.String("2024-02-29")

	output := captureTableOutput(func() {
		DrawCostTable(model.RenderCostComparisonInput{
			AccountID:        "123456789012",
			LastTotalCost:    "0.00 USD",
			CurrentTotalCost: "0.00 USD",
			LastMonth:        lastMonthGroups,
			CurrentMonth:     currentMonthGroups,
		})
	})

	// Should still produce header and table structure
	if !strings.Contains(output, "AWS COST DIAGNOSIS") {
		t.Error("DrawCostTable() with empty services missing header")
	}
}

func BenchmarkDrawCostTable(b *testing.B) {
	lastMonthGroups := &model.CostInfo{
		CostGroup: model.CostGroup{
			"Amazon EC2": {Amount: 100.0, Unit: "USD"},
			"Amazon S3":  {Amount: 50.0, Unit: "USD"},
			"AWS Lambda": {Amount: 25.0, Unit: "USD"},
		},
	}
	lastMonthGroups.Start = aws.String("2024-01-01")
	lastMonthGroups.End = aws.String("2024-01-31")

	currentMonthGroups := &model.CostInfo{
		CostGroup: model.CostGroup{
			"Amazon EC2": {Amount: 120.0, Unit: "USD"},
			"Amazon S3":  {Amount: 45.0, Unit: "USD"},
			"AWS Lambda": {Amount: 30.0, Unit: "USD"},
		},
	}
	currentMonthGroups.Start = aws.String("2024-02-01")
	currentMonthGroups.End = aws.String("2024-02-29")

	// Redirect stdout to discard
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)

	defer func() { os.Stdout = old }()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		DrawCostTable(model.RenderCostComparisonInput{
			AccountID:        "123456789012",
			LastTotalCost:    "175.00 USD",
			CurrentTotalCost: "195.00 USD",
			LastMonth:        lastMonthGroups,
			CurrentMonth:     currentMonthGroups,
		})
	}
}
