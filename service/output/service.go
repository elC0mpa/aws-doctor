// Package output provides a service for rendering results to the console.
package output

import (
	"github.com/elC0mpa/aws-doctor/model"
)

// NewService creates a new output service with the specified format
func NewService(format string) Service {
	f := FormatTable
	switch format {
	case "json":
		f = FormatJSON
	case "csv":
		f = FormatCSV
	}

	return &service{
		format:   f,
		renderer: &realRenderer{},
	}
}

func (s *service) RenderCostComparison(input model.RenderCostComparisonInput) error {
	if s.format == FormatJSON {
		return s.renderer.OutputCostComparisonJSON(input)
	}

	s.renderer.DrawCostTable(input)

	return nil
}

func (s *service) RenderTrend(accountID string, costInfo []model.CostInfo) error {
	if s.format == FormatJSON {
		return s.renderer.OutputTrendJSON(accountID, costInfo)
	}

	s.renderer.DrawTrendChart(accountID, costInfo)

	return nil
}

func (s *service) RenderWaste(input model.RenderWasteInput) error {
	switch s.format {
	case FormatJSON:
		return s.renderer.OutputWasteJSON(input)
	case FormatCSV:
		return s.renderer.OutputWasteCSV(input)
	default:
		s.renderer.DrawWasteTable(input)
		return nil
	}
}

func (s *service) StopSpinner() {
	s.renderer.StopSpinner()
}
