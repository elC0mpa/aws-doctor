package report

import (
	"github.com/elC0mpa/aws-doctor/model"
)

// ReportType represents the type of report being generated.
type ReportType string

const (
	// CostReport represents a cost comparison report.
	CostReport ReportType = "cost"
	// TrendReport represents a cost trend report.
	TrendReport ReportType = "trend"
	// WasteReport represents an AWS waste report.
	WasteReport ReportType = "waste"
)

// Service defines the interface for generating PDF reports.
type Service interface {
	// GenerateCostComparisonReport creates a PDF report for cost comparison.
	GenerateCostComparisonReport(input model.RenderCostComparisonInput, reportPath string) error

	// GenerateTrendReport creates a PDF report for cost trends.
	GenerateTrendReport(accountID string, costInfo []model.CostInfo, services []string, reportPath string) error

	// GenerateWasteReport creates a PDF report for AWS waste detection.
	GenerateWasteReport(input model.RenderWasteInput, reportPath string) error
}
