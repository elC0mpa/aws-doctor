package update

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/version"
	"github.com/google/go-github/v62/github"
)

const homebrewCellarPath = "/Cellar/aws-doctor/"

var goInstallBinPath = string(filepath.Separator) + filepath.Join("go", "bin") + string(filepath.Separator)

type realRunner struct{}

func (r *realRunner) Run(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

type realPathResolver struct{}

func (r *realPathResolver) ResolvedExecutablePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.EvalSymlinks(exePath)
}

// NewService creates a new update service.
func NewService(versionInfo model.VersionInfo) Service {
	client := github.NewClient(nil)

	return &service{
		runner:       &realRunner{},
		versionInfo:  versionInfo,
		repositories: client.Repositories,
		pathResolver: &realPathResolver{},
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

	resolvedPath, err := s.pathResolver.ResolvedExecutablePath()
	if err == nil && strings.Contains(resolvedPath, homebrewCellarPath) {
		return model.ErrHomebrewInstall
	}

	if err == nil && strings.Contains(resolvedPath, goInstallBinPath) {
		return model.ErrGoInstall
	}

	// Proceed with update
	if err := s.runner.Run("sh", "-c", "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"); err != nil {
		return fmt.Errorf("failed to run update script: %w", err)
	}

	return nil
}

func (s *service) CheckForUpdate(ctx context.Context) (*string, error) {
	if s.versionInfo.Version == "dev" {
		return nil, nil
	}

	release, _, err := s.repositories.GetLatestRelease(ctx, model.GitHubOwner, model.GitHubRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	if release == nil || release.TagName == nil {
		return nil, fmt.Errorf("latest release is nil")
	}

	latestVersion := *release.TagName
	if version.IsEqual(latestVersion, s.versionInfo.Version) {
		return nil, nil
	}

	return &latestVersion, nil
}

func (s *service) shouldUpdate(ctx context.Context) (bool, error) {
	if s.versionInfo.Version == "dev" {
		return true, nil
	}

	latest, err := s.CheckForUpdate(ctx)
	if err != nil {
		return false, err
	}

	return latest != nil, nil
}
