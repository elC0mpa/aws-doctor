package update

import (
	"context"
	"io"

	"github.com/elC0mpa/aws-doctor/model"
)

// Service is the interface for the update service.
type Service interface {
	Update() error
}

type commandRunner interface {
	Run(name string, arg ...string) error
}

type releaseService interface {
	GetLatestRelease(ctx context.Context) (model.ReleaseInfo, error)
}

type executablePathResolver interface {
	ExecutablePath() (string, error)
}

type service struct {
	runner         commandRunner
	releaseService releaseService
	pathResolver   executablePathResolver
	currentVersion string
	stdout         io.Writer
}
