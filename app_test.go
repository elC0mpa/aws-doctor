package main

import (
	"bytes"
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

func TestRunInvalidSubcommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"aws-doctor", "invalid-subcommand"}
	err := run()
	if err == nil {
		t.Error("run() with invalid subcommand should return an error")
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

func TestRunRootShowsHelp(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"aws-doctor"}
	// Root command without RunE and subcommands should just print help via Execute()
	// Cobra returns nil for Execute() when just printing help.
	err := run()
	if err != nil {
		t.Errorf("run() with no args failed: %v", err)
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
