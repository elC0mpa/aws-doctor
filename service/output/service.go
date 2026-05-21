// Package output provides a service for rendering results to the console.
package output

import (
	"os"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	"golang.org/x/term"
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
	switch s.format {
	case FormatJSON:
		return s.renderer.OutputCostComparisonJSON(input)
	case FormatCSV:
		return s.renderer.OutputCostComparisonCSV(input)
	default:
		s.renderer.DrawCostTable(input)
		return nil
	}
}

func (s *service) RenderTrend(accountID string, costInfo []model.CostInfo, services []string) error {
	switch s.format {
	case FormatJSON:
		return s.renderer.OutputTrendJSON(accountID, costInfo, services)
	case FormatCSV:
		return s.renderer.OutputTrendCSV(costInfo, services)
	default:
		s.renderer.DrawTrendChart(accountID, costInfo)
		return nil
	}
}

func (s *service) RenderWaste(input model.RenderWasteInput, pricingSvc pricing.Service) error {
	switch s.format {
	case FormatJSON:
		return s.renderer.OutputWasteJSON(input, pricingSvc)
	case FormatCSV:
		return s.renderer.OutputWasteCSV(input, pricingSvc)
	default:
		s.renderer.DrawWasteTable(input, pricingSvc)
		return nil
	}
}

var isTerminalFn = term.IsTerminal

func (s *service) IsInteractive() bool {
	return s.format == FormatTable && isTerminalFn(int(os.Stdout.Fd()))
}

func (s *service) RenderWasteInteractive(accountID string, resultCh <-chan model.ScopeResult, scopes []string, pricingSvc pricing.Service) error {
	return s.renderer.RenderWasteInteractive(accountID, resultCh, scopes, pricingSvc)
}

func (s *service) StopSpinner() {
	s.renderer.StopSpinner()
}

func (s *service) SetSpinnerMessage(message string) {
	s.renderer.SetSpinnerMessage(message)
}

func (s *service) PrintReportSuccess(path string) {
	s.renderer.PrintReportSuccess(path)
}

func (s *service) PrintAlreadyLatest(version string) {
	s.renderer.PrintAlreadyLatest(version)
}

func (s *service) PrintHomebrewUpdate() {
	s.renderer.PrintHomebrewUpdate()
}

func (s *service) PrintGoInstallUpdate() {
	s.renderer.PrintGoInstallUpdate()
}

func (s *service) PrintRateLimitError() {
	s.renderer.PrintRateLimitError()
}

func (s *service) PrintUpdateError(err error) {
	s.renderer.PrintUpdateError(err)
}

func (s *service) PrintWasteError(err error) {
	s.renderer.PrintWasteError(err)
}

func (s *service) RenderVersion(versionInfo model.VersionInfo) {
	s.renderer.RenderVersion(versionInfo)
}

func (s *service) PrintFirstDayOfMonthError() {
	s.renderer.PrintFirstDayOfMonthError()
}

func (s *service) PrintNewVersionAvailable(currentVersion, latestVersion string) {
	s.renderer.PrintNewVersionAvailable(currentVersion, latestVersion)
}
