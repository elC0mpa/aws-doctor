package orchestrator

import (
	"context"
	"errors"
	"time"

	"github.com/elC0mpa/aws-doctor/service/output"
	"github.com/elC0mpa/aws-doctor/utils/spinner"

	"github.com/google/go-github/v62/github"

	"github.com/elC0mpa/aws-doctor/model"
)

type systemService struct {
	cfg SystemConfig
}

// NewSystemService creates a new system orchestrator service.
func NewSystemService(cfg SystemConfig) SystemService {
	return &systemService{cfg: cfg}
}

func (s *systemService) Version() error {
	spinner.StopSpinner()
	output.RenderVersion(s.cfg.VersionInfo)

	return nil
}

func (s *systemService) Update() error {
	spinner.StopSpinner()

	err := s.cfg.UpdateService.Update()
	if err == nil {
		return nil
	}

	if errors.Is(err, model.ErrHomebrewInstall) {
		output.PrintHomebrewUpdate()
		return nil
	}

	if errors.Is(err, model.ErrGoInstall) {
		output.PrintGoInstallUpdate()
		return nil
	}

	if errors.Is(err, model.ErrAlreadyLatest) {
		output.PrintAlreadyLatest(s.cfg.VersionInfo.Version)
		return nil
	}

	if errors.Is(err, model.ErrRateLimit) {
		output.PrintRateLimitError()
		return err
	}

	var rateLimitErr *github.RateLimitError
	if errors.As(err, &rateLimitErr) {
		output.PrintRateLimitError()
		return err
	}

	output.PrintUpdateError(err)

	return err
}

func (s *systemService) CheckForUpdateInBackground() <-chan model.VersionCheckResult {
	versionCh := make(chan model.VersionCheckResult, 1)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		latest, err := s.cfg.UpdateService.CheckForUpdate(ctx)
		versionCh <- model.VersionCheckResult{LatestVersion: latest, Err: err}
	}()

	return versionCh
}
