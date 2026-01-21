package model

type Flags struct {
	Region  string
	Profile string
	Trend   bool
	Waste   bool
	Output  string // Output format: "table" (default) or "json"
}
