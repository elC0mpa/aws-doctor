package output

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cetypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
)

func TestRealRenderer_DrawCostTable(t *testing.T) {
	r := &realRenderer{}

	lastMonth := &model.CostInfo{
		DateInterval: cetypes.DateInterval{
			Start: aws.String("2023-01-01"),
			End:   aws.String("2023-01-31"),
		},
		CostGroup: model.CostGroup{
			"EC2": {Amount: 100, Unit: "USD"},
		},
	}
	currentMonth := &model.CostInfo{
		DateInterval: cetypes.DateInterval{
			Start: aws.String("2023-02-01"),
			End:   aws.String("2023-02-28"),
		},
		CostGroup: model.CostGroup{
			"EC2": {Amount: 120, Unit: "USD"},
		},
	}

	// This calls external utils which print to stdout.
	// We just want to ensure it doesn't panic and covers the code.
	assert.NotPanics(t, func() {
		r.DrawCostTable(model.RenderCostComparisonInput{
			AccountID:        "123456789012",
			LastTotalCost:    "100.00 USD",
			CurrentTotalCost: "120.00 USD",
			LastMonth:        lastMonth,
			CurrentMonth:     currentMonth,
		})
	})
}

func TestRealRenderer_OutputCostComparisonJSON(t *testing.T) {
	r := &realRenderer{}

	lastMonth := &model.CostInfo{
		DateInterval: cetypes.DateInterval{
			Start: aws.String("2023-01-01"),
			End:   aws.String("2023-01-31"),
		},
		CostGroup: model.CostGroup{
			"EC2": {Amount: 100, Unit: "USD"},
		},
	}
	currentMonth := &model.CostInfo{
		DateInterval: cetypes.DateInterval{
			Start: aws.String("2023-02-01"),
			End:   aws.String("2023-02-28"),
		},
		CostGroup: model.CostGroup{
			"EC2": {Amount: 120, Unit: "USD"},
		},
	}

	err := r.OutputCostComparisonJSON(model.RenderCostComparisonInput{
		AccountID:        "123456789012",
		LastTotalCost:    "100.00 USD",
		CurrentTotalCost: "120.00 USD",
		LastMonth:        lastMonth,
		CurrentMonth:     currentMonth,
	})
	assert.NoError(t, err)
}

func TestRealRenderer_DrawTrendChart(t *testing.T) {
	r := &realRenderer{}

	costInfo := []model.CostInfo{
		{
			DateInterval: cetypes.DateInterval{
				Start: aws.String("2023-01-01"),
				End:   aws.String("2023-01-31"),
			},
			CostGroup: model.CostGroup{
				"Total": {Amount: 100, Unit: "USD"},
			},
		},
	}

	assert.NotPanics(t, func() {
		r.DrawTrendChart("123456789012", costInfo)
	})
}

func TestRealRenderer_OutputTrendJSON(t *testing.T) {
	r := &realRenderer{}

	costInfo := []model.CostInfo{
		{
			DateInterval: cetypes.DateInterval{
				Start: aws.String("2023-01-01"),
				End:   aws.String("2023-01-31"),
			},
			CostGroup: model.CostGroup{
				"Total": {Amount: 100, Unit: "USD"},
			},
		},
	}

	err := r.OutputTrendJSON("123456789012", costInfo)
	assert.NoError(t, err)
}

func TestRealRenderer_DrawWasteTable(t *testing.T) {
	r := &realRenderer{}

	assert.NotPanics(t, func() {
		r.DrawWasteTable(model.RenderWasteInput{AccountID: "123456789012"})
	})
}

func TestRealRenderer_OutputWasteJSON(t *testing.T) {
	r := &realRenderer{}

	err := r.OutputWasteJSON(model.RenderWasteInput{AccountID: "123456789012"})
	assert.NoError(t, err)
}

func TestRealRenderer_StopSpinner(t *testing.T) {
	r := &realRenderer{}

	assert.NotPanics(t, func() {
		r.StopSpinner()
	})
}
