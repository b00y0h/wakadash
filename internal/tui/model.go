package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/b00y0h/wakadash/internal/api"
	"github.com/b00y0h/wakadash/internal/datasource"
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

	// DataSource for hybrid data fetching (API for recent, archive for old dates)
	dataSource *datasource.DataSource

	// Archived data for historical dates
	archiveData      *types.DayData
	archiveDayTotals [7]float64 // Per-day totals (Sun=0..Sat=6) for historical week

	// Prefetch cache for instant navigation
	prefetchedData map[string]cachedWeekData // Cache: weekStart -> data + daily totals

	// Date navigation - week-based (Sunday to Saturday)
	selectedWeekStart string // Start of currently viewed week (YYYY-MM-DD, always a Sunday), empty = current week
	atOldestData      bool   // True when viewing the oldest available data

	// End-of-history state
	showEndOfHistory bool   // True when navigated to week with no data
	oldestDataDate   string // Date when archive data started (for banner display)

	// Theme
	theme theme.Theme // Active color theme

	// UI components
	spinner        spinner.Model
	help           help.Model
	keys           keymap
	sparklineChart sparkline.Model

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

	// Weekly browser
	showWeeklyBrowser bool               // True when showing weekly history browser
	weeklyBrowser     WeeklyBrowserModel // Weekly browser sub-model

	// State
	quitting    bool
	showHelp    bool
	rateLimited bool // Visual indicator for rate limit status
}

// cachedWeekData holds prefetched data and daily totals for a week.
type cachedWeekData struct {
	data        *types.DayData
	dailyTotals [7]float64
}

// NewModel creates a new Model with the given API client, time range, and refresh interval.
// rangeStr defaults to "last_7_days" if empty.
// Valid values: last_7_days, last_30_days, last_6_months, last_year, all_time.
// refreshInterval defaults to 60s if zero.
// dataSource routes fetches to API (recent dates) or archive (old dates).
func NewModel(client *api.Client, rangeStr string, refreshInterval time.Duration, dataSource *datasource.DataSource) Model {
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

	return Model{
		// Safe defaults — overridden by WindowSizeMsg before meaningful renders.
		width:             80,
		height:            24,
		loading:           true,
		client:            client,
		rangeStr:          rangeStr,
		refreshInterval:   refreshInterval,
		dataSource:        dataSource,
		prefetchedData:    make(map[string]cachedWeekData),
		selectedWeekStart: "", // Empty means current week (live data)
		atOldestData:      false,
		theme:             activeTheme,
		spinner:           s,
		help:              h,
		keys:              defaultKeymap,
		sparklineChart:    sparklineChart,
		showSummary:       true,
		showLanguages:     true,
		showProjects:      true,
		showSparkline:     true,
		showHeatmap:       true,
		showCategories:    true,
		showEditors:       true,
		showOS:            true,
		showMachines:      true,
	}
}

// Init starts the initial async stats fetch, spinner animation, and countdown ticker.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		fetchStatsCmd(m.client, m.rangeStr),
		fetchDurationsCmd(m.client),
		fetchSummaryCmd(m.client),
		m.spinner.Tick,
		tickEverySecond(),
	}

	// Fetch today's data using hybrid DataSource (API for today, since it's recent)
	if m.dataSource != nil {
		today := time.Now().Format("2006-01-02")
		cmds = append(cmds, fetchDataCmd(m.dataSource, today))
	}

	return tea.Batch(cmds...)
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

	// Delegate to weekly browser when in browser mode
	if m.showWeeklyBrowser {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.weeklyBrowser.width = msg.Width
			m.weeklyBrowser.height = msg.Height
			m.width = msg.Width
			m.height = msg.Height
		case weeklyDataFetchedMsg:
			// Forward data to browser
			newBrowser, _ := m.weeklyBrowser.Update(msg)
			m.weeklyBrowser = newBrowser.(WeeklyBrowserModel)
		case tea.KeyMsg:
			newBrowser, _ := m.weeklyBrowser.Update(msg)
			m.weeklyBrowser = newBrowser.(WeeklyBrowserModel)
			if m.weeklyBrowser.IsConfirmed() {
				// User selected a week — navigate to it
				selectedWeek := m.weeklyBrowser.SelectedWeek()
				m.showWeeklyBrowser = false
				// Check if selected week is current week
				currentWeekStart := getWeekStart(time.Now()).Format("2006-01-02")
				if selectedWeek == currentWeekStart {
					m.selectedWeekStart = ""
					m.atOldestData = false
					return m, fetchDataCmd(m.dataSource, time.Now().Format("2006-01-02"))
				}
				// Navigate to historical week
				m.selectedWeekStart = selectedWeek
				m.atOldestData = !m.dataSource.HasOlderData(selectedWeek)
				m.showEndOfHistory = false
				return m, fetchDataCmd(m.dataSource, selectedWeek)
			}
			if m.weeklyBrowser.IsCancelled() {
				m.showWeeklyBrowser = false
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		// Resize sparkline for full width
		fullWidth := msg.Width - 4
		sparklineHeight := 5
		m.sparklineChart.Resize(fullWidth, sparklineHeight)

		// Redraw sparkline with new dimensions
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
			m.picker = NewThemePicker(false, m.width, m.height) // false = not first run, cancel allowed
			// Pre-select current theme in picker
			for i, name := range theme.AllThemes() {
				if name == m.theme.Name || strings.ToLower(m.theme.Name) == name {
					m.picker.selectedIdx = i
					break
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.WeeklyBrowser):
			// Only open if dataSource is available (needs archive access)
			if m.dataSource != nil {
				m.showWeeklyBrowser = true
				m.weeklyBrowser = NewWeeklyBrowser(m.theme, m.width, m.height)
				return m, fetchWeeklySummariesCmd(m.dataSource, 52)
			}
			return m, nil
		case key.Matches(msg, m.keys.PrevDay):
			// Navigate to previous week with data (auto-skip blank weeks)
			var searchStart time.Time
			if m.selectedWeekStart == "" {
				searchStart = getWeekStart(time.Now())
			} else {
				parsed, err := time.Parse("2006-01-02", m.selectedWeekStart)
				if err != nil {
					return m, nil
				}
				searchStart = parsed
			}
			// Start search from week before current
			prevWeekStart := searchStart.AddDate(0, 0, -7).Format("2006-01-02")

			// Check cache first
			if cached, ok := m.prefetchedData[prevWeekStart]; ok {
				if cached.data == nil {
					// No data for this week - show end-of-history banner
					m.showEndOfHistory = true
					m.oldestDataDate = m.selectedWeekStart // Last week with data
					return m, nil
				}
				m.archiveData = cached.data
				m.archiveDayTotals = cached.dailyTotals
				m.selectedWeekStart = prevWeekStart
				m.atOldestData = !m.dataSource.HasOlderData(prevWeekStart)
				// Trigger prefetch for week before this one
				nextPrefetch := getPreviousWeekStart(prevWeekStart)
				if nextPrefetch != "" {
					if _, cached := m.prefetchedData[nextPrefetch]; !cached {
						return m, prefetchWeekCmd(m.dataSource, nextPrefetch)
					}
				}
				return m, nil
			}
			// Fall through to existing fetch logic if not cached
			m.loading = true
			return m, findNonEmptyWeekCmd(m.dataSource, prevWeekStart, -1)
		case key.Matches(msg, m.keys.NextDay):
			m.atOldestData = false // Moving forward, not at oldest
			// Clear end-of-history state when navigating forward
			if m.showEndOfHistory {
				m.showEndOfHistory = false
			}
			// Navigate to next week (capped at current week)
			if m.selectedWeekStart == "" {
				// Already at current week, can't go forward
				return m, nil
			}
			parsed, err := time.Parse("2006-01-02", m.selectedWeekStart)
			if err != nil {
				return m, nil
			}
			nextWeekStart := parsed.AddDate(0, 0, 7)
			currentWeekStart := getWeekStart(time.Now())
			// Compare dates as strings to avoid timezone/time-of-day mismatches
			nextDate := nextWeekStart.Format("2006-01-02")
			currentDate := currentWeekStart.Format("2006-01-02")
			if nextDate >= currentDate {
				// Reached current week, return to live view
				m.selectedWeekStart = ""
			} else {
				m.selectedWeekStart = nextDate
			}
			dateToFetch := nextDate
			if m.selectedWeekStart == "" {
				dateToFetch = time.Now().Format("2006-01-02")
			}
			return m, fetchDataCmd(m.dataSource, dateToFetch)
		case key.Matches(msg, m.keys.Today):
			if m.selectedWeekStart == "" && !m.showEndOfHistory {
				return m, nil // Already at current week
			}
			// Per user decision: Today key (0/Home) from end-of-history: Jump directly to today
			m.selectedWeekStart = ""
			m.atOldestData = false
			m.showEndOfHistory = false
			m.oldestDataDate = ""
			// Returning to today enables auto-refresh (isViewingHistory becomes false)
			return m, fetchDataCmd(m.dataSource, time.Now().Format("2006-01-02"))
		}
		return m, nil

	case statsFetchedMsg:
		// Note: Stats from /stats endpoint may be incomplete - summaryFetchedMsg
		// will override with aggregated data from /summaries which is more reliable
		m.loading = false
		m.stats = msg.stats
		m.err = nil
		m.rateLimited = false // Clear rate limit indicator on success
		m.lastFetch = time.Now()
		m.nextRefresh = time.Now().Add(m.refreshInterval)
		return m, scheduleRefresh(m.refreshInterval)

	case durationsFetchedMsg:
		m.hourlyData = groupDurationsByHour(msg.durations.Data)
		m.updateSparkline()
		return m, nil

	case summaryFetchedMsg:
		m.summaryData = msg.summary
		// Aggregate summary data into stats format (more reliable than /stats endpoint)
		if aggregated := types.AggregateFromSummary(msg.summary); aggregated != nil {
			m.stats = &types.StatsResponse{Data: *aggregated}
		}
		return m, nil

	case archiveFetchedMsg:
		m.archiveData = msg.data
		// Data may be nil if archive not found (404) - that's graceful, not an error
		// Future phases will use this data for historical date navigation
		return m, nil

	case dataFetchedMsg:
		// Only update archive fields when viewing historical data
		if m.selectedWeekStart != "" {
			m.archiveData = msg.data
			m.archiveDayTotals = msg.dailyTotals
		} else {
			// Live mode: update archiveData (used by getActiveStatsData fallback)
			// but don't touch archiveDayTotals
			m.archiveData = msg.data
		}

		// If navigating to historical week and no data exists
		if m.selectedWeekStart != "" && msg.data == nil {
			m.showEndOfHistory = true
			m.oldestDataDate = msg.date
			m.loading = false
			return m, nil
		}

		// Trigger background prefetch of previous week
		prevWeek := getPreviousWeekStart(m.selectedWeekStart)
		if prevWeek != "" {
			if _, cached := m.prefetchedData[prevWeek]; !cached {
				return m, prefetchWeekCmd(m.dataSource, prevWeek)
			}
		}
		return m, nil

	case weekSearchResultMsg:
		m.loading = false
		if !msg.found {
			m.showEndOfHistory = true
			m.oldestDataDate = m.selectedWeekStart
			m.loading = false
			return m, nil
		}
		// Update to found week
		m.selectedWeekStart = msg.weekStart
		m.atOldestData = msg.atOldest
		return m, fetchDataCmd(m.dataSource, m.selectedWeekStart)

	case prefetchResultMsg:
		// Store result in cache (even nil for no-data weeks)
		// Per user decision: Silent failure — no UI feedback on prefetch errors
		if msg.err == nil {
			m.prefetchedData[msg.weekStart] = cachedWeekData{
				data:        msg.data,
				dailyTotals: msg.dailyTotals,
			}
		}
		return m, nil

	case fetchErrMsg:
		m.loading = false
		m.err = msg.err
		m.rateLimited = strings.Contains(msg.err.Error(), "429")
		m.nextRefresh = time.Now().Add(m.refreshInterval)
		return m, scheduleRefresh(m.refreshInterval)

	case refreshMsg:
		// Skip refresh when viewing historical data - user is browsing archive
		if m.isViewingHistory() {
			// Reschedule but don't fetch - will resume when returning to today
			return m, scheduleRefresh(m.refreshInterval)
		}
		// Time to refresh - kick off new fetch (current week only)
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

	// Show weekly browser if active
	if m.showWeeklyBrowser {
		return m.weeklyBrowser.View()
	}

	// Check for end-of-history state first
	if m.showEndOfHistory {
		return m.renderEndOfHistory()
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

	// Title - show date range in historical mode
	var titleStr string
	if m.selectedWeekStart != "" {
		parsed, err := time.Parse("2006-01-02", m.selectedWeekStart)
		if err == nil {
			titleStr = fmt.Sprintf("WakaTime Stats (%s)", formatWeekRange(parsed))
		} else {
			titleStr = fmt.Sprintf("WakaTime Stats (%s)", m.rangeStr)
		}
	} else {
		titleStr = fmt.Sprintf("WakaTime Stats (%s)", m.rangeStr)
	}
	title := TitleStyle(m.theme).Render(titleStr)
	sb.WriteString(title + "\n\n")

	// Totals
	sb.WriteString(fmt.Sprintf("  Total time:    %s\n", data.HumanReadableTotal))
	sb.WriteString(fmt.Sprintf("  Daily average: %s\n", data.HumanReadableDailyAverage))
	sb.WriteString("\n")

	// Calculate panel width for 2-column layout
	panelWidth := (m.width - 4) / 2
	panelStyle := lipgloss.NewStyle().Width(panelWidth)

	// Build visible panels
	var panels []string

	// Left panel: Languages
	if m.showLanguages {
		panels = append(panels, panelStyle.Render(m.renderLanguagesPanel()))
	}

	// Right panel: Projects
	if m.showProjects {
		panels = append(panels, panelStyle.Render(m.renderProjectsPanel()))
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

	// Show week indicator when viewing historical data
	var weekIndicator string
	if m.selectedWeekStart != "" {
		parsed, err := time.Parse("2006-01-02", m.selectedWeekStart)
		if err == nil {
			weekIndicator = fmt.Sprintf("[%s] ", formatWeekRange(parsed))
		}
	}

	// Show historical data indicator
	var historyIndicator string
	if m.isViewingHistory() {
		historyIndicator = WarningStyle(m.theme).Render("[HISTORICAL] ")
	}

	// Show end-of-history indicator
	var oldestIndicator string
	if m.atOldestData {
		oldestIndicator = WarningStyle(m.theme).Render("[oldest data] ")
	}

	if m.rateLimited {
		status = WarningStyle(m.theme).Render("Rate limited - retrying with backoff...")
	} else if m.loading {
		status = m.spinner.View() + " Fetching..."
	} else if m.err != nil {
		status = ErrorStyle(m.theme).Render("Error: " + m.err.Error())
	} else {
		if m.isViewingHistory() {
			// Show paused indicator when viewing historical data
			status = DimStyle(m.theme).Render("Auto-refresh paused (viewing history)")
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
	}

	// Prepend indicators if viewing historical data
	if historyIndicator != "" || weekIndicator != "" || oldestIndicator != "" {
		status = oldestIndicator + historyIndicator + weekIndicator + status
	}

	helpHint := DimStyle(m.theme).Render("? help  w weeks  1-9 panels  a/h all  r refresh  q quit")
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

// getWeekStart returns the Sunday that starts the week containing the given date.
func getWeekStart(date time.Time) time.Time {
	// Calculate days since Sunday (Sunday = 0)
	daysSinceSunday := int(date.Weekday())
	return date.AddDate(0, 0, -daysSinceSunday)
}

// getPreviousWeekStart returns the Sunday of the previous week.
func getPreviousWeekStart(currentWeek string) string {
	if currentWeek == "" {
		currentWeek = getWeekStart(time.Now()).Format("2006-01-02")
	}
	parsed, err := time.Parse("2006-01-02", currentWeek)
	if err != nil {
		return ""
	}
	return parsed.AddDate(0, 0, -7).Format("2006-01-02")
}

// formatWeekRange returns a display string like "Feb 16-22" for a week starting on Sunday.
func formatWeekRange(weekStart time.Time) string {
	weekEnd := weekStart.AddDate(0, 0, 6) // Saturday
	if weekStart.Month() == weekEnd.Month() {
		return fmt.Sprintf("%s %d-%d", weekStart.Format("Jan"), weekStart.Day(), weekEnd.Day())
	}
	return fmt.Sprintf("%s %d - %s %d", weekStart.Format("Jan"), weekStart.Day(), weekEnd.Format("Jan"), weekEnd.Day())
}

// getActiveStatsData returns StatsData to display based on current view state.
// When viewing historical data (selectedWeekStart != ""), converts archiveData to StatsData.
// When viewing current week (selectedWeekStart == ""), returns m.stats.Data.
func (m Model) getActiveStatsData() *types.StatsData {
	if m.selectedWeekStart != "" && m.archiveData != nil {
		// Convert DayData to StatsData format for rendering
		// archiveData is a single day's data (DayData), need to wrap and aggregate
		summaryWrapper := &types.SummaryResponse{
			Data: []types.DayData{*m.archiveData},
		}
		if aggregated := types.AggregateFromSummary(summaryWrapper); aggregated != nil {
			return aggregated
		}
	}
	// Default: return current stats (API data)
	if m.stats != nil {
		return &m.stats.Data
	}
	return nil
}

// isViewingHistory returns true when viewing historical data (not current week).
func (m Model) isViewingHistory() bool {
	return m.selectedWeekStart != ""
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
// Expands 24 hourly values across the full canvas width so bars span left to right.
func (m *Model) updateSparkline() {
	m.sparklineChart.Clear()
	w := m.sparklineChart.Width()
	if w <= 0 || len(m.hourlyData) == 0 {
		return
	}
	// Stretch 24 hours across the full canvas width.
	expanded := make([]float64, w)
	for col := 0; col < w; col++ {
		hour := col * 24 / w
		if hour >= len(m.hourlyData) {
			hour = len(m.hourlyData) - 1
		}
		expanded[col] = m.hourlyData[hour]
	}
	m.sparklineChart.PushAll(expanded)
	m.sparklineChart.Draw()
}

// renderSparkline renders the sparkline chart showing hourly activity (live)
// or daily activity breakdown (historical).
func (m Model) renderSparkline() string {
	if m.selectedWeekStart != "" {
		return m.renderWeeklySparkline()
	}

	content := m.sparklineChart.View()
	w := m.sparklineChart.Width()

	// Build hour labels aligned with stretched bar positions.
	keyHours := []int{0, 3, 6, 9, 12, 15, 18, 21}
	labelRow := make([]byte, w)
	for i := range labelRow {
		labelRow[i] = ' '
	}
	for _, h := range keyHours {
		col := h * w / 24
		label := fmt.Sprintf("%d", h)
		for j := 0; j < len(label) && col+j < w; j++ {
			labelRow[col+j] = label[j]
		}
	}
	styled := DimStyle(m.theme).Render(string(labelRow))

	content = content + "\n" + styled
	return renderBorderedPanel("Hourly Activity (Today)", content, m.width-4, m.theme)
}

// renderWeeklySparkline renders a daily activity bar chart for historical weeks.
func (m Model) renderWeeklySparkline() string {
	dayLabels := []string{"S", "M", "T", "W", "T", "F", "S"}
	maxSeconds := 0.0
	for _, s := range m.archiveDayTotals {
		if s > maxSeconds {
			maxSeconds = s
		}
	}

	barHeight := 6
	panelWidth := m.width - 4
	colWidth := panelWidth / 7

	var rows []string
	for row := barHeight; row >= 1; row-- {
		threshold := maxSeconds * float64(row) / float64(barHeight)
		var cols []string
		for _, s := range m.archiveDayTotals {
			bar := strings.Repeat(" ", colWidth)
			if maxSeconds > 0 && s >= threshold {
				blockWidth := colWidth / 2
				pad := (colWidth - blockWidth) / 2
				bar = strings.Repeat(" ", pad) +
					lipgloss.NewStyle().Background(m.theme.Primary).Render(strings.Repeat(" ", blockWidth)) +
					strings.Repeat(" ", colWidth-pad-blockWidth)
			}
			cols = append(cols, bar)
		}
		rows = append(rows, strings.Join(cols, ""))
	}

	// Day labels row
	var labels []string
	for _, l := range dayLabels {
		pad := (colWidth - len(l)) / 2
		labels = append(labels, strings.Repeat(" ", pad)+l+strings.Repeat(" ", colWidth-pad-len(l)))
	}
	labelRow := DimStyle(m.theme).Render(strings.Join(labels, ""))

	// Time labels row
	var times []string
	for _, s := range m.archiveDayTotals {
		var label string
		if s == 0 {
			label = ""
		} else {
			h := int(s) / 3600
			mins := (int(s) % 3600) / 60
			if h > 0 {
				label = fmt.Sprintf("%dh", h)
			} else {
				label = fmt.Sprintf("%dm", mins)
			}
		}
		pad := (colWidth - len(label)) / 2
		if pad < 0 {
			pad = 0
		}
		times = append(times, strings.Repeat(" ", pad)+label+strings.Repeat(" ", colWidth-pad-len(label)))
	}
	timeRow := DimStyle(m.theme).Render(strings.Join(times, ""))

	content := strings.Join(rows, "\n") + "\n" + labelRow + "\n" + timeRow

	title := "Daily Activity"
	if parsed, err := time.Parse("2006-01-02", m.selectedWeekStart); err == nil {
		title = fmt.Sprintf("Daily Activity (%s)", formatWeekRange(parsed))
	}
	return renderBorderedPanel(title, content, panelWidth, m.theme)
}

// renderHeatmapPanel renders the heatmap section with title.
func (m Model) renderHeatmapPanel() string {
	if m.selectedWeekStart != "" {
		return m.renderHistoricalHeatmapPanel()
	}
	heatmapContent := m.renderHeatmap()
	return renderBorderedPanel("Activity (Last 7 Days)", heatmapContent, m.width-4, m.theme)
}

// renderHistoricalHeatmapPanel renders the heatmap for a historical week.
func (m Model) renderHistoricalHeatmapPanel() string {
	parsed, err := time.Parse("2006-01-02", m.selectedWeekStart)
	if err != nil {
		return renderBorderedPanel("Activity", DimStyle(m.theme).Render("No data"), m.width-4, m.theme)
	}

	var blocks []string
	for i := 0; i < 7; i++ {
		day := parsed.AddDate(0, 0, i)
		hours := m.archiveDayTotals[i] / 3600.0
		color := getThemedActivityColor(hours, m.theme)
		label := day.Format("01-02") // MM-DD
		block := lipgloss.NewStyle().
			Background(color).
			Foreground(m.theme.Foreground).
			Padding(0, 1).
			Render(label)
		blocks = append(blocks, block)
	}

	title := fmt.Sprintf("Activity (%s)", formatWeekRange(parsed))
	return renderBorderedPanel(title, lipgloss.JoinHorizontal(lipgloss.Top, blocks...), m.width-4, m.theme)
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

// renderEndOfHistory renders the full-screen end-of-history banner.
// Per user decision: Banner text includes date when archive data started.
func (m Model) renderEndOfHistory() string {
	title := EndOfHistoryTitleStyle(m.theme).Render("End of History")

	var dateInfo string
	if m.oldestDataDate != "" {
		dateInfo = EndOfHistoryTextStyle(m.theme).Render(
			fmt.Sprintf("Archive data starts: %s", m.oldestDataDate))
	} else {
		dateInfo = EndOfHistoryTextStyle(m.theme).Render(
			"No archived data available for this period")
	}

	// Per user decision: Show navigation hints
	hints := EndOfHistoryHintStyle(m.theme).Render(
		"Press → or 0 to return")

	content := lipgloss.JoinVertical(lipgloss.Center, title, dateInfo, hints)

	return EndOfHistoryStyle(m.theme, m.width, m.height).Render(content)
}
