package utils

import (
	"time"

	"github.com/briandowns/spinner"
)

var loader *spinner.Spinner

func StartSpinner() {
	loader = spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	loader.Color("yellow")
	loader.Suffix = " Please wait while data is being fetched..."
	loader.Start()
}

func StopSpinner() {
	if loader != nil {
		loader.Stop()
	}
}
