package outputshared

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elC0mpa/aws-doctor/model"
)

// TrendRow represents a row in the trend output
type TrendRow struct {
	PeriodStart string
	PeriodEnd   string
	TotalCost   string
	Unit        string
}

// ToSlice converts the TrendRow to a string slice
func (r TrendRow) ToSlice() []string {
	return []string{
		r.PeriodStart,
		r.PeriodEnd,
		r.TotalCost,
		r.Unit,
	}
}

// PresentTrend returns a slice of TrendRow for the given monthly costs
func PresentTrend(monthlyCosts []model.CostInfo) []TrendRow {
	result := make([]TrendRow, 0, len(monthlyCosts))

	for _, cost := range monthlyCosts {
		total := cost.CostGroup["Total"]
		result = append(result, TrendRow{
			PeriodStart: aws.ToString(cost.Start),
			PeriodEnd:   aws.ToString(cost.End),
			TotalCost:   FormatCost(total.Amount, ""),
			Unit:        total.Unit,
		})
	}

	return result
}
