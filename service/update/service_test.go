package update

import (
	"context"
	"errors"
	"testing"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/google/go-github/v62/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRunner is a mock implementation of commandRunner
type mockRunner struct {
	mock.Mock
}

const tag = "v1.2.3"

func (m *mockRunner) Run(name string, arg ...string) error {
	args := m.Called(name, arg)
	return args.Error(0)
}

// mockRepositories is a mock implementation of repositoryService
type mockRepositories struct {
	mock.Mock
}

func (m *mockRepositories) GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
	args := m.Called(ctx, owner, repo)

	var rel *github.RepositoryRelease
	if args.Get(0) != nil {
		rel = args.Get(0).(*github.RepositoryRelease)
	}

	var resp *github.Response
	if args.Get(1) != nil {
		resp = args.Get(1).(*github.Response)
	}

	return rel, resp, args.Error(2)
}

func TestNewService(t *testing.T) {
	v := model.VersionInfo{Version: "v1.0.0"}
	svc := NewService(v)
	assert.NotNil(t, svc)

	s, ok := svc.(*service)
	assert.True(t, ok)
	assert.NotNil(t, s.runner)
	assert.Equal(t, v, s.versionInfo)
}

func TestUpdate_AlreadyLatest(t *testing.T) {
	mr := new(mockRunner)
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: tag}
	s := &service{runner: mr, repositories: mrepo, versionInfo: v}

	tagName := tag
	release := &github.RepositoryRelease{TagName: &tagName}
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(release, nil, nil)

	err := s.Update()
	assert.ErrorIs(t, err, model.ErrAlreadyLatest)
	mr.AssertNotCalled(t, "Run", mock.Anything, mock.Anything)
}

func TestUpdate_AlreadyLatestVPrefix(t *testing.T) {
	mr := new(mockRunner)
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: "1.2.3"}
	s := &service{runner: mr, repositories: mrepo, versionInfo: v}

	tagName := tag
	release := &github.RepositoryRelease{TagName: &tagName}
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(release, nil, nil)

	err := s.Update()
	assert.ErrorIs(t, err, model.ErrAlreadyLatest)
	mr.AssertNotCalled(t, "Run", mock.Anything, mock.Anything)
}

func TestUpdate_DevVersion(t *testing.T) {
	mr := new(mockRunner)
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: "dev"}
	s := &service{runner: mr, repositories: mrepo, versionInfo: v}

	// No GetLatestRelease expectation here as it should short-circuit

	mr.On("Run", "sh", []string{"-c", "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"}).Return(nil)

	err := s.Update()
	assert.NoError(t, err)
	mr.AssertExpectations(t)
	mrepo.AssertNotCalled(t, "GetLatestRelease", mock.Anything, mock.Anything, mock.Anything)
}

func TestUpdate_NeedsUpdate(t *testing.T) {
	mr := new(mockRunner)
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: "v1.2.2"}
	s := &service{runner: mr, repositories: mrepo, versionInfo: v}

	tagName := tag
	release := &github.RepositoryRelease{TagName: &tagName}
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(release, nil, nil)

	mr.On("Run", "sh", []string{"-c", "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"}).Return(nil)

	err := s.Update()
	assert.NoError(t, err)
	mr.AssertExpectations(t)
}

func TestUpdate_RateLimitError(t *testing.T) {
	mr := new(mockRunner)
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: tag}
	s := &service{runner: mr, repositories: mrepo, versionInfo: v}

	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(nil, nil, &github.RateLimitError{})

	err := s.Update()
	assert.Error(t, err)

	var rateLimitErr *github.RateLimitError
	assert.True(t, errors.As(err, &rateLimitErr))
	mr.AssertNotCalled(t, "Run", mock.Anything, mock.Anything)
}

func TestUpdate_FetchError(t *testing.T) {
	mr := new(mockRunner)
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: tag}
	s := &service{runner: mr, repositories: mrepo, versionInfo: v}

	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(nil, nil, errors.New("github error"))

	err := s.Update()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch latest release")
}

func TestUpdate_ExecutionError(t *testing.T) {
	mr := new(mockRunner)
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: "v1.2.2"}
	s := &service{runner: mr, repositories: mrepo, versionInfo: v}

	tagName := tag
	release := &github.RepositoryRelease{TagName: &tagName}
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(release, nil, nil)

	mr.On("Run", "sh", []string{"-c", "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"}).Return(errors.New("execution failed"))

	err := s.Update()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to run update script")
}

func TestCheckForUpdate_DevVersion(t *testing.T) {
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: "dev"}
	s := &service{repositories: mrepo, versionInfo: v}

	result, err := s.CheckForUpdate(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, result)
	mrepo.AssertNotCalled(t, "GetLatestRelease", mock.Anything, mock.Anything, mock.Anything)
}

func TestCheckForUpdate_AlreadyLatest(t *testing.T) {
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: tag}
	s := &service{repositories: mrepo, versionInfo: v}

	tagName := tag
	release := &github.RepositoryRelease{TagName: &tagName}
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(release, nil, nil)

	result, err := s.CheckForUpdate(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestCheckForUpdate_NewVersionAvailable(t *testing.T) {
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: "v1.2.2"}
	s := &service{repositories: mrepo, versionInfo: v}

	tagName := tag
	release := &github.RepositoryRelease{TagName: &tagName}
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(release, nil, nil)

	result, err := s.CheckForUpdate(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, tag, *result)
}

func TestCheckForUpdate_GitHubError(t *testing.T) {
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: tag}
	s := &service{repositories: mrepo, versionInfo: v}

	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(nil, nil, errors.New("github error"))

	result, err := s.CheckForUpdate(context.Background())
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to fetch latest release")
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
