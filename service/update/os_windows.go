//go:build windows

package update

import "fmt"

func runUpdateScript(runner commandRunner) error {
	if err := runner.Run("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command",
		"irm https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.ps1 | iex"); err != nil {
		return fmt.Errorf("failed to run update script: %w", err)
	}

	return nil
}
