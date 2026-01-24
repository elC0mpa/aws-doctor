package utils //nolint:revive

import (
	"testing"
	"time"
)

func TestStartAndStopSpinner(_ *testing.T) {
	// Simple test to ensure it doesn't panic
	// Note: We can't easily verify output since it writes to stdout/stderr directly
	// and uses ANSI codes.
	StartSpinner()
	time.Sleep(100 * time.Millisecond)
	StopSpinner()
}

func TestSpinnerSequence(_ *testing.T) {
	// Test sequence of start, stop, start, stop
	StartSpinner()
	time.Sleep(50 * time.Millisecond)
	StopSpinner()

	time.Sleep(50 * time.Millisecond)

	StartSpinner()
	time.Sleep(50 * time.Millisecond)
	StopSpinner()
}

func TestStartSpinner_InitializesLoader(t *testing.T) {
	// After calling StartSpinner, the global loader should be non-nil
	StartSpinner()
	defer StopSpinner()

	if loader == nil {
		t.Error("StartSpinner() did not initialize loader")
	}
}
