package output

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"runtime"

	"github.com/elC0mpa/aws-doctor/model"
)

// captureOutput intercepts stdout and stderr for testing
func captureOutput(f func()) (string, string) {
	origStdout := os.Stdout
	origStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	outC := make(chan string)
	errC := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		outC <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rErr)
		errC <- buf.String()
	}()

	f()

	wOut.Close()
	wErr.Close()

	os.Stdout = origStdout
	os.Stderr = origStderr

	stdout := <-outC
	stderr := <-errC

	return stdout, stderr
}

func TestPrintAlreadyLatest(t *testing.T) {
	stdout, _ := captureOutput(func() {
		PrintAlreadyLatest("v1.2.3")
	})
	
	if !strings.Contains(stdout, "v1.2.3") {
		t.Errorf("Expected stdout to contain 'v1.2.3', got: %s", stdout)
	}
}

func TestPrintHomebrewUpdate(t *testing.T) {
	stdout, _ := captureOutput(func() {
		PrintHomebrewUpdate()
	})
	
	if !strings.Contains(stdout, "Homebrew") {
		t.Errorf("Expected stdout to contain 'Homebrew', got: %s", stdout)
	}
}

func TestPrintGoInstallUpdate(t *testing.T) {
	stdout, _ := captureOutput(func() {
		PrintGoInstallUpdate()
	})
	
	if !strings.Contains(stdout, "go install") {
		t.Errorf("Expected stdout to contain 'go install', got: %s", stdout)
	}
}

func TestPrintRateLimitError(t *testing.T) {
	_, stderr := captureOutput(func() {
		PrintRateLimitError()
	})
	
	if !strings.Contains(stderr, "rate limit exceeded") {
		t.Errorf("Expected stderr to contain 'rate limit exceeded', got: %s", stderr)
	}
}

func TestPrintUpdateError(t *testing.T) {
	_, stderr := captureOutput(func() {
		PrintUpdateError(errors.New("test error"))
	})
	
	if !strings.Contains(stderr, "test error") {
		t.Errorf("Expected stderr to contain 'test error', got: %s", stderr)
	}
}

func TestRenderVersion(t *testing.T) {
	stdout, _ := captureOutput(func() {
		RenderVersion(model.VersionInfo{
			Version: "v1.0.0",
			Commit:  "abcdef",
			Date:    "2023-10-27",
		})
	})
	
	if !strings.Contains(stdout, "v1.0.0") {
		t.Errorf("Expected stdout to contain 'v1.0.0', got: %s", stdout)
	}
	if !strings.Contains(stdout, "abcdef") {
		t.Errorf("Expected stdout to contain 'abcdef', got: %s", stdout)
	}
	if !strings.Contains(stdout, runtime.GOOS) {
		t.Errorf("Expected stdout to contain '%s', got: %s", runtime.GOOS, stdout)
	}
}

func TestPrintWasteError(t *testing.T) {
	_, stderr := captureOutput(func() {
		PrintWasteError(errors.New("waste error"))
	})
	
	if !strings.Contains(stderr, "waste error") {
		t.Errorf("Expected stderr to contain 'waste error', got: %s", stderr)
	}
}

func TestPrintReportSuccess(t *testing.T) {
	stdout, _ := captureOutput(func() {
		PrintReportSuccess("/path/to/report")
	})
	
	if !strings.Contains(stdout, "/path/to/report") {
		t.Errorf("Expected stdout to contain '/path/to/report', got: %s", stdout)
	}
}

func TestPrintFirstDayOfMonthError(t *testing.T) {
	stdout, _ := captureOutput(func() {
		PrintFirstDayOfMonthError()
	})
	
	if !strings.Contains(stdout, "first day of the month") {
		t.Errorf("Expected stdout to contain 'first day of the month', got: %s", stdout)
	}
}

func TestPrintNewVersionAvailable(t *testing.T) {
	_, stderr := captureOutput(func() {
		PrintNewVersionAvailable("v1.0.0", "v1.1.0")
	})
	
	if !strings.Contains(stderr, "v1.0.0") || !strings.Contains(stderr, "v1.1.0") {
		t.Errorf("Expected stderr to contain both versions, got: %s", stderr)
	}
}
