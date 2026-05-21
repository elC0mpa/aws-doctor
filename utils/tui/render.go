package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
)

// RenderWasteInteractive launches the Bubble Tea UI for the waste checks.
func RenderWasteInteractive(accountID string, resultCh <-chan model.ScopeResult, scopes []string, pricingSvc pricing.Service) error {
	p := tea.NewProgram(NewWasteModel(accountID, resultCh, scopes, pricingSvc), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
	return nil
}
