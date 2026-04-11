package update

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/version"
	"github.com/google/go-github/v62/github"
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
func NewService(versionInfo model.VersionInfo) Service {
	client := github.NewClient(nil)

	return &service{
		runner:       &realRunner{},
		versionInfo:  versionInfo,
		repositories: client.Repositories,
	}
}

func (s *service) Update() error {
	shouldUpdate, err := s.shouldUpdate(context.Background())
	if err != nil {
		return err
	}

	if !shouldUpdate {
		return model.ErrAlreadyLatest
	}

	// Proceed with update
	// Reutilize the install.sh script from the repository
	if err := s.runner.Run("sh", "-c", "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"); err != nil {
		return fmt.Errorf("failed to run update script: %w", err)
	}

	return nil
}

func (s *service) shouldUpdate(ctx context.Context) (bool, error) {
	if s.versionInfo.Version == "dev" {
		return true, nil
	}

	release, _, err := s.repositories.GetLatestRelease(ctx, model.GitHubOwner, model.GitHubRepo)
	if err != nil {
		return false, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	latestVersion := *release.TagName
	equalVersions := version.IsEqual(latestVersion, s.versionInfo.Version)

	return !equalVersions, nil
}
