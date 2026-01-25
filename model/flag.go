package model

// Flags represents the command-line flags for the application.
type Flags struct {
	Region  string
	Profile string
	Trend   bool
	Waste   bool
	Output  string // Output format: "table" (default) or "json"
}
