package model

type Flags struct {
	Region  string
	Profile string
	Trend   bool
	Waste   bool
	Daily   bool   // Show daily cost analysis for last 30 days
	Output  string // Output format: "table" (default) or "json"
}
