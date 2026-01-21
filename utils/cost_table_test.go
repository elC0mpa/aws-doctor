package utils

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
			got := orderCostServices(&tt.costGroups)

			// Check length
			if len(got) != len(tt.costGroups) {
				t.Errorf("orderCostServices() returned %d items, want %d", len(got), len(tt.costGroups))
				return
			}

			// Skip order check for equal values test
			if tt.wantOrder == nil {
				return
			}

			// Check order
			for i, wantName := range tt.wantOrder {
				if got[i].Name != wantName {
					t.Errorf("orderCostServices()[%d].Name = %q, want %q", i, got[i].Name, wantName)
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

	result := orderCostServices(&costGroups)

	// Verify descending order
	for i := 1; i < len(result); i++ {
		if result[i].Amount > result[i-1].Amount {
			t.Errorf("Not sorted descending: index %d (%.2f) > index %d (%.2f)",
				i, result[i].Amount, i-1, result[i-1].Amount)
		}
	}
}

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

func BenchmarkOrderCostServices(b *testing.B) {
	costGroups := model.CostGroup{
		"Amazon EC2":         {Amount: 500.0, Unit: "USD"},
		"Amazon S3":          {Amount: 200.0, Unit: "USD"},
		"Amazon RDS":         {Amount: 300.0, Unit: "USD"},
		"AWS Lambda":         {Amount: 50.0, Unit: "USD"},
		"Amazon CloudFront":  {Amount: 100.0, Unit: "USD"},
		"Amazon DynamoDB":    {Amount: 75.0, Unit: "USD"},
		"Amazon ElastiCache": {Amount: 150.0, Unit: "USD"},
		"AWS Fargate":        {Amount: 125.0, Unit: "USD"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orderCostServices(&costGroups)
	}
}
