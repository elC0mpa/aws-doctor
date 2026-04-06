// Package release provides access to published GitHub releases for aws-doctor.
package release

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/elC0mpa/aws-doctor/model"
)

const latestReleaseURL = "https://api.github.com/repos/elC0mpa/aws-doctor/releases/latest"

var releaseVersionPattern = regexp.MustCompile(`v?\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?`)

type realGitHubClient struct {
	client *http.Client
}

type latestReleaseResponse struct {
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// NewService creates a new release service.
func NewService() Service {
	return &service{
		client: &realGitHubClient{
			client: http.DefaultClient,
		},
	}
}

func (s *service) GetLatestRelease(ctx context.Context) (model.ReleaseInfo, error) {
	release, err := s.client.LatestRelease(ctx)
	if err != nil {
		return model.ReleaseInfo{}, err
	}

	version, err := versionFromReleaseName(release.Name)
	if err != nil {
		return model.ReleaseInfo{}, err
	}

	return model.ReleaseInfo{
		Version: version,
		Name:    release.Name,
		URL:     release.URL,
	}, nil
}

func (c *realGitHubClient) LatestRelease(ctx context.Context) (githubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return githubRelease{}, fmt.Errorf("create latest release request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "aws-doctor")

	resp, err := c.client.Do(req)
	if err != nil {
		return githubRelease{}, fmt.Errorf("request latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return githubRelease{}, fmt.Errorf("unexpected GitHub status: %s", resp.Status)
	}

	var payload latestReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return githubRelease{}, fmt.Errorf("decode latest release response: %w", err)
	}

	if strings.TrimSpace(payload.Name) == "" {
		return githubRelease{}, fmt.Errorf("latest release name is empty")
	}

	return githubRelease{
		Name: payload.Name,
		URL:  payload.HTMLURL,
	}, nil
}

func versionFromReleaseName(name string) (string, error) {
	match := releaseVersionPattern.FindString(strings.TrimSpace(name))
	if match == "" {
		return "", fmt.Errorf("latest release name does not include a version: %q", name)
	}

	return normalizeVersion(match), nil
}

func normalizeVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}
