package output

import (
	"fmt"
	"os"
	"runtime"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/jedib0t/go-pretty/v6/text"
)

// PrintAlreadyLatest outputs a message when the user is already on the latest version
func PrintAlreadyLatest(version string) {
	fmt.Printf("aws-doctor %s is already the latest version.\n", version)
}

// PrintHomebrewUpdate outputs a message when the binary was installed via Homebrew
func PrintHomebrewUpdate() {
	fmt.Println("It appears aws-doctor was installed via Homebrew.")
	fmt.Println("To update, please run:")
	fmt.Println(text.FgCyan.Sprint("brew upgrade elC0mpa/tap/aws-doctor"))
}

// PrintGoInstallUpdate outputs a message when the binary was installed via go install
func PrintGoInstallUpdate() {
	fmt.Println("It appears aws-doctor was installed via go install.")
	fmt.Println("To update, please run:")
	fmt.Println(text.FgCyan.Sprint("go install github.com/elC0mpa/aws-doctor@latest"))
}

// PrintRateLimitError outputs a message when GitHub API rate limit is reached
func PrintRateLimitError() {
	fmt.Fprintln(os.Stderr, text.FgRed.Sprint("GitHub API rate limit exceeded. Please try again later."))
}

// PrintUpdateError outputs a message when an update check fails
func PrintUpdateError(err error) {
	fmt.Fprint(os.Stderr, text.FgRed.Sprintf("Update failed: %v\n", err))
}

// RenderVersion outputs the version information
func RenderVersion(versionInfo model.VersionInfo) {
	fmt.Printf("aws-doctor version %s\n", versionInfo.Version)
	fmt.Printf("commit: %s\n", versionInfo.Commit)
	fmt.Printf("built at: %s\n", versionInfo.Date)
	fmt.Printf("os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// PrintWasteError outputs a message when an interactive waste rendering fails
func PrintWasteError(err error) {
	fmt.Fprintf(os.Stderr, "Error rendering waste interactive: %v\n", err)
}

// PrintReportSuccess outputs a success message with the report path
func PrintReportSuccess(path string) {
	fmt.Println()
	fmt.Println(text.FgGreen.Sprint("✅ Report generated successfully!"))
	fmt.Println(text.FgHiWhite.Sprintf("📄 Path: %s", path))
}

// PrintFirstDayOfMonthError outputs a message when cost data is not available
func PrintFirstDayOfMonthError() {
	fmt.Println()
	fmt.Println(text.FgRed.Sprint("Cost data is not available on the first day of the month. Please try again tomorrow."))
}

// PrintNewVersionAvailable outputs a notification when a newer version exists
func PrintNewVersionAvailable(currentVersion, latestVersion string) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, text.FgHiYellow.Sprintf(
		"A new version of aws-doctor is available: %s → %s. Run 'aws-doctor update' to upgrade.",
		currentVersion, latestVersion,
	))
}
