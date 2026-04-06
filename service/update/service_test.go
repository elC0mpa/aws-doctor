package update

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRunner is a mock implementation of commandRunner
type mockRunner struct {
	mock.Mock
}

func (m *mockRunner) Run(name string, arg ...string) error {
	args := m.Called(name, arg)
	return args.Error(0)
}

type mockReleaseService struct {
	mock.Mock
}

func (m *mockReleaseService) GetLatestRelease(ctx context.Context) (model.ReleaseInfo, error) {
	args := m.Called(ctx)
	releaseValue, _ := args.Get(0).(model.ReleaseInfo)
	return releaseValue, args.Error(1)
}

type mockExecutablePathResolver struct {
	mock.Mock
}

func (m *mockExecutablePathResolver) ExecutablePath() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestNewService(t *testing.T) {
	svc := NewService("v1.2.3", &mockReleaseService{})
	assert.NotNil(t, svc)

	s, ok := svc.(*service)
	assert.True(t, ok)
	assert.NotNil(t, s.runner)
	assert.NotNil(t, s.releaseService)
	assert.NotNil(t, s.pathResolver)
	assert.Equal(t, "v1.2.3", s.currentVersion)
	assert.NotNil(t, s.stdout)
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestRelease  model.ReleaseInfo
		latestErr      error
		executablePath string
		pathErr        error
		runErr         error
		wantErr        string
		wantOutput     string
		expectRun      bool
	}{
		{
			name:           "skips update when current version is latest",
			currentVersion: "v1.2.3",
			latestRelease:  model.ReleaseInfo{Version: "1.2.3", Name: "aws-doctor v1.2.3"},
			executablePath: "/usr/local/bin/aws-doctor",
			wantOutput:     "aws-doctor v1.2.3 is already the latest release.\n",
		},
		{
			name:           "prints brew upgrade guidance for homebrew install",
			currentVersion: "v1.2.2",
			latestRelease:  model.ReleaseInfo{Version: "1.2.3", Name: "aws-doctor v1.2.3"},
			executablePath: "/opt/homebrew/Cellar/aws-doctor/1.2.2/bin/aws-doctor",
			wantOutput:     "aws-doctor was installed via Homebrew. Please update it with `brew upgrade aws-doctor`.\n",
		},
		{
			name:           "runs install script when direct install is outdated",
			currentVersion: "v1.2.2",
			latestRelease:  model.ReleaseInfo{Version: "1.2.3", Name: "aws-doctor v1.2.3"},
			executablePath: "/usr/local/bin/aws-doctor",
			expectRun:      true,
		},
		{
			name:           "returns error when release lookup fails",
			currentVersion: "v1.2.2",
			latestErr:      errors.New("github down"),
			wantErr:        "failed to fetch latest release version: github down",
		},
		{
			name:           "returns error when installation lookup fails",
			currentVersion: "v1.2.2",
			latestRelease:  model.ReleaseInfo{Version: "1.2.3", Name: "aws-doctor v1.2.3"},
			pathErr:        errors.New("cannot resolve executable"),
			wantErr:        "failed to inspect current installation: cannot resolve executable",
		},
		{
			name:           "returns error when install script fails",
			currentVersion: "v1.2.2",
			latestRelease:  model.ReleaseInfo{Version: "1.2.3", Name: "aws-doctor v1.2.3"},
			executablePath: "/usr/local/bin/aws-doctor",
			runErr:         errors.New("execution failed"),
			wantErr:        "failed to run update script: execution failed",
			expectRun:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := new(mockRunner)
			mc := new(mockReleaseService)
			mp := new(mockExecutablePathResolver)
			stdout := &bytes.Buffer{}

			s := &service{
				runner:         mr,
				releaseService: mc,
				pathResolver:   mp,
				currentVersion: tt.currentVersion,
				stdout:         stdout,
			}

			mc.On("GetLatestRelease", mock.Anything).Return(tt.latestRelease, tt.latestErr).Once()

			if tt.latestErr == nil && normalizeVersion(tt.currentVersion) != normalizeVersion(tt.latestRelease.Version) {
				mp.On("ExecutablePath").Return(tt.executablePath, tt.pathErr).Once()
			}

			if tt.expectRun {
				mr.On("Run", "sh", []string{"-c", installScriptCommand}).Return(tt.runErr).Once()
			}

			err := s.Update()

			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantOutput, stdout.String())
			mc.AssertExpectations(t)
			mp.AssertExpectations(t)
			mr.AssertExpectations(t)
		})
	}
}

func TestRealRunner_Run(t *testing.T) {
	// This actually tries to run a command.
	// We'll run something harmless like 'true'.
	r := &realRunner{}
	err := r.Run("go", "version")
	assert.NoError(t, err)
}

func TestRealRunner_Run_Error(t *testing.T) {
	r := &realRunner{}
	// Running a non-existent command should return an error
	err := r.Run("non-existent-command-12345")
	assert.Error(t, err)
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "keeps bare semver", version: "1.2.3", want: "1.2.3"},
		{name: "trims v prefix", version: "v1.2.3", want: "1.2.3"},
		{name: "trims whitespace", version: " v1.2.3\n", want: "1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeVersion(tt.version))
		})
	}
}
