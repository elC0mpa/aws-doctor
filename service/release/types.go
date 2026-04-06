package release

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
)

type service struct {
	client githubClient
}

type githubClient interface {
	LatestRelease(ctx context.Context) (githubRelease, error)
}

type githubRelease struct {
	Name    string
	TagName string
	URL     string
}

// Service fetches published release metadata.
type Service interface {
	GetLatestRelease(ctx context.Context) (model.ReleaseInfo, error)
}
