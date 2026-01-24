package utils //nolint:revive

import (
	"time"

	"github.com/briandowns/spinner"
)

var loader *spinner.Spinner

// StartSpinner starts the CLI loading spinner.
func StartSpinner() {
	loader = spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	loader.Color("yellow") //nolint:errcheck
	loader.Suffix = " Please wait while data is being fetched..."
	loader.Start()
}

// StopSpinner stops the CLI loading spinner.
func StopSpinner() {
	if loader != nil {
		loader.Stop()
	}
}
