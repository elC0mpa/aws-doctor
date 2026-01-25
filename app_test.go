package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestVersionVariablesHaveDefaults(t *testing.T) {
	// Verify default values are set for development builds
	if version == "" {
		t.Error("version variable should have a default value")
	}

	if commit == "" {
		t.Error("commit variable should have a default value")
	}

	if date == "" {
		t.Error("date variable should have a default value")
	}
}

func TestVersionOutput(t *testing.T) {
	// Build the binary
	tmpBinary := t.TempDir() + "/aws-doctor-test"

	cmd := exec.Command("go", "build", "-o", tmpBinary, "./app.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	tests := []struct {
		name string
		flag string
	}{
		{"double dash version", "--version"},
		{"single dash version", "-version"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(tmpBinary, tt.flag)

			var stdout bytes.Buffer

			cmd.Stdout = &stdout

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}

			output := stdout.String()

			// Check that version info is present
			if !strings.Contains(output, "aws-doctor version") {
				t.Errorf("Output should contain 'aws-doctor version', got: %s", output)
			}

			if !strings.Contains(output, "commit:") {
				t.Errorf("Output should contain 'commit:', got: %s", output)
			}

			if !strings.Contains(output, "built at:") {
				t.Errorf("Output should contain 'built at:', got: %s", output)
			}
		})
	}
}

func TestVersionWithLdflags(t *testing.T) {
	// Build the binary with custom ldflags
	tmpBinary := t.TempDir() + "/aws-doctor-test"
	testVersion := "1.2.3"
	testCommit := "abc123def"
	testDate := "2026-01-21T12:00:00Z"

	ldflags := fmt.Sprintf("-X main.version=%s -X main.commit=%s -X main.date=%s",
		testVersion, testCommit, testDate)

	cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", tmpBinary, "./app.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Run with --version
	cmd = exec.Command(tmpBinary, "--version")

	var stdout bytes.Buffer

	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()

	// Check that injected values are present
	if !strings.Contains(output, testVersion) {
		t.Errorf("Output should contain version '%s', got: %s", testVersion, output)
	}

	if !strings.Contains(output, testCommit) {
		t.Errorf("Output should contain commit '%s', got: %s", testCommit, output)
	}

	if !strings.Contains(output, testDate) {
		t.Errorf("Output should contain date '%s', got: %s", testDate, output)
	}
}

func TestVersionExitsCleanly(t *testing.T) {
	// Build the binary
	tmpBinary := t.TempDir() + "/aws-doctor-test"

	cmd := exec.Command("go", "build", "-o", tmpBinary, "./app.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Run with --version and check exit code
	cmd = exec.Command(tmpBinary, "--version")
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		t.Errorf("--version should exit with code 0, got error: %v", err)
	}
}

func TestVersionDoesNotShowBanner(t *testing.T) {
	// Build the binary
	tmpBinary := t.TempDir() + "/aws-doctor-test"

	cmd := exec.Command("go", "build", "-o", tmpBinary, "./app.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Run with --version
	cmd = exec.Command(tmpBinary, "--version")

	var stdout bytes.Buffer

	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()

	// Banner contains ASCII art - should NOT be present
	if strings.Contains(output, "___") {
		t.Errorf("--version output should not contain the ASCII banner")
	}
}
