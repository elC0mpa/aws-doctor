package update

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	installScriptCommand = "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"
)

type realRunner struct{}
type realExecutablePathResolver struct{}

func (r *realRunner) Run(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func (r *realExecutablePathResolver) ExecutablePath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}

	resolvedPath, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolvedPath, nil
	}

	return path, nil
}

// NewService creates a new update service.
func NewService(currentVersion string, releaseService releaseService) Service {
	return &service{
		runner:         &realRunner{},
		releaseService: releaseService,
		pathResolver:   &realExecutablePathResolver{},
		currentVersion: currentVersion,
		stdout:         os.Stdout,
	}
}

func (s *service) Update() error {
	latestRelease, err := s.releaseService.GetLatestRelease(context.Background())
	if err != nil {
		return fmt.Errorf("failed to fetch latest release version: %w", err)
	}

	currentVersion := normalizeVersion(s.currentVersion)
	if currentVersion != "" && currentVersion != "dev" && currentVersion == latestRelease.Version {
		_, _ = fmt.Fprintf(s.stdout, "aws-doctor v%s is already the latest release.\n", latestRelease.Version)

		return nil
	}

	if packageManager, ok, err := s.detectPackageManager(); err != nil {
		return fmt.Errorf("failed to inspect current installation: %w", err)
	} else if ok {
		_, _ = fmt.Fprintf(s.stdout, "aws-doctor was installed via %s. Please update it with `%s`.\n", packageManager, updateCommandFor(packageManager))

		return nil
	}

	// Reuse the install.sh script from the repository for direct script installs.
	if err := s.runner.Run("sh", "-c", installScriptCommand); err != nil {
		return fmt.Errorf("failed to run update script: %w", err)
	}

	return nil
}

func (s *service) detectPackageManager() (string, bool, error) {
	executablePath, err := s.pathResolver.ExecutablePath()
	if err != nil {
		return "", false, err
	}

	normalizedPath := strings.ToLower(filepath.ToSlash(executablePath))

	if strings.Contains(normalizedPath, "/cellar/") || strings.Contains(normalizedPath, "/homebrew/") || strings.Contains(normalizedPath, "/.linuxbrew/") {
		return "Homebrew", true, nil
	}

	return "", false, nil
}

func updateCommandFor(packageManager string) string {
	if packageManager == "Homebrew" {
		return "brew upgrade aws-doctor"
	}

	return "your package manager's upgrade command"
}

func normalizeVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}
