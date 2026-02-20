package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/b00y0h/wakadash/internal/theme"
)

// ThemePickerModel is a BubbleTea model for theme selection.
type ThemePickerModel struct {
	themes        []string      // Theme names from theme.AllThemes()
	selectedIdx   int           // Current selection
	width         int           // Terminal width
	height        int           // Terminal height
	confirmed     bool          // True when user pressed Enter
	cancelled     bool          // True when user pressed Esc/Q (runtime only)
	selectedTheme string        // Theme name when confirmed
	isFirstRun    bool          // True for first-run flow (no cancel allowed)
}

// NewThemePicker creates a new theme picker model.
// isFirstRun controls cancel behavior:
// - First-run (isFirstRun=true): No dashboard to return to, user MUST select a theme
// - Runtime (isFirstRun=false): Dashboard exists, Esc/Q cancels and returns without changing theme
func NewThemePicker(isFirstRun bool) ThemePickerModel {
	return ThemePickerModel{
		themes:      theme.AllThemes(),
		selectedIdx: 0, // Start at Dracula
		isFirstRun:  isFirstRun,
	}
}

// Init initializes the theme picker (no initial command needed).
func (m ThemePickerModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the picker state.
func (m ThemePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.selectedIdx = (m.selectedIdx - 1 + len(m.themes)) % len(m.themes)
			return m, nil
		case "down", "j":
			m.selectedIdx = (m.selectedIdx + 1) % len(m.themes)
			return m, nil
		case "enter":
			// Confirm selection and save to config
			m.confirmed = true
			m.selectedTheme = m.themes[m.selectedIdx]
			theme.SaveThemeToConfig(m.selectedTheme)
			return m, tea.Quit
		case "esc", "q":
			if m.isFirstRun {
				// First-run: ignore cancel (user MUST select a theme)
				return m, nil
			}
			// Runtime: cancel and return to dashboard
			m.cancelled = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the theme picker interface.
func (m ThemePickerModel) View() string {
	if m.width < 40 || m.height < 10 {
		return "Terminal too small. Please resize."
	}

	// Get current theme
	currentTheme := theme.GetTheme(m.themes[m.selectedIdx])

	var sb strings.Builder

	// Theme label
	themeLabel := lipgloss.NewStyle().
		Foreground(currentTheme.Title).
		Bold(true).
		Render(fmt.Sprintf("Theme: %s", currentTheme.Name))
	sb.WriteString(themeLabel + "\n\n")

	// Navigation hint
	navHint := lipgloss.NewStyle().
		Foreground(currentTheme.Dim).
		Render("Arrow keys to browse, Enter to select")
	if !m.isFirstRun {
		navHint = lipgloss.NewStyle().
			Foreground(currentTheme.Dim).
			Render("Arrow keys to browse, Enter to select, Esc/Q to cancel")
	}
	sb.WriteString(navHint + "\n\n")

	// Mini dashboard preview
	preview := m.renderMiniDashboard(currentTheme)
	sb.WriteString(preview)

	return sb.String()
}

// renderMiniDashboard renders a scaled-down mock dashboard with sample data.
func (m ThemePickerModel) renderMiniDashboard(t theme.Theme) string {
	var sb strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Foreground(t.Title).
		Bold(true).
		Render("WakaTime Stats (last_7_days)")
	sb.WriteString(title + "\n\n")

	// Sample stats
	sb.WriteString(lipgloss.NewStyle().
		Foreground(t.Foreground).
		Render("  Total time:    42h 15m") + "\n")
	sb.WriteString(lipgloss.NewStyle().
		Foreground(t.Foreground).
		Render("  Daily average: 6h 2m") + "\n\n")

	// Languages section with sample bars
	langTitle := lipgloss.NewStyle().
		Foreground(t.Title).
		Render("Languages")
	sb.WriteString(langTitle + "\n")

	// Three sample language bars
	languages := []struct {
		name  string
		hours float64
		color lipgloss.Color
	}{
		{"Go", 18.5, t.Accent1},
		{"TypeScript", 14.2, t.Accent2},
		{"Python", 9.5, t.Accent3},
	}

	maxHours := 18.5
	barWidth := 20

	for _, lang := range languages {
		// Calculate bar length (proportional to hours)
		length := int((lang.hours / maxHours) * float64(barWidth))
		if length < 1 {
			length = 1
		}
		bar := strings.Repeat("█", length)

		// Render language bar
		langLine := fmt.Sprintf("  %-12s ", lang.name)
		coloredBar := lipgloss.NewStyle().
			Foreground(lang.color).
			Render(bar)
		hours := lipgloss.NewStyle().
			Foreground(t.Dim).
			Render(fmt.Sprintf(" %.1fh", lang.hours))

		sb.WriteString(langLine + coloredBar + hours + "\n")
	}

	sb.WriteString("\n")

	// Activity heatmap row
	heatmapTitle := lipgloss.NewStyle().
		Foreground(t.Title).
		Render("Activity (Last 7 Days)")
	sb.WriteString(heatmapTitle + "\n")

	// 7 blocks with varying intensity (using heatmap gradient)
	intensities := []int{1, 2, 3, 4, 4, 3, 2} // Low to VeryHigh pattern
	var blocks []string
	for i, intensity := range intensities {
		day := fmt.Sprintf("D%d", i+1)
		block := lipgloss.NewStyle().
			Background(t.HeatmapColors[intensity]).
			Foreground(t.Foreground).
			Padding(0, 1).
			Render(day)
		blocks = append(blocks, block)
	}
	heatmap := lipgloss.JoinHorizontal(lipgloss.Top, blocks...)
	sb.WriteString(heatmap + "\n\n")

	// Wrap in border using theme border color
	content := sb.String()
	bordered := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(1, 2).
		Render(content)

	return bordered
}

// IsConfirmed returns true if the user confirmed a theme selection.
func (m ThemePickerModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if the user cancelled the picker (runtime mode only).
func (m ThemePickerModel) IsCancelled() bool {
	return m.cancelled
}

// SelectedTheme returns the selected theme name (only valid after confirmation).
func (m ThemePickerModel) SelectedTheme() string {
	return m.selectedTheme
}
