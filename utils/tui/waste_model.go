package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	wastetable "github.com/elC0mpa/aws-doctor/utils/waste_table"
)

const statusEOF = "EOF"

const (
	statusError = "error"
	statusDone  = "done"
	scopeErrors = "Errors"
)

var (
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}
	inactiveTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tabStyle = lipgloss.NewStyle().
			Border(inactiveTabBorder, true).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
	activeTabStyle = lipgloss.NewStyle().
			Border(activeTabBorder, true).
			BorderForeground(lipgloss.Color("39")).
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Padding(0, 1)

	windowStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("39")).
			Padding(1, 0)
)

type wasteModel struct {
	scopes         []string
	activeTab      int
	scopeStatus    map[string]string // "loading", statusDone, statusError
	scopeDuration  map[string]time.Duration
	aggregatedData model.RenderWasteInput
	resultCh       <-chan model.ScopeResult
	spinner        spinner.Model
	viewport       viewport.Model
	pricingSvc     pricing.Service
	accountID      string
	done           bool
	ready          bool
}

type scopeMsg model.ScopeResult

func waitForResult(resultCh <-chan model.ScopeResult) tea.Cmd {
	return func() tea.Msg {
		res, ok := <-resultCh
		if !ok {
			return scopeMsg{Scope: statusEOF}
		}

		return scopeMsg(res)
	}
}

// NewWasteModel creates a new Bubble Tea model for the waste TUI
func NewWasteModel(accountID string, resultCh <-chan model.ScopeResult, scopes []string, pricingSvc pricing.Service) wasteModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return wasteModel{
		scopes:        scopes,
		activeTab:     0,
		scopeStatus:   make(map[string]string),
		scopeDuration: make(map[string]time.Duration),
		resultCh:      resultCh,
		spinner:       s,
		pricingSvc:    pricingSvc,
		accountID:     accountID,
		aggregatedData: model.RenderWasteInput{
			AccountID: accountID,
		},
	}
}

func (m wasteModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		waitForResult(m.resultCh),
	)
}

func (m wasteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 8 // Approximate header + tabs height
		footerHeight := 2 // Instructions height
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "right", "tab":
			m.activeTab = (m.activeTab + 1) % len(m.scopes)
			m.viewport.GotoTop() // Reset scroll on tab switch
		case "left", "shift+tab":
			m.activeTab = (m.activeTab - 1 + len(m.scopes)) % len(m.scopes)
			m.viewport.GotoTop() // Reset scroll on tab switch
		}

	case scopeMsg:
		if msg.Scope == statusEOF && !m.done {
			m.done = true

			if len(m.aggregatedData.Errors) > 0 {
				m.scopes = append(m.scopes, scopeErrors)
				m.scopeStatus[scopeErrors] = statusDone
			}

			m.scopes = append(m.scopes, "Summary")

			m.scopeStatus["Summary"] = statusDone

			m.syncViewportContent()

			return m, nil
		}

		m.scopeStatus[msg.Scope] = statusDone

		m.scopeDuration[msg.Scope] = msg.Duration
		if msg.Err != nil {
			m.scopeStatus[msg.Scope] = statusError
		}

		if m.accountID == "" && msg.Input.AccountID != "" {
			m.accountID = msg.Input.AccountID
			m.aggregatedData.AccountID = msg.Input.AccountID
		}

		m.aggregatedData.Merge(msg.Input)

		cmds = append(cmds, waitForResult(m.resultCh))

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.syncViewportContent()

	// Always update viewport for scrolling
	if m.ready {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *wasteModel) syncViewportContent() {
	if !m.ready {
		return
	}

	activeScope := m.scopes[m.activeTab]

	var contentStr string

	switch m.scopeStatus[activeScope] {
	case "":
		contentStr = windowStyle.Render("Scanning " + activeScope + "... " + m.spinner.View())
	case statusError:
		text := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("⚠️  Failed to scan " + activeScope + ". See 'Errors' tab for details.")
		contentStr = windowStyle.Render(text)
	default:
		var tableStr string
		if activeScope == scopeErrors {
			tableStr = wastetable.RenderErrorsTable(m.aggregatedData.Errors)
		} else {
			tableStr = wastetable.RenderScopeTable(activeScope, m.aggregatedData, m.pricingSvc)
		}

		durationStr := ""

		if activeScope != "Summary" && activeScope != scopeErrors {
			duration := m.scopeDuration[activeScope]
			durationStr = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Fetched in " + duration.Round(time.Millisecond).String())
		}

		if strings.TrimSpace(tableStr) == "" {
			joined := lipgloss.JoinVertical(lipgloss.Left, durationStr, "", "✅ No waste found for "+activeScope)
			contentStr = windowStyle.Render(joined)
		} else {
			joined := lipgloss.JoinVertical(lipgloss.Left, durationStr, "", tableStr)
			contentStr = windowStyle.Render(joined)
		}
	}

	m.viewport.SetContent(contentStr)
}

func (m wasteModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	doc := strings.Builder{}
	doc.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("🏥 AWS DOCTOR WASTE REPORT"))
	doc.WriteString("\n")

	if m.accountID != "" {
		doc.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render("Account: " + m.accountID))
	} else {
		doc.WriteString("Account: Loading...")
	}

	doc.WriteString("\n\n")

	// Tabs
	var tabs []string

	for i, scope := range m.scopes {
		name := scope

		if scope == scopeErrors {
			name += " ⚠️"
		} else {
			status := m.scopeStatus[scope]
			switch status {
			case "":
				name += " " + m.spinner.View()
			case statusError:
				name += " ❌"
			default:
				name += " ✅"
			}
		}

		if i == m.activeTab {
			tabs = append(tabs, activeTabStyle.Render(name))
		} else {
			tabs = append(tabs, tabStyle.Render(name))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	doc.WriteString(row)
	doc.WriteString("\n")

	// Content
	doc.WriteString(m.viewport.View())

	doc.WriteString("\n\nPress 'tab' or left/right arrows to navigate. Press up/down to scroll. Press 'q' to quit.")

	return doc.String()
}
