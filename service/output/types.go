package output

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/barchart"
	costtable "github.com/elC0mpa/aws-doctor/utils/cost_table"
	csvoutput "github.com/elC0mpa/aws-doctor/utils/csv_output"
	jsonoutput "github.com/elC0mpa/aws-doctor/utils/json_output"
	"github.com/elC0mpa/aws-doctor/utils/spinner"
	wastetable "github.com/elC0mpa/aws-doctor/utils/waste_table"
)

// Format represents the output format type
type Format string

// FormatTable represents the table output format.
const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
)

// Renderer defines the interface for drawing tables and charts
type Renderer interface {
	DrawCostTable(input model.RenderCostComparisonInput)
	OutputCostComparisonJSON(input model.RenderCostComparisonInput) error
	OutputCostComparisonCSV(input model.RenderCostComparisonInput) error
	DrawTrendChart(accountID string, costInfo []model.CostInfo)
	OutputTrendJSON(accountID string, costInfo []model.CostInfo, services []string) error
	OutputTrendCSV(monthlyCosts []model.CostInfo, services []string) error
	DrawWasteTable(input model.RenderWasteInput)
	OutputWasteJSON(input model.RenderWasteInput) error
	OutputWasteCSV(input model.RenderWasteInput) error
	StopSpinner()
}

type realRenderer struct{}

func (r *realRenderer) DrawCostTable(input model.RenderCostComparisonInput) {
	costtable.DrawCostTable(input)
}

func (r *realRenderer) OutputCostComparisonJSON(input model.RenderCostComparisonInput) error {
	return jsonoutput.OutputCostComparisonJSON(input)
}

func (r *realRenderer) OutputCostComparisonCSV(input model.RenderCostComparisonInput) error {
	return csvoutput.OutputCostComparisonCSV(input)
}

func (r *realRenderer) DrawTrendChart(accountID string, costInfo []model.CostInfo) {
	barchart.DrawTrendChart(accountID, costInfo)
}

func (r *realRenderer) OutputTrendJSON(accountID string, costInfo []model.CostInfo, services []string) error {
	return jsonoutput.OutputTrendJSON(accountID, costInfo, services)
}

func (r *realRenderer) OutputTrendCSV(monthlyCosts []model.CostInfo, services []string) error {
	return csvoutput.OutputTrendCSV(monthlyCosts, services)
}

func (r *realRenderer) DrawWasteTable(input model.RenderWasteInput) {
	wastetable.DrawWasteTable(input)
}

func (r *realRenderer) OutputWasteJSON(input model.RenderWasteInput) error {
	return jsonoutput.OutputWasteJSON(input)
}

func (r *realRenderer) OutputWasteCSV(input model.RenderWasteInput) error {
	return csvoutput.OutputWasteCSV(input)
}

func (r *realRenderer) StopSpinner() {
	spinner.StopSpinner()
}

// service is the internal implementation
type service struct {
	format   Format
	renderer Renderer
}

// Service defines the interface for output operations
type Service interface {
	// RenderCostComparison outputs cost comparison data in the configured format
	RenderCostComparison(input model.RenderCostComparisonInput) error

	// RenderTrend outputs trend data in the configured format
	RenderTrend(accountID string, costInfo []model.CostInfo, services []string) error

	// RenderWaste outputs waste report data in the configured format
	RenderWaste(input model.RenderWasteInput) error

	// StopSpinner stops the loading spinner before rendering output
	StopSpinner()
}
