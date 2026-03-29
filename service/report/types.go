package report

import (
	"github.com/elC0mpa/aws-doctor/model"
)

// Type represents the type of report being generated.
type Type string

const (
	// CostReport represents a cost comparison report.
	CostReport Type = "cost"
	// TrendReport represents a cost trend report.
	TrendReport Type = "trend"
	// WasteReport represents an AWS waste report.
	WasteReport Type = "waste"
)

// Service defines the interface for generating PDF reports.
type Service interface {
	// GenerateCostComparisonReport creates a PDF report for cost comparison.
	GenerateCostComparisonReport(input model.RenderCostComparisonInput, reportPath string) (*string, error)

	// GenerateTrendReport creates a PDF report for cost trends.
	GenerateTrendReport(accountID string, costInfo []model.CostInfo, services []string, reportPath string) (*string, error)

	// GenerateWasteReport creates a PDF report for AWS waste detection.
	GenerateWasteReport(input model.RenderWasteInput, reportPath string) (*string, error)
}
