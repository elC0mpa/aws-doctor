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

func TestRunVersionJSON(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"aws-doctor", "version", "--output", "json"}
	err := run()
	if err != nil {
		t.Errorf("run() with version and --output json failed: %v", err)
	}
}

func TestRunInvalidFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"aws-doctor", "--invalid-flag"}
	err := run()
	if err == nil {
		t.Error("run() with invalid flag should return an error")
	}
}

func TestRunVersion(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"aws-doctor", "version"}
	err := run()
	if err != nil {
		t.Errorf("run() with version failed: %v", err)
	}
}

func TestVersionOutput(t *testing.T) {
	tmpBinary := t.TempDir() + "/aws-doctor-test"

	cmd := exec.Command("go", "build", "-o", tmpBinary, "./app.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	cmdRun := exec.Command(tmpBinary, "version")

	var stdout bytes.Buffer
	cmdRun.Stdout = &stdout

	err := cmdRun.Run()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()

	if !strings.Contains(output, "aws-doctor version") {
		t.Errorf("Output should contain 'aws-doctor version', got: %s", output)
	}

	if !strings.Contains(output, "commit:") {
		t.Errorf("Output should contain 'commit:', got: %s", output)
	}

	if !strings.Contains(output, "built at:") {
		t.Errorf("Output should contain 'built at:', got: %s", output)
	}
}

func TestVersionWithLdflags(t *testing.T) {
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

	cmdRun := exec.Command(tmpBinary, "version")
	var stdout bytes.Buffer
	cmdRun.Stdout = &stdout

	err := cmdRun.Run()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()

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
	tmpBinary := t.TempDir() + "/aws-doctor-test"

	cmd := exec.Command("go", "build", "-o", tmpBinary, "./app.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	cmdRun := exec.Command(tmpBinary, "version")
	cmdRun.Stdout = os.Stdout

	err := cmdRun.Run()
	if err != nil {
		t.Errorf("version should exit with code 0, got error: %v", err)
	}
}

func TestVersionDoesNotShowBanner(t *testing.T) {
	tmpBinary := t.TempDir() + "/aws-doctor-test"

	cmd := exec.Command("go", "build", "-o", tmpBinary, "./app.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	cmdRun := exec.Command(tmpBinary, "version")
	var stdout bytes.Buffer
	cmdRun.Stdout = &stdout

	err := cmdRun.Run()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()

	if strings.Contains(output, "___") {
		t.Errorf("version output should not contain the ASCII banner")
	}
}
