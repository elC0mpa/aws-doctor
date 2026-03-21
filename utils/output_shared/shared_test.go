package outputshared

import (
	"testing"

	"github.com/elC0mpa/aws-doctor/model"
)

func TestOrderCostServices(t *testing.T) {
	tests := []struct {
		name       string
		costGroups model.CostGroup
		wantOrder  []string // Expected service names in order (highest to lowest)
	}{
		{
			name:       "empty_groups",
			costGroups: model.CostGroup{},
			wantOrder:  []string{},
		},
		{
			name: "single_service",
			costGroups: model.CostGroup{
				"Amazon EC2": {Amount: 100.0, Unit: "USD"},
			},
			wantOrder: []string{"Amazon EC2"},
		},
		{
			name: "two_services_already_sorted",
			costGroups: model.CostGroup{
				"Amazon EC2": {Amount: 200.0, Unit: "USD"},
				"Amazon S3":  {Amount: 100.0, Unit: "USD"},
			},
			wantOrder: []string{"Amazon EC2", "Amazon S3"},
		},
		{
			name: "two_services_reverse_sorted",
			costGroups: model.CostGroup{
				"Amazon S3":  {Amount: 100.0, Unit: "USD"},
				"Amazon EC2": {Amount: 200.0, Unit: "USD"},
			},
			wantOrder: []string{"Amazon EC2", "Amazon S3"},
		},
		{
			name: "multiple_services",
			costGroups: model.CostGroup{
				"AWS Lambda":  {Amount: 50.0, Unit: "USD"},
				"Amazon EC2":  {Amount: 300.0, Unit: "USD"},
				"Amazon S3":   {Amount: 100.0, Unit: "USD"},
				"Amazon RDS":  {Amount: 200.0, Unit: "USD"},
				"AWS Fargate": {Amount: 75.0, Unit: "USD"},
			},
			wantOrder: []string{"Amazon EC2", "Amazon RDS", "Amazon S3", "AWS Fargate", "AWS Lambda"},
		},
		{
			name: "services_with_zero_cost",
			costGroups: model.CostGroup{
				"Amazon EC2": {Amount: 100.0, Unit: "USD"},
				"Free Tier":  {Amount: 0.0, Unit: "USD"},
				"Amazon S3":  {Amount: 50.0, Unit: "USD"},
			},
			wantOrder: []string{"Amazon EC2", "Amazon S3", "Free Tier"},
		},
		{
			name: "services_with_equal_cost",
			costGroups: model.CostGroup{
				"Service A": {Amount: 100.0, Unit: "USD"},
				"Service B": {Amount: 100.0, Unit: "USD"},
				"Service C": {Amount: 100.0, Unit: "USD"},
			},
			wantOrder: nil, // Order among equal values is not deterministic
		},
		{
			name: "services_with_decimal_amounts",
			costGroups: model.CostGroup{
				"Service A": {Amount: 100.50, Unit: "USD"},
				"Service B": {Amount: 100.49, Unit: "USD"},
				"Service C": {Amount: 100.51, Unit: "USD"},
			},
			wantOrder: []string{"Service C", "Service A", "Service B"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OrderCostServices(&tt.costGroups)

			// Check length
			if len(got) != len(tt.costGroups) {
				t.Errorf("OrderCostServices() returned %d items, want %d", len(got), len(tt.costGroups))
				return
			}

			// Skip order check for equal values test
			if tt.wantOrder == nil {
				return
			}

			// Check order
			for i, wantName := range tt.wantOrder {
				if got[i].Name != wantName {
					t.Errorf("OrderCostServices()[%d].Name = %q, want %q", i, got[i].Name, wantName)
				}
			}

			// Verify amounts and units are preserved
			for _, service := range got {
				original := tt.costGroups[service.Name]
				if service.Amount != original.Amount {
					t.Errorf("Amount mismatch for %s: got %v, want %v", service.Name, service.Amount, original.Amount)
				}

				if service.Unit != original.Unit {
					t.Errorf("Unit mismatch for %s: got %v, want %v", service.Name, service.Unit, original.Unit)
				}
			}
		})
	}
}

func TestOrderCostServices_IsSortedDescending(t *testing.T) {
	costGroups := model.CostGroup{
		"A": {Amount: 10.0, Unit: "USD"},
		"B": {Amount: 50.0, Unit: "USD"},
		"C": {Amount: 30.0, Unit: "USD"},
		"D": {Amount: 20.0, Unit: "USD"},
		"E": {Amount: 40.0, Unit: "USD"},
	}

	result := OrderCostServices(&costGroups)

	// Verify descending order
	for i := 1; i < len(result); i++ {
		if result[i].Amount > result[i-1].Amount {
			t.Errorf("Not sorted descending: index %d (%.2f) > index %d (%.2f)",
				i, result[i].Amount, i-1, result[i-1].Amount)
		}
	}
}

func TestParseCostString(t *testing.T) {
	tests := []struct {
		input      string
		wantAmount float64
		wantUnit   string
	}{
		{"10.50 USD", 10.50, "USD"},
		{"0.00 EUR", 0.00, "EUR"},
		{"123.456", 123.456, ""},
	}

	for _, tt := range tests {
		gotAmount, gotUnit := ParseCostString(tt.input)
		if gotAmount != tt.wantAmount || gotUnit != tt.wantUnit {
			t.Errorf("ParseCostString(%q) = (%v, %q), want (%v, %q)", tt.input, gotAmount, gotUnit, tt.wantAmount, tt.wantUnit)
		}
	}
}

func TestFormatCost(t *testing.T) {
	tests := []struct {
		amount float64
		unit   string
		want   string
	}{
		{10.50, "USD", "10.50 USD"},
		{0.00, "EUR", "0.00 EUR"},
	}

	for _, tt := range tests {
		got := FormatCost(tt.amount, tt.unit)
		if got != tt.want {
			t.Errorf("FormatCost(%v, %q) = %q, want %q", tt.amount, tt.unit, got, tt.want)
		}
	}
}
