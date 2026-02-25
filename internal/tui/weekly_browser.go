package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/b00y0h/wakadash/internal/theme"
)

// WeeklyBrowserModel is a BubbleTea sub-model for browsing up to 52 weeks of history.
// It follows the same delegation pattern as ThemePickerModel: the parent model (Model)
// forwards messages to it and checks IsConfirmed/IsCancelled to react to user choices.
type WeeklyBrowserModel struct {
	weeks        []WeeklySummary // Available weeks with data (index 0 = most recent)
	selectedIdx  int             // Currently highlighted week
	scrollOffset int             // First visible row index (for scrolling)
	width        int
	height       int
	loading      bool   // True while fetching weekly summaries
	confirmed    bool   // True when user pressed Enter
	cancelled    bool   // True when user pressed Esc/q
	selectedWeek string // WeekStart date when confirmed (YYYY-MM-DD)
	err          error  // Fetch error
	thm          theme.Theme
}

// NewWeeklyBrowser creates a WeeklyBrowserModel in its initial loading state.
func NewWeeklyBrowser(t theme.Theme) WeeklyBrowserModel {
	return WeeklyBrowserModel{
		loading: true,
		thm:     t,
	}
}

// Init satisfies the tea.Model interface (no initial command needed; parent fires the cmd).
func (m WeeklyBrowserModel) Init() tea.Cmd {
	return nil
}

// Update handles messages forwarded from the parent Model.
func (m WeeklyBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case weeklyDataFetchedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.weeks = msg.weeks
		m.selectedIdx = 0
		m.scrollOffset = 0
		return m, nil

	case tea.KeyMsg:
		if m.loading || m.err != nil {
			// Allow Esc to dismiss even in error/loading state
			if msg.String() == "esc" || msg.String() == "q" {
				m.cancelled = true
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if len(m.weeks) == 0 {
				break
			}
			m.selectedIdx = (m.selectedIdx - 1 + len(m.weeks)) % len(m.weeks)
			m.clampScroll()

		case "down", "j":
			if len(m.weeks) == 0 {
				break
			}
			m.selectedIdx = (m.selectedIdx + 1) % len(m.weeks)
			m.clampScroll()

		case "home", "g":
			m.selectedIdx = 0
			m.scrollOffset = 0

		case "end", "G":
			if len(m.weeks) > 0 {
				m.selectedIdx = len(m.weeks) - 1
				m.clampScroll()
			}

		case "enter":
			if len(m.weeks) > 0 {
				m.confirmed = true
				m.selectedWeek = m.weeks[m.selectedIdx].WeekStart
			}

		case "esc", "q":
			m.cancelled = true
		}
	}

	return m, nil
}

// clampScroll adjusts scrollOffset so that selectedIdx is always visible.
func (m *WeeklyBrowserModel) clampScroll() {
	maxVisible := m.maxVisibleRows()
	if maxVisible <= 0 {
		return
	}
	// Scroll down
	if m.selectedIdx >= m.scrollOffset+maxVisible {
		m.scrollOffset = m.selectedIdx - maxVisible + 1
	}
	// Scroll up
	if m.selectedIdx < m.scrollOffset {
		m.scrollOffset = m.selectedIdx
	}
}

// maxVisibleRows calculates how many week rows fit in the current terminal height.
// Reserves 6 lines for header, navigation hint, blank lines, and scroll hint.
func (m WeeklyBrowserModel) maxVisibleRows() int {
	reserved := 6
	available := m.height - reserved
	if available < 1 {
		available = 1
	}
	return available
}

// View renders the weekly browser as a full-screen overlay.
func (m WeeklyBrowserModel) View() string {
	t := m.thm

	if m.loading {
		loadingStyle := lipgloss.NewStyle().Foreground(t.Dim)
		return "\n  " + loadingStyle.Render("Scanning weekly history...") + "\n"
	}

	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(t.Error)
		dimStyle := lipgloss.NewStyle().Foreground(t.Dim)
		return "\n  " + errStyle.Render(fmt.Sprintf("Error loading weekly data: %v", m.err)) +
			"\n\n  " + dimStyle.Render("Press Esc to return") + "\n"
	}

	var sb strings.Builder

	// Header
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(t.Title)
	countStr := fmt.Sprintf("Weekly History (%d weeks found)", len(m.weeks))
	sb.WriteString(titleStyle.Render(countStr) + "\n")

	navHint := lipgloss.NewStyle().Foreground(t.Dim).Render(
		"Arrow keys to browse, Enter to select, Esc to cancel")
	sb.WriteString(navHint + "\n\n")

	if len(m.weeks) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(t.Dim).Render("No historical data found.") + "\n")
		return sb.String()
	}

	maxVisible := m.maxVisibleRows()
	end := m.scrollOffset + maxVisible
	if end > len(m.weeks) {
		end = len(m.weeks)
	}

	// Render visible rows
	for i := m.scrollOffset; i < end; i++ {
		week := m.weeks[i]
		selected := i == m.selectedIdx

		cursor := "  "
		if selected {
			cursor = "> "
		}

		dateRange := formatWeekRangeFromStrings(week.WeekStart, week.WeekEnd)
		timeStr := formatSeconds(week.TotalSeconds)

		// Current week (index 0, always HasData but may have zero TotalSeconds from live source)
		if i == 0 {
			timeStr = "current week"
		}

		topLang := week.TopLanguage
		if topLang == "" {
			topLang = "-"
		}
		projStr := fmt.Sprintf("%d projects", week.ProjectCount)
		if i == 0 {
			projStr = ""
		}

		rowText := fmt.Sprintf("%-3s%-20s  %-12s  %-14s  %s",
			cursor, dateRange, timeStr, topLang, projStr)

		var rowStyle lipgloss.Style
		if selected {
			rowStyle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
		} else {
			rowStyle = lipgloss.NewStyle().Foreground(t.Foreground)
		}

		sb.WriteString(rowStyle.Render(rowText) + "\n")
	}

	// Scroll hint when list is longer than viewport
	if len(m.weeks) > maxVisible {
		dimStyle := lipgloss.NewStyle().Foreground(t.Dim)
		showing := fmt.Sprintf("  Showing %d-%d of %d", m.scrollOffset+1, end, len(m.weeks))
		sb.WriteString(dimStyle.Render(showing) + "\n")
	}

	return sb.String()
}

// formatWeekRangeFromStrings formats a week range like "Feb 23 - Mar 1" from two date strings.
func formatWeekRangeFromStrings(startStr, endStr string) string {
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return startStr
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return startStr
	}
	if start.Month() == end.Month() {
		return fmt.Sprintf("%s %d-%d", start.Format("Jan"), start.Day(), end.Day())
	}
	return fmt.Sprintf("%s %d - %s %d", start.Format("Jan"), start.Day(), end.Format("Jan"), end.Day())
}

// IsConfirmed returns true if the user pressed Enter to select a week.
func (m WeeklyBrowserModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if the user pressed Esc/q to dismiss the browser.
func (m WeeklyBrowserModel) IsCancelled() bool {
	return m.cancelled
}

// SelectedWeek returns the selected week's start date (YYYY-MM-DD Sunday).
// Only valid after IsConfirmed() returns true.
func (m WeeklyBrowserModel) SelectedWeek() string {
	return m.selectedWeek
}
