package outputshared

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elC0mpa/aws-doctor/model"
)

func TestPresentTrend(t *testing.T) {
	costs := []model.CostInfo{
		{
			CostGroup: model.CostGroup{
				"Total": {Amount: 100.50, Unit: "USD"},
			},
		},
	}
	costs[0].Start = aws.String("2024-01-01")
	costs[0].End = aws.String("2024-01-31")

	rows := PresentTrend(costs)

	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}

	if rows[0].PeriodStart != "2024-01-01" {
		t.Errorf("expected PeriodStart '2024-01-01', got %q", rows[0].PeriodStart)
	}

	if rows[0].TotalCost != "100.50" {
		t.Errorf("expected TotalCost '100.50', got %q", rows[0].TotalCost)
	}

	if rows[0].Unit != "USD" {
		t.Errorf("expected Unit 'USD', got %q", rows[0].Unit)
	}

	slice := rows[0].ToSlice()
	if len(slice) != 4 {
		t.Errorf("expected slice length 4, got %d", len(slice))
	}
}
