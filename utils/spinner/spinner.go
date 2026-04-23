package spinner

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
)

var loader *spinner.Spinner

// StartSpinner starts the CLI loading spinner.
func StartSpinner() {
	loader = spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
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

// SetMessage updates the spinner suffix in place so callers can surface what phase of work is
// currently running without stopping and restarting the spinner.
func SetMessage(message string) {
	if loader != nil {
		loader.Suffix = " " + message
	}
}
