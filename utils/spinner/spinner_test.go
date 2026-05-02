package spinner

import (
	"testing"
	"time"
)

func TestStartAndStopSpinner(t *testing.T) {
	// Mock terminal
	orig := isTerminal
	isTerminal = func(fd int) bool { return true }

	defer func() { isTerminal = orig }()

	StartSpinner()
	time.Sleep(100 * time.Millisecond)
	StopSpinner()
}

func TestSpinnerSequence(t *testing.T) {
	// Mock terminal
	orig := isTerminal
	isTerminal = func(fd int) bool { return true }

	defer func() { isTerminal = orig }()

	StartSpinner()
	time.Sleep(50 * time.Millisecond)
	StopSpinner()

	time.Sleep(50 * time.Millisecond)

	StartSpinner()
	time.Sleep(50 * time.Millisecond)
	StopSpinner()
}

func TestSetMessage(t *testing.T) {
	// Mock terminal
	orig := isTerminal
	isTerminal = func(fd int) bool { return true }

	defer func() { isTerminal = orig }()

	// Should not panic even if spinner is not started
	SetMessage("Test message")

	StartSpinner()

	defer StopSpinner()

	SetMessage("Another message")
}

func TestStartSpinner_InitializesLoader(t *testing.T) {
	// Mock terminal
	orig := isTerminal
	isTerminal = func(fd int) bool { return true }

	defer func() { isTerminal = orig }()

	StartSpinner()

	defer StopSpinner()

	if loader == nil {
		t.Error("StartSpinner() did not initialize loader")
	}
}

func TestStartSpinner_NonTTY_LoaderNil(t *testing.T) {
	// Mock non-terminal
	orig := isTerminal
	isTerminal = func(fd int) bool { return false }

	defer func() { isTerminal = orig }()

	prev := loader
	loader = nil

	defer func() { loader = prev }()

	StartSpinner()

	if loader != nil {
		t.Error("StartSpinner() should not initialize loader in non-TTY mode")
	}
}
