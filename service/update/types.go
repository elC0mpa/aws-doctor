package update

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/google/go-github/v62/github"
)

type commandRunner interface {
	Run(name string, arg ...string) error
}

type repositoryService interface {
	GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)
}

type executablePathResolver interface {
	ResolvedExecutablePath() (string, error)
}

type service struct {
	runner       commandRunner
	versionInfo  model.VersionInfo
	repositories repositoryService
	pathResolver executablePathResolver
}

// Service is the interface for the update service.
type Service interface {
	Update() error
	CheckForUpdate(ctx context.Context) (*string, error)
}
