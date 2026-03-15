package cmd

import (
	"testing"
)

func TestExecuteVersion(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})
	
	// Temporarily capture output if needed, but here we just ensure it doesn't panic
	err := Execute("dev", "none", "unknown")
	if err != nil {
		t.Errorf("Execute(version) failed: %v", err)
	}
}

func TestExecuteUpdate(t *testing.T) {
	rootCmd.SetArgs([]string{"update"})
	
	// Update will attempt to self-update, which might fail in tests, so we skip actual execution
	// But we can verify the command exists
	c, _, err := rootCmd.Find([]string{"update"})
	if err != nil || c.Name() != "update" {
		t.Errorf("update command not found")
	}
}

func TestExecuteTrendCommandExists(t *testing.T) {
	c, _, err := rootCmd.Find([]string{"trend"})
	if err != nil || c.Name() != "trend" {
		t.Errorf("trend command not found")
	}
}

func TestExecuteWasteCommandExists(t *testing.T) {
	c, _, err := rootCmd.Find([]string{"waste"})
	if err != nil || c.Name() != "waste" {
		t.Errorf("waste command not found")
	}
}
