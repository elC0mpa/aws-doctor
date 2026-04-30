package spinner

import (
	"os"
	"testing"
	"time"

	"golang.org/x/term"
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

func TestSetMessage(t *testing.T) {
	// Should not panic even if spinner is not started
	SetMessage("Test message")

	StartSpinner()

	defer StopSpinner()

	SetMessage("Another message")
}

func TestStartSpinner_InitializesLoader(t *testing.T) {
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		t.Skip("skipping: not a TTY, spinner is intentionally disabled in non-interactive mode")
	}

	StartSpinner()

	defer StopSpinner()

	if loader == nil {
		t.Error("StartSpinner() did not initialize loader")
	}
}

func TestStartSpinner_NonTTY_LoaderNil(t *testing.T) {
	if term.IsTerminal(int(os.Stderr.Fd())) {
		t.Skip("skipping: stderr is a TTY, spinner will be initialized")
	}

	loader = nil

	StartSpinner()

	if loader != nil {
		t.Error("StartSpinner() should not initialize loader in non-TTY mode")
	}
}
