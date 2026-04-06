// Package release provides access to published GitHub releases for aws-doctor.
package release

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/elC0mpa/aws-doctor/model"
)

const latestReleaseURL = "https://api.github.com/repos/elC0mpa/aws-doctor/releases/latest"

var releaseVersionPattern = regexp.MustCompile(`v?\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?`)

type realGitHubClient struct {
	client *http.Client
}

type latestReleaseResponse struct {
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// NewService creates a new release service.
func NewService() Service {
	return &service{
		client: &realGitHubClient{
			client: &http.Client{
				Timeout: 10 * time.Second,
			},
		},
	}
}

func (s *service) GetLatestRelease(ctx context.Context) (model.ReleaseInfo, error) {
	release, err := s.client.LatestRelease(ctx)
	if err != nil {
		return model.ReleaseInfo{}, err
	}

	version, err := versionFromRelease(release)
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

	if strings.TrimSpace(payload.Name) == "" && strings.TrimSpace(payload.TagName) == "" {
		return githubRelease{}, fmt.Errorf("latest release name and tag are empty")
	}

	return githubRelease{
		Name:    payload.Name,
		TagName: payload.TagName,
		URL:     payload.HTMLURL,
	}, nil
}

func versionFromRelease(release githubRelease) (string, error) {
	source := strings.TrimSpace(release.Name)
	if source == "" {
		source = strings.TrimSpace(release.TagName)
	}

	match := releaseVersionPattern.FindString(source)
	if match == "" {
		return "", fmt.Errorf("latest release metadata does not include a version: name=%q tag=%q", release.Name, release.TagName)
	}

	return normalizeVersion(match), nil
}

func normalizeVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}
