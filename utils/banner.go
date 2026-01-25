package utils //nolint:revive

import "github.com/common-nighthawk/go-figure"

// DrawBanner prints the application banner to stdout.
func DrawBanner() {
	myFigure := figure.NewColorFigure("AWS Doctor", "isometric3", "yellow", false)
	myFigure.Print()
}
