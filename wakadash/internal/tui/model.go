package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/b00y0h/wakadash/internal/api"
	"github.com/b00y0h/wakadash/internal/archive"
	"github.com/b00y0h/wakadash/internal/theme"
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

	// Archive fetcher (nil when history_repo not configured)
	archiveFetcher *archive.Fetcher

	// Archived data for historical dates
	archiveData *types.DayData

	// Theme
	theme theme.Theme // Active color theme

	// UI components
	spinner         spinner.Model
	help            help.Model
	keys            keymap
	sparklineChart  sparkline.Model
	languagesChart  barchart.Model
	projectsChart   barchart.Model
	categoriesChart barchart.Model
	editorsChart    barchart.Model
	osChart         barchart.Model
	machinesChart   barchart.Model

	// Sparkline data
	hourlyData []float64 // 24 hours of activity

	// Heatmap data
	summaryData *types.SummaryResponse // For heatmap

	// Panel visibility (toggled by number keys)
	showSummary    bool // Summary panel visibility
	showLanguages  bool // 1 key
	showProjects   bool // 2 key
	showSparkline  bool // 3 key
	showHeatmap    bool // 4 key
	showCategories bool // 5 key (will be mapped in plan 09-03)
	showEditors    bool // 6 key
	showOS         bool // 7 key
	showMachines   bool // 8 key

	// Theme picker
	showPicker bool             // True when showing theme picker
	picker     ThemePickerModel // Theme picker model

	// State
	quitting    bool
	showHelp    bool
	rateLimited bool // Visual indicator for rate limit status
}

// NewModel creates a new Model with the given API client, time range, and refresh interval.
// rangeStr defaults to "last_7_days" if empty.
// Valid values: last_7_days, last_30_days, last_6_months, last_year, all_time.
// refreshInterval defaults to 60s if zero.
// archiveFetcher may be nil if history_repo is not configured.
func NewModel(client *api.Client, rangeStr string, refreshInterval time.Duration, archiveFetcher *archive.Fetcher) Model {
	if rangeStr == "" {
		rangeStr = "last_7_days"
	}
	if refreshInterval == 0 {
		refreshInterval = 60 * time.Second
	}

	// Load theme from config
	themeName, _ := theme.LoadThemeFromConfig()
	if themeName == "" {
		themeName = theme.DefaultTheme
	}
	activeTheme := theme.GetTheme(themeName)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(activeTheme.Primary)

	h := help.New()
	h.Width = 80

	sparklineChart := sparkline.New(70, 5)
	languagesChart := barchart.New(35, 8)
	projectsChart := barchart.New(35, 8)
	categoriesChart := barchart.New(35, 10)
	editorsChart := barchart.New(35, 10)
	osChart := barchart.New(35, 10)
	machinesChart := barchart.New(35, 10)

	return Model{
		// Safe defaults — overridden by WindowSizeMsg before meaningful renders.
		width:           80,
		height:          24,
		loading:         true,
		client:          client,
		rangeStr:        rangeStr,
		refreshInterval: refreshInterval,
		archiveFetcher:  archiveFetcher,
		theme:           activeTheme,
		spinner:         s,
		help:            h,
		keys:            defaultKeymap,
		sparklineChart:  sparklineChart,
		languagesChart:  languagesChart,
		projectsChart:   projectsChart,
		categoriesChart: categoriesChart,
		editorsChart:    editorsChart,
		osChart:         osChart,
		machinesChart:   machinesChart,
		showSummary:     true,
		showLanguages:   true,
		showProjects:    true,
		showSparkline:   true,
		showHeatmap:     true,
		showCategories:  true,
		showEditors:     true,
		showOS:          true,
		showMachines:    true,
	}
}

// Init starts the initial async stats fetch, spinner animation, and countdown ticker.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchStatsCmd(m.client, m.rangeStr),
		fetchDurationsCmd(m.client),
		fetchSummaryCmd(m.client),
		m.spinner.Tick,
		tickEverySecond(),
	)
}

// Update handles incoming messages and returns an updated model and next command.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Delegate to picker when in picker mode
	if m.showPicker {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.picker.width = msg.Width
			m.picker.height = msg.Height
			m.width = msg.Width
			m.height = msg.Height
		case tea.KeyMsg:
			var cmd tea.Cmd
			newPicker, _ := m.picker.Update(msg)
			m.picker = newPicker.(ThemePickerModel)
			if m.picker.IsConfirmed() {
				// Picker done - apply theme (already saved to config by picker)
				themeName := m.picker.SelectedTheme()
				m.theme = theme.GetTheme(themeName)
				m.showPicker = false
				// Update spinner style with new theme
				m.spinner.Style = lipgloss.NewStyle().Foreground(m.theme.Primary)
				return m, cmd
			}
			if m.picker.IsCancelled() {
				// User cancelled - return to dashboard with existing theme unchanged
				m.showPicker = false
				return m, cmd
			}
			return m, cmd
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		// Calculate layout dimensions
		panelWidth := (msg.Width / 2) - 4 // For side-by-side panels
		fullWidth := msg.Width - 4        // For full-width panels
		chartHeight := 8                  // Bar chart height
		sparklineHeight := 5              // Sparkline height

		// Resize all chart models
		m.languagesChart.Resize(panelWidth, chartHeight)
		m.projectsChart.Resize(panelWidth, chartHeight)
		m.categoriesChart.Resize(panelWidth, 10)
		m.editorsChart.Resize(panelWidth, 10)
		m.osChart.Resize(panelWidth, 10)
		m.machinesChart.Resize(panelWidth, 10)
		m.sparklineChart.Resize(fullWidth, sparklineHeight)

		// Redraw charts with new dimensions
		if m.stats != nil {
			m.updateLanguagesChart()
			m.updateProjectsChart()
			m.updateCategoriesChart()
			m.updateEditorsChart()
			m.updateOSChart()
			m.updateMachinesChart()
		}
		if len(m.hourlyData) > 0 {
			m.updateSparkline()
		}

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
			return m, tea.Batch(
				fetchStatsCmd(m.client, m.rangeStr),
				fetchDurationsCmd(m.client),
				fetchSummaryCmd(m.client),
				m.spinner.Tick,
			)
		case key.Matches(msg, m.keys.Toggle1):
			m.showLanguages = !m.showLanguages
			return m, nil
		case key.Matches(msg, m.keys.Toggle2):
			m.showProjects = !m.showProjects
			return m, nil
		case key.Matches(msg, m.keys.Toggle3):
			m.showSparkline = !m.showSparkline
			return m, nil
		case key.Matches(msg, m.keys.Toggle4):
			m.showHeatmap = !m.showHeatmap
			return m, nil
		case key.Matches(msg, m.keys.Toggle5):
			m.showCategories = !m.showCategories
			return m, nil
		case key.Matches(msg, m.keys.Toggle6):
			m.showEditors = !m.showEditors
			return m, nil
		case key.Matches(msg, m.keys.Toggle7):
			m.showOS = !m.showOS
			return m, nil
		case key.Matches(msg, m.keys.Toggle8):
			m.showMachines = !m.showMachines
			return m, nil
		case key.Matches(msg, m.keys.Toggle9):
			m.showSummary = !m.showSummary
			return m, nil
		case key.Matches(msg, m.keys.ShowAll):
			m.showLanguages = true
			m.showProjects = true
			m.showSparkline = true
			m.showHeatmap = true
			m.showCategories = true
			m.showEditors = true
			m.showOS = true
			m.showMachines = true
			m.showSummary = true
			return m, nil
		case key.Matches(msg, m.keys.HideAll):
			m.showLanguages = false
			m.showProjects = false
			m.showSparkline = false
			m.showHeatmap = false
			m.showCategories = false
			m.showEditors = false
			m.showOS = false
			m.showMachines = false
			m.showSummary = false
			return m, nil
		case key.Matches(msg, m.keys.ChangeTheme):
			m.showPicker = true
			m.picker = NewThemePicker(false) // false = not first run, cancel allowed
			// Pre-select current theme in picker
			for i, name := range theme.AllThemes() {
				if name == m.theme.Name || strings.ToLower(m.theme.Name) == name {
					m.picker.selectedIdx = i
					break
				}
			}
			return m, nil
		}
		return m, nil

	case statsFetchedMsg:
		m.loading = false
		m.stats = msg.stats
		m.err = nil
		m.rateLimited = false // Clear rate limit indicator on success
		m.lastFetch = time.Now()
		m.nextRefresh = time.Now().Add(m.refreshInterval)
		m.updateLanguagesChart()
		m.updateProjectsChart()
		m.updateCategoriesChart()
		m.updateEditorsChart()
		m.updateOSChart()
		m.updateMachinesChart()
		return m, scheduleRefresh(m.refreshInterval)

	case durationsFetchedMsg:
		m.hourlyData = groupDurationsByHour(msg.durations.Data)
		m.updateSparkline()
		return m, nil

	case summaryFetchedMsg:
		m.summaryData = msg.summary
		return m, nil

	case fetchErrMsg:
		m.loading = false
		m.err = msg.err
		m.rateLimited = strings.Contains(msg.err.Error(), "429")
		m.nextRefresh = time.Now().Add(m.refreshInterval)
		return m, scheduleRefresh(m.refreshInterval)

	case refreshMsg:
		// Time to refresh - kick off new fetch
		m.loading = true
		return m, tea.Batch(
			fetchStatsCmd(m.client, m.rangeStr),
			fetchDurationsCmd(m.client),
			fetchSummaryCmd(m.client),
			m.spinner.Tick,
		)

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

	// Show picker if active
	if m.showPicker {
		return m.picker.View()
	}

	// Check minimum terminal size
	const minWidth = 40
	const minHeight = 10
	if m.width < minWidth || m.height < minHeight {
		errorStyle := lipgloss.NewStyle().
			Foreground(m.theme.Error).
			Bold(true)
		dimStyle := lipgloss.NewStyle().
			Foreground(m.theme.Dim)

		var sb strings.Builder
		sb.WriteString("\n")
		sb.WriteString(errorStyle.Render("Terminal Too Small") + "\n\n")
		sb.WriteString(dimStyle.Render(fmt.Sprintf(
			"Current size:  %d cols x %d rows\n"+
				"Required:      %d cols x %d rows\n\n"+
				"Please resize your terminal window.\n"+
				"The dashboard will adjust automatically.",
			m.width, m.height, minWidth, minHeight,
		)))
		return sb.String()
	}

	if m.showHelp {
		return m.renderHelp()
	}

	if m.loading {
		return fmt.Sprintf("\n  %s Fetching stats...\n", m.spinner.View())
	}

	if m.err != nil {
		return ErrorStyle(m.theme).Render(fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.", m.err))
	}

	return m.renderDashboard()
}

// renderDashboard renders the full stats dashboard.
func (m Model) renderDashboard() string {
	statusBar := m.renderStatusBar()

	// Use renderDashboardLayout for all content
	content := m.renderDashboardLayout()

	// Account for border (-2) and status bar height.
	panelHeight := m.height - lipgloss.Height(statusBar) - 4
	if panelHeight < 1 {
		panelHeight = 1
	}

	statsPanel := BorderStyle(m.theme).
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
	title := TitleStyle(m.theme).Render(fmt.Sprintf("WakaTime Stats (%s)", m.rangeStr))
	sb.WriteString(title + "\n\n")

	// Totals
	sb.WriteString(fmt.Sprintf("  Total time:    %s\n", data.HumanReadableTotal))
	sb.WriteString(fmt.Sprintf("  Daily average: %s\n", data.HumanReadableDailyAverage))
	sb.WriteString("\n")

	// Calculate panel widths for 2-column layout
	panelWidth := (m.width / 2) - 6
	if panelWidth < 20 {
		panelWidth = 20
	}

	// Build visible panels
	var panels []string

	// Left panel: Languages chart
	if m.showLanguages {
		var leftPanel strings.Builder
		leftPanel.WriteString(TitleStyle(m.theme).Render("Languages") + "\n")
		if len(data.Languages) > 0 {
			leftPanel.WriteString(m.languagesChart.View())
		} else {
			leftPanel.WriteString("  No data")
		}
		panels = append(panels, leftPanel.String())
	}

	// Right panel: Projects chart
	if m.showProjects {
		var rightPanel strings.Builder
		rightPanel.WriteString(TitleStyle(m.theme).Render("Projects") + "\n")
		if len(data.Projects) > 0 {
			rightPanel.WriteString(m.projectsChart.View())
		} else {
			rightPanel.WriteString("  No data")
		}
		panels = append(panels, rightPanel.String())
	}

	// Join panels horizontally if both visible, otherwise just show the one
	if len(panels) > 0 {
		charts := lipgloss.JoinHorizontal(lipgloss.Top, panels...)
		sb.WriteString(charts)
	}

	return sb.String()
}

// renderStatusBar renders the bottom status line.
func (m Model) renderStatusBar() string {
	var status string
	if m.rateLimited {
		status = WarningStyle(m.theme).Render("Rate limited - retrying with backoff...")
	} else if m.loading {
		status = m.spinner.View() + " Fetching..."
	} else if m.err != nil {
		status = ErrorStyle(m.theme).Render("Error: " + m.err.Error())
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

	helpHint := DimStyle(m.theme).Render("? help  1-9 panels  a/h all  r refresh  q quit")
	gap := strings.Repeat(" ", max(0, m.width-lipgloss.Width(status)-lipgloss.Width(helpHint)))
	return DimStyle(m.theme).Render(status) + gap + helpHint
}

// renderHelp renders the help overlay showing keyboard shortcuts.
func (m Model) renderHelp() string {
	title := TitleStyle(m.theme).Render("Keyboard Shortcuts")
	helpText := m.help.View(m.keys)
	hint := DimStyle(m.theme).Render("\nPress ? to return to dashboard")
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

// groupDurationsByHour aggregates durations into 24 hourly buckets.
func groupDurationsByHour(durations []types.Duration) []float64 {
	hourly := make([]float64, 24)
	for _, d := range durations {
		t := time.Unix(int64(d.Time), 0)
		hour := t.Hour()
		hourly[hour] += d.Duration / 3600.0 // Convert to hours
	}
	return hourly
}

// updateSparkline updates the sparkline chart with current hourly data.
func (m *Model) updateSparkline() {
	m.sparklineChart.Clear()
	m.sparklineChart.PushAll(m.hourlyData)
	m.sparklineChart.Draw()
}

// renderSparkline renders the sparkline chart showing hourly activity.
func (m Model) renderSparkline() string {
	sparklineTitle := TitleStyle(m.theme).Render("\nHourly Activity (Today)")
	return lipgloss.JoinVertical(lipgloss.Left, sparklineTitle, m.sparklineChart.View())
}

// renderHeatmapPanel renders the heatmap section with title.
func (m Model) renderHeatmapPanel() string {
	heatmapTitle := TitleStyle(m.theme).Render("\nActivity (Last 7 Days)")
	heatmapContent := m.renderHeatmap()
	return lipgloss.JoinVertical(lipgloss.Left, heatmapTitle, heatmapContent)
}

// updateLanguagesChart updates the languages bar chart with current stats.
func (m *Model) updateLanguagesChart() {
	if m.stats == nil {
		return
	}

	m.languagesChart.Clear()
	data := m.stats.Data

	// Add top 5 languages
	limit := 5
	if len(data.Languages) < limit {
		limit = len(data.Languages)
	}

	for _, lang := range data.Languages[:limit] {
		hours := lang.TotalSeconds / 3600.0
		color := getLanguageColor(lang.Name)
		barStyle := lipgloss.NewStyle().Foreground(color)
		m.languagesChart.Push(barchart.BarData{
			Label: lang.Name,
			Values: []barchart.BarValue{
				{
					Name:  "",
					Value: hours,
					Style: barStyle,
				},
			},
		})
	}

	m.languagesChart.Draw()
}

// updateProjectsChart updates the projects bar chart with current stats.
func (m *Model) updateProjectsChart() {
	if m.stats == nil {
		return
	}

	m.projectsChart.Clear()
	data := m.stats.Data

	// Add top 5 projects
	limit := 5
	if len(data.Projects) < limit {
		limit = len(data.Projects)
	}

	// Fixed cyan color for all projects
	projectColor := lipgloss.Color("#00d7ff")
	barStyle := lipgloss.NewStyle().Foreground(projectColor)

	for _, proj := range data.Projects[:limit] {
		hours := proj.TotalSeconds / 3600.0
		m.projectsChart.Push(barchart.BarData{
			Label: proj.Name,
			Values: []barchart.BarValue{
				{
					Name:  "",
					Value: hours,
					Style: barStyle,
				},
			},
		})
	}

	m.projectsChart.Draw()
}

// renderHeatmap renders a GitHub-style activity heatmap for the last 7 days.
func (m Model) renderHeatmap() string {
	if m.summaryData == nil || len(m.summaryData.Data) == 0 {
		return DimStyle(m.theme).Render("No activity data")
	}

	var blocks []string
	for _, day := range m.summaryData.Data {
		hours := day.GrandTotal.TotalSeconds / 3600.0
		color := getThemedActivityColor(hours, m.theme)
		// Unicode block character with day label
		label := day.Range.Date[5:] // MM-DD format
		block := lipgloss.NewStyle().
			Background(color).
			Foreground(m.theme.Foreground).
			Padding(0, 1).
			Render(label)
		blocks = append(blocks, block)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, blocks...)
}

// getThemedActivityColor returns a theme-aware contribution color based on hours coded.
func getThemedActivityColor(hours float64, t theme.Theme) lipgloss.Color {
	switch {
	case hours < 0.5:
		return t.HeatmapColors[0] // None
	case hours < 2:
		return t.HeatmapColors[1] // Low
	case hours < 4:
		return t.HeatmapColors[2] // Medium
	case hours < 6:
		return t.HeatmapColors[3] // High
	default:
		return t.HeatmapColors[4] // VeryHigh
	}
}
