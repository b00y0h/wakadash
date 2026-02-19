package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/b00y0h/wakadash/internal/api"
	"github.com/b00y0h/wakadash/internal/types"
)

// Model is the bubbletea tea.Model for the wakadash dashboard.
// It follows The Elm Architecture: all state is in Model, mutations
// only happen in Update, and View is a pure render function.
type Model struct {
	// Layout — initialized to 80x24 (safe defaults before WindowSizeMsg arrives,
	// per research pitfall #1: View is called before the first WindowSizeMsg).
	width  int
	height int

	// Data
	stats     *types.StatsResponse
	loading   bool
	err       error
	lastFetch time.Time

	// Refresh timer
	refreshInterval time.Duration
	nextRefresh     time.Time

	// Dependencies
	client   *api.Client
	rangeStr string

	// UI components
	spinner spinner.Model
	help    help.Model
	keys    keymap

	// State
	quitting bool
	showHelp bool
}

// NewModel creates a new Model with the given API client, time range, and refresh interval.
// rangeStr defaults to "last_7_days" if empty.
// Valid values: last_7_days, last_30_days, last_6_months, last_year, all_time.
// refreshInterval defaults to 60s if zero.
func NewModel(client *api.Client, rangeStr string, refreshInterval time.Duration) Model {
	if rangeStr == "" {
		rangeStr = "last_7_days"
	}
	if refreshInterval == 0 {
		refreshInterval = 60 * time.Second
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	h := help.New()
	h.Width = 80

	return Model{
		// Safe defaults — overridden by WindowSizeMsg before meaningful renders.
		width:           80,
		height:          24,
		loading:         true,
		client:          client,
		rangeStr:        rangeStr,
		refreshInterval: refreshInterval,
		spinner:         s,
		help:            h,
		keys:            defaultKeymap,
	}
}

// Init starts the initial async stats fetch, spinner animation, and countdown ticker.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchStatsCmd(m.client, m.rangeStr),
		m.spinner.Tick,
		tickEverySecond(),
	)
}

// Update handles incoming messages and returns an updated model and next command.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		case key.Matches(msg, m.keys.Refresh):
			m.loading = true
			return m, tea.Batch(fetchStatsCmd(m.client, m.rangeStr), m.spinner.Tick)
		}
		return m, nil

	case statsFetchedMsg:
		m.loading = false
		m.stats = msg.stats
		m.err = nil
		m.lastFetch = time.Now()
		m.nextRefresh = time.Now().Add(m.refreshInterval)
		return m, scheduleRefresh(m.refreshInterval)

	case fetchErrMsg:
		m.loading = false
		m.err = msg.err
		m.nextRefresh = time.Now().Add(m.refreshInterval)
		return m, scheduleRefresh(m.refreshInterval)

	case refreshMsg:
		// Time to refresh - kick off new fetch
		m.loading = true
		return m, tea.Batch(fetchStatsCmd(m.client, m.rangeStr), m.spinner.Tick)

	case countdownTickMsg:
		// Continue countdown ticker (self-loop)
		return m, tickEverySecond()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the current state of the dashboard as a string.
// This is a pure function — it reads model state and produces terminal output.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.showHelp {
		return m.renderHelp()
	}

	if m.loading {
		return fmt.Sprintf("\n  %s Fetching stats...\n", m.spinner.View())
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.", m.err))
	}

	return m.renderDashboard()
}

// renderDashboard renders the full stats dashboard.
func (m Model) renderDashboard() string {
	content := m.renderStats()
	statusBar := m.renderStatusBar()

	// Account for border (-2) and status bar height.
	panelHeight := m.height - lipgloss.Height(statusBar) - 4
	if panelHeight < 1 {
		panelHeight = 1
	}

	statsPanel := borderStyle.
		Width(m.width - 2).
		Height(panelHeight).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, statsPanel, statusBar)
}

// renderStats renders the stats content shown inside the border panel.
func (m Model) renderStats() string {
	if m.stats == nil {
		return ""
	}

	data := m.stats.Data
	var sb strings.Builder

	// Title
	title := titleStyle.Render(fmt.Sprintf("WakaTime Stats (%s)", m.rangeStr))
	sb.WriteString(title + "\n\n")

	// Totals
	sb.WriteString(fmt.Sprintf("  Total time:    %s\n", data.HumanReadableTotal))
	sb.WriteString(fmt.Sprintf("  Daily average: %s\n", data.HumanReadableDailyAverage))

	// Top 5 languages
	if len(data.Languages) > 0 {
		sb.WriteString("\n")
		sb.WriteString(titleStyle.Render("  Languages") + "\n")
		limit := 5
		if len(data.Languages) < limit {
			limit = len(data.Languages)
		}
		for _, lang := range data.Languages[:limit] {
			sb.WriteString(fmt.Sprintf("    %-20s %s\n", lang.Name, formatSeconds(lang.TotalSeconds)))
		}
	}

	// Top 5 projects
	if len(data.Projects) > 0 {
		sb.WriteString("\n")
		sb.WriteString(titleStyle.Render("  Projects") + "\n")
		limit := 5
		if len(data.Projects) < limit {
			limit = len(data.Projects)
		}
		for _, proj := range data.Projects[:limit] {
			sb.WriteString(fmt.Sprintf("    %-20s %s\n", proj.Name, formatSeconds(proj.TotalSeconds)))
		}
	}

	return sb.String()
}

// renderStatusBar renders the bottom status line.
func (m Model) renderStatusBar() string {
	var status string
	if m.loading {
		status = m.spinner.View() + " Fetching..."
	} else if m.err != nil {
		status = errorStyle.Render("Error: " + m.err.Error())
	} else {
		remaining := time.Until(m.nextRefresh).Round(time.Second)
		if remaining < 0 {
			remaining = 0
		}
		status = fmt.Sprintf("Updated: %s  Next: %s",
			m.lastFetch.Format("15:04:05"),
			remaining,
		)
	}

	helpHint := dimStyle.Render("? help  r refresh  q quit")
	gap := strings.Repeat(" ", max(0, m.width-lipgloss.Width(status)-lipgloss.Width(helpHint)))
	return dimStyle.Render(status) + gap + helpHint
}

// renderHelp renders the help overlay showing keyboard shortcuts.
func (m Model) renderHelp() string {
	title := titleStyle.Render("Keyboard Shortcuts")
	helpText := m.help.View(m.keys)
	hint := dimStyle.Render("\nPress ? to return to dashboard")
	return lipgloss.JoinVertical(lipgloss.Left, title, "", helpText, hint)
}

// formatSeconds converts a float64 seconds value to a human-readable string.
func formatSeconds(secs float64) string {
	total := int(secs)
	h := total / 3600
	m := (total % 3600) / 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
