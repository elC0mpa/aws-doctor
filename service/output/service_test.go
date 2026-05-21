package output

import (
	"testing"
	"github.com/elC0mpa/aws-doctor/model"
)

func TestIsInteractive(t *testing.T) {
	s := NewService("table")
	if s.IsInteractive() {
		t.Error("Expected IsInteractive to be false (non-TTY default in tests)")
	}
}

func TestRenderWasteInteractive_Smoke(t *testing.T) {
	s := NewService("table")
	resultCh := make(chan model.ScopeResult)
	go func() {
		close(resultCh)
	}()
	err := s.RenderWasteInteractive("123456789012", resultCh, []string{"EC2"}, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
