//go:build !windows

package update

import "fmt"

func runUpdateScript(runner commandRunner) error {
	if err := runner.Run("sh", "-c", "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"); err != nil {
		return fmt.Errorf("failed to run update script: %w", err)
	}

	return nil
}
