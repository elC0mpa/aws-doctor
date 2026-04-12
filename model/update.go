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
