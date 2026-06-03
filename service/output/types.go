package output

import (
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
)

// Format represents the output format type
type Format string

// FormatTable represents the table output format.
const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
)

// Renderer defines the interface for formatting and outputting data.
type Renderer interface {
	// RenderCostComparison outputs cost comparison data in the configured format
	RenderCostComparison(input model.RenderCostComparisonInput) error

	// RenderTrend outputs trend data in the configured format
	RenderTrend(accountID string, costInfo []model.CostInfo, services []string) error

	// RenderWaste outputs waste report data in the configured format
	RenderWaste(input model.RenderWasteInput, pricingSvc pricing.Service) error

	// RenderWasteInteractive launches the Bubble Tea TUI for waste output
	RenderWasteInteractive(accountID string, resultCh <-chan model.ScopeResult, scopes []string, pricingSvc pricing.Service) error

	// IsInteractive returns true if the current format should use the TUI
	IsInteractive() bool
}
