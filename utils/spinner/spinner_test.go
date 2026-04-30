package spinner

import (
	"os"
	"testing"
	"time"

	"golang.org/x/term"
)

func TestStartAndStopSpinner(t *testing.T) {
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		t.Skip("skipping: not a TTY, spinner is intentionally disabled in non-interactive mode")
	}

	StartSpinner()
	time.Sleep(100 * time.Millisecond)
	StopSpinner()
}

func TestSpinnerSequence(t *testing.T) {
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		t.Skip("skipping: not a TTY, spinner is intentionally disabled in non-interactive mode")
	}

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

	prev := loader
	loader = nil

	defer func() { loader = prev }()

	StartSpinner()

	if loader != nil {
		t.Error("StartSpinner() should not initialize loader in non-TTY mode")
	}
}
