package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
)

// RenderWasteInteractive launches the Bubble Tea UI for the waste checks.
func RenderWasteInteractive(accountID string, resultCh <-chan model.ScopeResult, scopes []string, pricingSvc pricing.Service) error {
	p := tea.NewProgram(NewWasteModel(accountID, resultCh, scopes, pricingSvc), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run interactive waste rendering: %w", err)
	}

	return nil
}
