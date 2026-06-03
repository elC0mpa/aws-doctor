package output

// NewRenderer creates a new renderer based on the specified format
func NewRenderer(format string) Renderer {
	switch format {
	case "json":
		return &jsonRenderer{}
	case "csv":
		return &csvRenderer{}
	default:
		return &tableRenderer{}
	}
}
