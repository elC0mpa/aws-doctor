package update

import (
	"context"
	"errors"
	"testing"

	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/cache"
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

// mockPathResolver is a mock implementation of executablePathResolver
type mockPathResolver struct {
	mock.Mock
}

func (m *mockPathResolver) ResolvedExecutablePath() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestNewService(t *testing.T) {
	v := model.VersionInfo{Version: "v1.0.0"}
	mc := new(services.MockCacheService)
	svc := NewService(v, mc)
	assert.NotNil(t, svc)

	s, ok := svc.(*service)
	assert.True(t, ok)
	assert.NotNil(t, s.runner)
	assert.Equal(t, v, s.versionInfo)
	assert.NotNil(t, s.pathResolver)
	assert.Equal(t, mc, s.cacheService)
}

func TestUpdate_AlreadyLatest(t *testing.T) {
	mr := new(mockRunner)
	mrepo := new(mockRepositories)
	mp := new(mockPathResolver)
	v := model.VersionInfo{Version: tag}
	mc := new(services.MockCacheService)
	s := &service{runner: mr, repositories: mrepo, versionInfo: v, pathResolver: mp, cacheService: mc}

	mp.On("ResolvedExecutablePath").Return("/usr/local/bin/aws-doctor", nil)

	tagName := tag
	release := &github.RepositoryRelease{TagName: &tagName}

	mc.On("Get", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(false, nil)
	mc.On("Set", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(nil)
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(release, nil, nil)

	err := s.Update()
	assert.ErrorIs(t, err, model.ErrAlreadyLatest)
	mr.AssertNotCalled(t, "Run", mock.Anything, mock.Anything)
}

func TestUpdate_AlreadyLatestVPrefix(t *testing.T) {
	mr := new(mockRunner)
	mrepo := new(mockRepositories)
	mp := new(mockPathResolver)
	v := model.VersionInfo{Version: "1.2.3"}
	mc := new(services.MockCacheService)
	s := &service{runner: mr, repositories: mrepo, versionInfo: v, pathResolver: mp, cacheService: mc}

	mp.On("ResolvedExecutablePath").Return("/usr/local/bin/aws-doctor", nil)

	tagName := tag
	release := &github.RepositoryRelease{TagName: &tagName}

	mc.On("Get", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(false, nil)
	mc.On("Set", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(nil)
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(release, nil, nil)

	err := s.Update()
	assert.ErrorIs(t, err, model.ErrAlreadyLatest)
	mr.AssertNotCalled(t, "Run", mock.Anything, mock.Anything)
}

func TestUpdate_DevVersion(t *testing.T) {
	mr := new(mockRunner)
	mrepo := new(mockRepositories)
	mp := new(mockPathResolver)
	v := model.VersionInfo{Version: "dev"}
	mc := new(services.MockCacheService)
	s := &service{runner: mr, repositories: mrepo, versionInfo: v, pathResolver: mp, cacheService: mc}

	mp.On("ResolvedExecutablePath").Return("/usr/local/bin/aws-doctor", nil)

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
	mp := new(mockPathResolver)
	v := model.VersionInfo{Version: "v1.2.2"}
	mc := new(services.MockCacheService)
	s := &service{runner: mr, repositories: mrepo, versionInfo: v, pathResolver: mp, cacheService: mc}

	mp.On("ResolvedExecutablePath").Return("/usr/local/bin/aws-doctor", nil)

	tagName := tag
	release := &github.RepositoryRelease{TagName: &tagName}

	mc.On("Get", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(false, nil)
	mc.On("Set", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(nil)
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(release, nil, nil)

	mr.On("Run", "sh", []string{"-c", "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"}).Return(nil)

	err := s.Update()
	assert.NoError(t, err)
	mr.AssertExpectations(t)
}

func TestUpdate_InstallMethod(t *testing.T) {
	tests := []struct {
		name         string
		resolvedPath string
		version      string
		pathErr      error
		expectedErr  error
	}{
		{
			name:         "homebrew_apple_silicon",
			resolvedPath: "/opt/homebrew/Cellar/aws-doctor/1.0.0/bin/aws-doctor",
			expectedErr:  model.ErrHomebrewInstall,
		},
		{
			name:         "homebrew_intel_mac",
			resolvedPath: "/usr/local/Cellar/aws-doctor/1.0.0/bin/aws-doctor",
			expectedErr:  model.ErrHomebrewInstall,
		},
		{
			name:         "homebrew_linux",
			resolvedPath: "/home/linuxbrew/.linuxbrew/Cellar/aws-doctor/1.0.0/bin/aws-doctor",
			expectedErr:  model.ErrHomebrewInstall,
		},
		{
			name:         "go_install_default_gopath",
			resolvedPath: "/home/user/go/bin/aws-doctor",
			version:      "dev",
			expectedErr:  model.ErrGoInstall,
		},
		{
			name:         "go_install_custom_gopath",
			resolvedPath: "/opt/go/bin/aws-doctor",
			version:      "dev",
			expectedErr:  model.ErrGoInstall,
		},
		{
			name:         "go_install_custom_gopath_with_release_version",
			resolvedPath: "/opt/go/bin/aws-doctor",
			version:      "v2.9.0",
			expectedErr:  nil, // Should proceed with update
		},
		{
			name:         "non_homebrew_install",
			resolvedPath: "/usr/local/bin/aws-doctor",
		},
		{
			name:    "path_resolution_error_falls_through",
			pathErr: errors.New("cannot resolve path"),
		},
	}

	installCmd := []string{"-c", "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := new(mockRunner)
			mrepo := new(mockRepositories)
			mp := new(mockPathResolver)
			version := tt.version

			if version == "" {
				version = "dev"
			}

			v := model.VersionInfo{Version: version}
			mc := new(services.MockCacheService)
			s := &service{runner: mr, repositories: mrepo, versionInfo: v, pathResolver: mp, cacheService: mc}

			mp.On("ResolvedExecutablePath").Return(tt.resolvedPath, tt.pathErr)
			mc.On("Get", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(false, nil)

			if version != "dev" {
				tagName := "v3.0.0"
				release := &github.RepositoryRelease{TagName: &tagName}

				mc.On("Set", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(nil)
				mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(release, nil, nil)
			}

			if tt.expectedErr == nil {
				mr.On("Run", "sh", installCmd).Return(nil)
			}

			err := s.Update()

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				mr.AssertNotCalled(t, "Run", mock.Anything, mock.Anything)
				mrepo.AssertNotCalled(t, "GetLatestRelease", mock.Anything, mock.Anything, mock.Anything)
			} else {
				assert.NoError(t, err)
				mr.AssertCalled(t, "Run", "sh", installCmd)
			}

			mp.AssertExpectations(t)
			mr.AssertExpectations(t)
		})
	}
}

func TestUpdate_RateLimitError(t *testing.T) {
	mr := new(mockRunner)
	mrepo := new(mockRepositories)
	mp := new(mockPathResolver)
	v := model.VersionInfo{Version: tag}
	mc := new(services.MockCacheService)
	s := &service{runner: mr, repositories: mrepo, versionInfo: v, pathResolver: mp, cacheService: mc}

	mp.On("ResolvedExecutablePath").Return("/usr/local/bin/aws-doctor", nil)
	mc.On("Get", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(false, nil)
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
	mp := new(mockPathResolver)
	v := model.VersionInfo{Version: tag}
	mc := new(services.MockCacheService)
	s := &service{runner: mr, repositories: mrepo, versionInfo: v, pathResolver: mp, cacheService: mc}

	mp.On("ResolvedExecutablePath").Return("/usr/local/bin/aws-doctor", nil)
	mc.On("Get", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(false, nil)
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(nil, nil, errors.New("github error"))

	err := s.Update()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch latest release")
}

func TestUpdate_ExecutionError(t *testing.T) {
	mr := new(mockRunner)
	mrepo := new(mockRepositories)
	mp := new(mockPathResolver)
	v := model.VersionInfo{Version: "v1.2.2"}
	mc := new(services.MockCacheService)
	s := &service{runner: mr, repositories: mrepo, versionInfo: v, pathResolver: mp, cacheService: mc}

	mp.On("ResolvedExecutablePath").Return("/usr/local/bin/aws-doctor", nil)

	tagName := tag
	release := &github.RepositoryRelease{TagName: &tagName}

	mc.On("Get", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(false, nil)
	mc.On("Set", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(nil)
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(release, nil, nil)

	mr.On("Run", "sh", []string{"-c", "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"}).Return(errors.New("execution failed"))

	err := s.Update()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to run update script")
}

func TestCheckForUpdate_DevVersion(t *testing.T) {
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: "dev"}
	mc := new(services.MockCacheService)
	s := &service{repositories: mrepo, versionInfo: v, cacheService: mc}

	result, err := s.CheckForUpdate(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, result)
	mrepo.AssertNotCalled(t, "GetLatestRelease", mock.Anything, mock.Anything, mock.Anything)
}

func TestCheckForUpdate_AlreadyLatest(t *testing.T) {
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: tag}
	mc := new(services.MockCacheService)
	s := &service{repositories: mrepo, versionInfo: v, cacheService: mc}

	tagName := tag
	release := &github.RepositoryRelease{TagName: &tagName}

	mc.On("Get", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(false, nil)
	mc.On("Set", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(nil)
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(release, nil, nil)

	result, err := s.CheckForUpdate(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestCheckForUpdate_NewVersionAvailable(t *testing.T) {
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: "v1.2.2"}
	mc := new(services.MockCacheService)
	s := &service{repositories: mrepo, versionInfo: v, cacheService: mc}

	tagName := tag
	release := &github.RepositoryRelease{TagName: &tagName}

	mc.On("Get", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(false, nil)
	mc.On("Set", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(nil)
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(release, nil, nil)

	result, err := s.CheckForUpdate(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, tag, *result)
}

func TestCheckForUpdate_NilRelease(t *testing.T) {
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: tag}
	mc := new(services.MockCacheService)
	s := &service{repositories: mrepo, versionInfo: v, cacheService: mc}

	mc.On("Get", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(false, nil)
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(nil, nil, nil)

	result, err := s.CheckForUpdate(context.Background())
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "latest release is nil")
}

func TestCheckForUpdate_FromCache(t *testing.T) {
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: "v1.2.2"}
	mc := new(services.MockCacheService)
	s := &service{repositories: mrepo, versionInfo: v, cacheService: mc}

	mc.On("Get", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(true, nil).Run(func(args mock.Arguments) {
		target := args.Get(1).(*string)
		*target = tag
	})

	result, err := s.CheckForUpdate(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, tag, *result)

	// Verify GitHub was NOT called
	mrepo.AssertNotCalled(t, "GetLatestRelease", mock.Anything, mock.Anything, mock.Anything)
}

func TestCheckForUpdate_GitHubError(t *testing.T) {
	mrepo := new(mockRepositories)
	v := model.VersionInfo{Version: tag}
	mc := new(services.MockCacheService)
	s := &service{repositories: mrepo, versionInfo: v, cacheService: mc}

	mc.On("Get", cache.LatestVersionKey, mock.Anything, mock.Anything).Return(false, nil)
	mrepo.On("GetLatestRelease", mock.Anything, model.GitHubOwner, model.GitHubRepo).Return(nil, nil, errors.New("github error"))

	result, err := s.CheckForUpdate(context.Background())
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to fetch latest release")
}

func TestRealRunner_Run(t *testing.T) {
	// This actually tries to run a command.
	// We'll run something harmless like 'go version'.
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
