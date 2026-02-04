package update

import (
	"fmt"
	"os"
	"os/exec"
)

type realRunner struct{}

func (r *realRunner) Run(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// NewService creates a new update service.
func NewService() Service {
	return &service{
		runner: &realRunner{},
	}
}

func (s *service) Update() error {
	// Reutilize the install.sh script from the repository
	if err := s.runner.Run("sh", "-c", "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"); err != nil {
		return fmt.Errorf("failed to run update script: %w", err)
	}

	return nil
}
