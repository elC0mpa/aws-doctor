package update

import (
	"errors"
	"testing"

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

func TestNewService(t *testing.T) {
	svc := NewService()
	assert.NotNil(t, svc)

	s, ok := svc.(*service)
	assert.True(t, ok)
	assert.NotNil(t, s.runner)
}

func TestUpdate_Success(t *testing.T) {
	mr := new(mockRunner)
	s := &service{runner: mr}

	mr.On("Run", "sh", []string{"-c", "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"}).Return(nil)

	err := s.Update()
	assert.NoError(t, err)
	mr.AssertExpectations(t)
}

func TestUpdate_Error(t *testing.T) {
	mr := new(mockRunner)
	s := &service{runner: mr}

	mr.On("Run", "sh", []string{"-c", "curl -sSL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh"}).Return(errors.New("execution failed"))

	err := s.Update()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to run update script")
	assert.Contains(t, err.Error(), "execution failed")
	mr.AssertExpectations(t)
}

func TestRealRunner_Run(t *testing.T) {
	// This actually tries to run a command.
	// We'll run something harmless like 'true'.
	r := &realRunner{}
	err := r.Run("true")
	assert.NoError(t, err)
}

func TestRealRunner_Run_Error(t *testing.T) {
	r := &realRunner{}
	// Running a non-existent command should return an error
	err := r.Run("non-existent-command-12345")
	assert.Error(t, err)
}
