package output

import (
	"fmt"
	"os"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/barchart"
	costtable "github.com/elC0mpa/aws-doctor/utils/cost_table"
	csvoutput "github.com/elC0mpa/aws-doctor/utils/csv_output"
	jsonoutput "github.com/elC0mpa/aws-doctor/utils/json_output"
	"github.com/elC0mpa/aws-doctor/utils/spinner"
	wastetable "github.com/elC0mpa/aws-doctor/utils/waste_table"
	"github.com/jedib0t/go-pretty/v6/text"
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
	SetSpinnerMessage(message string)
	PrintAlreadyLatest(version string)
	PrintHomebrewUpdate()
	PrintGoInstallUpdate()
	PrintRateLimitError()
	PrintUpdateError(err error)
	RenderVersion(versionInfo model.VersionInfo)
	PrintReportSuccess(path string)
	PrintFirstDayOfMonthError()
	PrintNewVersionAvailable(currentVersion, latestVersion string)
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

func (r *realRenderer) SetSpinnerMessage(message string) {
	spinner.SetMessage(message)
}

func (r *realRenderer) PrintAlreadyLatest(version string) {
	fmt.Println()
	fmt.Println(text.FgHiWhite.Sprintf("ℹ️ aws-doctor version %s is already the latest version", version))
}

func (r *realRenderer) PrintHomebrewUpdate() {
	fmt.Println()
	fmt.Println(text.FgHiWhite.Sprint("ℹ️ aws-doctor was installed via Homebrew. To update, run:"))
	fmt.Println()
	fmt.Println(text.FgHiWhite.Sprint("  brew upgrade aws-doctor"))
	fmt.Println()
}

func (r *realRenderer) PrintGoInstallUpdate() {
	fmt.Println()
	fmt.Println(text.FgHiWhite.Sprint("ℹ️ aws-doctor was installed via go install and reports version \"dev\", so updates cannot be tracked."))
	fmt.Println(text.FgHiWhite.Sprint("   To receive proper update notifications, remove the binary from your GOPATH/bin and reinstall with the script:"))
	fmt.Println()
	fmt.Println(text.FgHiWhite.Sprint("  curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"))
	fmt.Println()
}

func (r *realRenderer) PrintRateLimitError() {
	fmt.Println()
	fmt.Println(text.FgRed.Sprint("❌ Error: could not check GitHub release because of rate limits"))
}

func (r *realRenderer) PrintUpdateError(err error) {
	fmt.Println()
	fmt.Println(text.FgRed.Sprintf("❌ Error: failed to check for updates: %v", err))
}

func (r *realRenderer) RenderVersion(versionInfo model.VersionInfo) {
	fmt.Printf("aws-doctor version %s\n", versionInfo.Version)
	fmt.Printf("commit: %s\n", versionInfo.Commit)
	fmt.Printf("built at: %s\n", versionInfo.Date)
}

func (r *realRenderer) PrintReportSuccess(path string) {
	fmt.Println()
	fmt.Println(text.FgGreen.Sprint("✅ Report generated successfully!"))
	fmt.Println(text.FgHiWhite.Sprintf("📄 Path: %s", path))
}

func (r *realRenderer) PrintFirstDayOfMonthError() {
	fmt.Println()
	fmt.Println(text.FgRed.Sprint("Cost data is not available on the first day of the month. Please try again tomorrow."))
}

func (r *realRenderer) PrintNewVersionAvailable(currentVersion, latestVersion string) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, text.FgHiYellow.Sprintf(
		"A new version of aws-doctor is available: %s → %s. Run 'aws-doctor update' to upgrade.",
		currentVersion, latestVersion,
	))
}

// service is the internal implementation
type service struct {
	format   Format
	renderer Renderer
}

// Service is the interface for the output service.
type Service interface {
	// RenderCostComparison outputs cost comparison data in the configured format
	RenderCostComparison(input model.RenderCostComparisonInput) error

	// RenderTrend outputs trend data in the configured format
	RenderTrend(accountID string, costInfo []model.CostInfo, services []string) error

	// RenderWaste outputs waste report data in the configured format
	RenderWaste(input model.RenderWasteInput) error

	// StopSpinner stops the loading spinner before rendering output
	StopSpinner()

	// SetSpinnerMessage updates the spinner's status text in place
	SetSpinnerMessage(message string)

	// PrintReportSuccess outputs a success message with the report path
	PrintReportSuccess(path string)

	// PrintAlreadyLatest outputs a message when the user is already on the latest version
	PrintAlreadyLatest(version string)

	// PrintHomebrewUpdate outputs a message when the binary was installed via Homebrew
	PrintHomebrewUpdate()

	// PrintGoInstallUpdate outputs a message when the binary was installed via go install
	PrintGoInstallUpdate()

	// PrintRateLimitError outputs a message when GitHub API rate limit is reached
	PrintRateLimitError()

	// PrintUpdateError outputs a message when an update check fails
	PrintUpdateError(err error)

	// RenderVersion outputs the version information
	RenderVersion(versionInfo model.VersionInfo)

	// PrintFirstDayOfMonthError outputs a message when cost data is not available
	PrintFirstDayOfMonthError()

	// PrintNewVersionAvailable outputs a notification when a newer version exists
	PrintNewVersionAvailable(currentVersion, latestVersion string)
}
