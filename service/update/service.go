package update

import (
	"fmt"
	"os"
	"os/exec"
)

// NewService creates a new update service.
func NewService() Service {
	return &service{}
}

func (s *service) Update() error {
	fmt.Println("Updating aws-doctor to the latest version...")

	// Reutilize the install.sh script from the repository
	cmd := exec.Command("sh", "-c", "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run update script: %w", err)
	}

	return nil
}
