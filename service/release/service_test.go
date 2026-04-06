package release

import (
	"context"
	"errors"
	"testing"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockGitHubClient struct {
	mock.Mock
}

func (m *mockGitHubClient) LatestRelease(ctx context.Context) (githubRelease, error) {
	args := m.Called(ctx)

	releaseValue, _ := args.Get(0).(githubRelease)

	return releaseValue, args.Error(1)
}

func TestNewService(t *testing.T) {
	svc := NewService()
	assert.NotNil(t, svc)

	impl, ok := svc.(*service)
	assert.True(t, ok)
	assert.NotNil(t, impl.client)
}

func TestGetLatestRelease(t *testing.T) {
	tests := []struct {
		name        string
		release     githubRelease
		clientErr   error
		want        model.ReleaseInfo
		wantErrText string
	}{
		{
			name: "returns version from release name",
			release: githubRelease{
				Name: "aws-doctor v2.6.4",
				URL:  "https://github.com/elC0mpa/aws-doctor/releases/tag/v2.6.4",
			},
			want: model.ReleaseInfo{
				Version: "2.6.4",
				Name:    "aws-doctor v2.6.4",
				URL:     "https://github.com/elC0mpa/aws-doctor/releases/tag/v2.6.4",
			},
		},
		{
			name:        "returns client error",
			clientErr:   errors.New("network down"),
			wantErrText: "network down",
		},
		{
			name: "errors when release name has no version",
			release: githubRelease{
				Name: "aws-doctor stable",
			},
			wantErrText: `latest release name does not include a version: "aws-doctor stable"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockGitHubClient)
			svc := &service{client: mc}

			mc.On("LatestRelease", mock.Anything).Return(tt.release, tt.clientErr).Once()

			got, err := svc.GetLatestRelease(context.Background())

			if tt.wantErrText != "" {
				assert.EqualError(t, err, tt.wantErrText)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mc.AssertExpectations(t)
		})
	}
}

func TestVersionFromReleaseName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        string
		wantErrText string
	}{
		{name: "extracts version with prefix", input: "aws-doctor v2.6.4", want: "2.6.4"},
		{name: "extracts prerelease version", input: "Release v2.6.4-beta.1", want: "2.6.4-beta.1"},
		{name: "errors when version missing", input: "latest stable", wantErrText: `latest release name does not include a version: "latest stable"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := versionFromReleaseName(tt.input)

			if tt.wantErrText != "" {
				assert.EqualError(t, err, tt.wantErrText)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
