package model

import "errors"

const (
	// GitHubOwner is the owner of the aws-doctor repository.
	GitHubOwner = "elC0mpa"
	// GitHubRepo is the name of the aws-doctor repository.
	GitHubRepo = "aws-doctor"
)

// ErrAlreadyLatest is returned when the current version is already the latest version.
var ErrAlreadyLatest = errors.New("already the latest version")

// ErrHomebrewInstall is returned when the binary was installed via Homebrew.
var ErrHomebrewInstall = errors.New("installed via Homebrew")

// ErrGoInstall is returned when the binary was installed via go install.
var ErrGoInstall = errors.New("installed via go install")

// ErrRateLimit is returned when the GitHub API rate limit is exceeded.
var ErrRateLimit = errors.New("github rate limit exceeded")

// VersionCheckResult holds the outcome of a background version check.
type VersionCheckResult struct {
	LatestVersion *string
	Err           error
}
