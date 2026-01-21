package utils

import (
	"testing"
	"time"
)

func TestStartAndStopSpinner(t *testing.T) {
	// Start spinner
	StartSpinner()

	// Give it a moment to actually start
	time.Sleep(50 * time.Millisecond)

	// Stop spinner - should not panic
	StopSpinner()
}

func TestSpinnerSequence(t *testing.T) {
	// Test multiple start/stop cycles
	for i := 0; i < 3; i++ {
		StartSpinner()
		time.Sleep(10 * time.Millisecond)
		StopSpinner()
	}
}

func TestStartSpinner_InitializesLoader(t *testing.T) {
	// After calling StartSpinner, the global loader should be non-nil
	StartSpinner()
	defer StopSpinner()

	if loader == nil {
		t.Error("StartSpinner() did not initialize loader")
	}
}
