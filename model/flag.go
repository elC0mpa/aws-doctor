package model

// Flags represents the command-line flags for the application.
type Flags struct {
	Region      string
	Profile     string
	Trend       bool
	TrendChecks []string
	Waste       bool
	WasteChecks []string
	Version     bool
	Update      bool
	Output      string // Output format: "table" (default) or "json"
	Report      bool   // Whether to generate a report
	ReportPath  string // Path to save the report (optional)
}
