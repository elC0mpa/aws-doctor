package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
)

func TestWasteModel_Init(t *testing.T) {
	resultCh := make(chan model.ScopeResult)
	mockPricing := pricing.NewService(aws.Config{})
	m := NewWasteModel("123456789012", resultCh, []string{"EC2", "VPC"}, mockPricing)

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a batch command")
	}
}

func TestWasteModel_Update(t *testing.T) {
	resultCh := make(chan model.ScopeResult)
	mockPricing := pricing.NewService(aws.Config{})
	wm := NewWasteModel("123456789012", resultCh, []string{"EC2", "VPC"}, mockPricing)

	// WindowSizeMsg
	var (
		cmd      tea.Cmd
		newModel tea.Model
	)

	newModel, _ = wm.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	wm = newModel.(wasteModel)
	if !wm.ready {
		t.Error("Model should be ready after WindowSizeMsg")
	}

	// KeyMsg tab
	newModel, _ = wm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("tab")})

	wm = newModel.(wasteModel)
	if wm.activeTab != 1 {
		t.Errorf("Expected activeTab 1, got %d", wm.activeTab)
	}

	// KeyMsg shift+tab
	newModel, _ = wm.Update(tea.KeyMsg{Type: tea.KeyShiftTab})

	wm = newModel.(wasteModel)
	if wm.activeTab != 0 {
		t.Errorf("Expected activeTab 0, got %d", wm.activeTab)
	}

	// KeyMsg quit
	_, cmd = wm.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd() != tea.Quit() {
		t.Error("Expected Quit command on ctrl+c")
	}

	// scopeMsg
	newModel, _ = wm.Update(scopeMsg{Scope: "EC2", Duration: time.Second})

	wm = newModel.(wasteModel)
	if wm.scopeStatus["EC2"] != statusDone {
		t.Error("Expected EC2 scope status to be done")
	}

	// scopeMsg EOF
	newModel, _ = wm.Update(scopeMsg{Scope: "EOF"})

	wm = newModel.(wasteModel)
	if !wm.done {
		t.Error("Expected done flag to be true on EOF")
	}
}

func TestWasteModel_View(t *testing.T) {
	resultCh := make(chan model.ScopeResult)
	mockPricing := pricing.NewService(aws.Config{})
	wm := NewWasteModel("123456789012", resultCh, []string{"EC2"}, mockPricing)

	viewStr := wm.View()
	if viewStr == "" {
		t.Error("View should return initializing string")
	}

	// Make ready
	newModel, _ := wm.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	wm = newModel.(wasteModel)

	wm.scopeStatus["EC2"] = statusDone

	viewStr = wm.View()
	if viewStr == "" {
		t.Error("View should return content string")
	}
}

func TestWasteModel_Update_EOFAndSummary(t *testing.T) {
	resultCh := make(chan model.ScopeResult)
	mockPricing := pricing.NewService(aws.Config{})
	wm := NewWasteModel("123456789012", resultCh, []string{"EC2"}, mockPricing)
	wm.ready = true

	// Simulate EOF
	updatedModel, _ := wm.Update(scopeMsg{Scope: "EOF"})

	wm = updatedModel.(wasteModel)
	if !wm.done {
		t.Error("Expected done to be true")
	}

	// Switch to Summary tab
	updatedModel, _ = wm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("tab")})

	wm = updatedModel.(wasteModel)
	if wm.activeTab != 1 {
		t.Errorf("Expected activeTab 1, got %d", wm.activeTab)
	}
}

func TestWasteModel_ScrollUpAndDown(t *testing.T) {
	resultCh := make(chan model.ScopeResult)
	mockPricing := pricing.NewService(aws.Config{})
	wm := NewWasteModel("123456789012", resultCh, []string{"EC2"}, mockPricing)
	wm.ready = true

	// Simulate ArrowUp and ArrowDown
	_, _ = wm.Update(tea.KeyMsg{Type: tea.KeyUp})
	_, _ = wm.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, _ = wm.Update(tea.KeyMsg{Type: tea.KeyLeft})
	_, _ = wm.Update(tea.KeyMsg{Type: tea.KeyRight})
}

func TestWasteModel_HandleResize(t *testing.T) {
	resultCh := make(chan model.ScopeResult)
	mockPricing := pricing.NewService(aws.Config{})
	wm := NewWasteModel("123456789012", resultCh, []string{"EC2"}, mockPricing)

	// Large resize
	wm2, _ := wm.Update(tea.WindowSizeMsg{Width: 200, Height: 100})

	// Small resize
	_, _ = wm2.Update(tea.WindowSizeMsg{Width: 20, Height: 10})
}

func TestWasteModel_RenderScopes_ThroughView(t *testing.T) {
	resultCh := make(chan model.ScopeResult)
	mockPricing := pricing.NewService(aws.Config{})
	scopes := []string{"EC2", "VPC", "ELB", "S3", "CloudWatch", "RDS", "Lambda", "SageMaker", "ECR", "SecretsManager", "Summary"}
	wm := NewWasteModel("123456789012", resultCh, scopes, mockPricing)
	wm.ready = true

	for i := range scopes {
		wm.activeTab = i
		wm.syncViewportContent()

		view := wm.View()
		if view == "" {
			t.Errorf("Empty view for tab %d (%s)", i, scopes[i])
		}
	}
}

func TestWasteModel_Update_ScopeError(t *testing.T) {
	resultCh := make(chan model.ScopeResult)
	mockPricing := pricing.NewService(aws.Config{})
	wm := NewWasteModel("123456789012", resultCh, []string{"EC2"}, mockPricing)

	updatedModel, _ := wm.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	wm = updatedModel.(wasteModel)

	updatedModel, _ = wm.Update(scopeMsg{Scope: "EC2", Err: errors.New("test error")})
	wm = updatedModel.(wasteModel)

	if wm.scopeStatus["EC2"] != "error" {
		t.Error("Expected EC2 status to be statusError")
	}

	wm.syncViewportContent()
	view := wm.View()
	if !strings.Contains(view, "⚠️ ") {
		t.Error("Expected error warning in viewport content")
	}
}

func TestWasteModel_Update_EOFWithErrors(t *testing.T) {
	resultCh := make(chan model.ScopeResult)
	mockPricing := pricing.NewService(aws.Config{})
	wm := NewWasteModel("123456789012", resultCh, []string{"EC2"}, mockPricing)

	updatedModel, _ := wm.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	wm = updatedModel.(wasteModel)

	wm.aggregatedData.Errors = map[string]string{
		"EC2": "Access Denied",
	}

	updatedModel, _ = wm.Update(scopeMsg{Scope: "EOF"})
	wm = updatedModel.(wasteModel)

	if !wm.done {
		t.Error("Expected done to be true")
	}

	hasErrorsTab := false
	for _, scope := range wm.scopes {
		if scope == "Errors" {
			hasErrorsTab = true
			break
		}
	}

	if !hasErrorsTab {
		t.Error("Expected 'Errors' tab to be appended")
	}

	// Switch to Errors tab
	wm.activeTab = len(wm.scopes) - 2 // "Errors" is inserted before "Summary"
	wm.syncViewportContent()

	view := wm.View()
	if !strings.Contains(view, "ERRORS ENCOUNTERED DURING SCAN") {
		t.Error("Expected errors section to be rendered in the view")
	}
}
