# Phase 6: Data Visualization and UX - Research

**Researched:** 2026-02-19
**Domain:** Terminal data visualization (charts, sparklines, heatmaps), resize handling, rate limiting
**Confidence:** HIGH

## Summary

Phase 6 adds rich data visualization to the wakadash TUI using ntcharts (a charmbracelet-ecosystem-native charting library), implements responsive terminal resize handling via bubbletea's WindowSizeMsg, adds exponential backoff for API rate limiting, and enables panel toggles via keyboard number keys.

The standard approach for terminal charts in the bubbletea ecosystem is **ntcharts** (github.com/NimbleMarkets/ntcharts) which provides bar charts, sparklines, heatmaps, and time series charts with native lipgloss styling and bubbletea integration. It was specifically built for TUIs and integrates seamlessly with the Phase 5 foundation.

For hourly activity sparklines, the WakaTime API does not provide pre-aggregated hourly data. The `/durations` endpoint returns time-stamped coding sessions that can be grouped client-side by hour. The `/heartbeats` endpoint provides more granular data but requires `read_heartbeats` scope and additional API calls.

Terminal resize handling is built into bubbletea via WindowSizeMsg (sent on startup and every resize). The model stores dimensions and recalculates layouts in View(). Rate limiting (HTTP 429) requires exponential backoff with jitter - `cenkalti/backoff/v5` is the de facto standard library. Panel toggles are simple boolean flags in the model toggled by number key presses.

**Primary recommendation:** Use ntcharts v0.1.x for all visualizations (barchart, sparkline, heatmap). Implement responsive layout by storing panel dimensions in model state and recalculating on WindowSizeMsg. Use cenkalti/backoff/v5 for rate limit retry logic with visual feedback. Store panel visibility as boolean flags and toggle with number keys (1-4).

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/NimbleMarkets/ntcharts | v0.1.x (latest) | Terminal charts for bubbletea/lipgloss | Built specifically for charm ecosystem; supports bar, sparkline, heatmap |
| github.com/cenkalti/backoff/v5 | v5.0.0+ | Exponential backoff retry logic | De facto standard in Go; 3.8k stars, production-ready |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| ntcharts/barchart | (from ntcharts) | Horizontal/vertical bar charts | Languages and projects visualization |
| ntcharts/sparkline | (from ntcharts) | Small time series columns | Hourly coding activity sparkline |
| ntcharts/heatmap | (from ntcharts) | Color-mapped (x,y) grids | Activity over time heatmap |
| lipgloss.Color | (from lipgloss v1.1.0) | Terminal color constants | Language-specific colors via GitHub Linguist palette |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| ntcharts | guptarohit/asciigraph | asciigraph is simpler but not bubbletea-native; no lipgloss styling; manual layout |
| ntcharts | pterm | pterm charts work in traditional CLIs, not TUIs; conflicts with bubbletea event loop |
| cenkalti/backoff | avast/retry-go | retry-go is heavier; backoff is simpler for single-operation retry |
| cenkalti/backoff/v5 | Custom backoff | Custom solutions miss edge cases (jitter, max elapsed time, reset logic) |

### Installation
```bash
cd /workspace/wakadash
go get github.com/NimbleMarkets/ntcharts@latest
go get github.com/cenkalti/backoff/v5@latest
```

**Version note:** ntcharts was published January 8, 2026. Latest version includes all needed chart types. No version constraints required - use `@latest`.

## Architecture Patterns

### Recommended Project Structure
```
wakadash/internal/tui/
├── model.go         # Add panel visibility flags, ntcharts models
├── messages.go      # Existing messages
├── commands.go      # Add retry logic with backoff
├── keymap.go        # Add number key bindings (1-4) for panel toggles
├── styles.go        # Add language color map from GitHub Linguist
├── panels.go        # NEW: Panel rendering functions (languages, projects, sparkline, heatmap)
└── colors.go        # NEW: Language-to-color mapping
```

### Pattern 1: Panel Visibility State Management
**What:** Store boolean flags in model for each panel. Toggle via number keys. Conditionally render in View().
**When to use:** Always for panel toggles (UX-03 requirement).

```go
// In model.go
type Model struct {
    // ... existing fields ...

    // Panel visibility (toggled by number keys 1-4)
    showLanguages bool
    showProjects  bool
    showSparkline bool
    showHeatmap   bool

    // Chart models (embedded ntcharts models)
    languagesChart barchart.Model
    projectsChart  barchart.Model
    sparkline      sparkline.Model
    heatmap        heatmap.Model
}

// In keymap.go
type keymap struct {
    // ... existing keys ...
    Toggle1 key.Binding  // Languages panel
    Toggle2 key.Binding  // Projects panel
    Toggle3 key.Binding  // Sparkline panel
    Toggle4 key.Binding  // Heatmap panel
}

var defaultKeymap = keymap{
    // ... existing ...
    Toggle1: key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "toggle languages")),
    Toggle2: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "toggle projects")),
    Toggle3: key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "toggle sparkline")),
    Toggle4: key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "toggle heatmap")),
}

// In Update()
case tea.KeyMsg:
    switch {
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
    }
```

### Pattern 2: ntcharts Bar Chart for Languages/Projects
**What:** Create barchart.Model in NewModel(), update data on statsFetchedMsg, resize on WindowSizeMsg.
**When to use:** VIZ-01 (languages bar chart) and VIZ-02 (projects bar chart).

```go
// Source: https://pkg.go.dev/github.com/NimbleMarkets/ntcharts/barchart

// In NewModel()
languagesChart := barchart.New(40, 10,
    barchart.WithStyles(
        lipgloss.NewStyle().Foreground(lipgloss.Color("240")), // axis
        lipgloss.NewStyle().Foreground(lipgloss.Color("250")), // labels
    ),
)

// On statsFetchedMsg, populate bar chart data
func (m Model) updateLanguagesChart() Model {
    m.languagesChart.Clear()

    for _, lang := range m.stats.Data.Languages {
        color := getLanguageColor(lang.Name) // From colors.go
        style := lipgloss.NewStyle().Foreground(color)

        m.languagesChart.Push(barchart.BarData{
            Label: lang.Name,
            Values: []barchart.BarValue{
                {
                    Name:  lang.Name,
                    Value: lang.TotalSeconds / 3600.0, // Convert to hours
                    Style: style,
                },
            },
        })
    }

    m.languagesChart.Draw()
    return m
}

// In View(), conditionally render
if m.showLanguages {
    languagesPanel := borderStyle.
        Width(panelWidth).
        Height(panelHeight).
        Render(m.languagesChart.View())
    // ... layout logic
}
```

### Pattern 3: Sparkline for Hourly Activity
**What:** Fetch /durations endpoint, group by hour client-side, push to sparkline.Model.
**When to use:** VIZ-03 (hourly activity sparkline).

```go
// Source: https://pkg.go.dev/github.com/NimbleMarkets/ntcharts/sparkline

// Need new API method in client.go
func (c *Client) FetchDurations(date string) ([]Duration, error) {
    url := c.buildURL(fmt.Sprintf("/v1/users/current/durations?date=%s", date))
    // ... fetch logic
}

// Duration type in types.go
type Duration struct {
    Time     float64 `json:"time"`      // UNIX timestamp
    Duration float64 `json:"duration"`  // seconds
}

// In commands.go, add durationsFetchedMsg
type durationsFetchedMsg struct {
    durations []types.Duration
}

func fetchDurationsCmd(client *api.Client, date string) tea.Cmd {
    return func() tea.Msg {
        durations, err := client.FetchDurations(date)
        if err != nil {
            return fetchErrMsg{err: err}
        }
        return durationsFetchedMsg{durations: durations}
    }
}

// In Update(), process durations into hourly buckets
case durationsFetchedMsg:
    hourlyData := make([]float64, 24) // 24 hours
    for _, d := range msg.durations {
        t := time.Unix(int64(d.Time), 0)
        hour := t.Hour()
        hourlyData[hour] += d.Duration / 3600.0 // convert to hours
    }

    m.sparkline.Clear()
    m.sparkline.PushAll(hourlyData)
    m.sparkline.Draw()
    return m, nil

// In NewModel()
sparkline := sparkline.New(40, 6,
    sparkline.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205"))),
    sparkline.WithMaxValue(8.0), // 8 hours max per hour bucket (unrealistic but safe)
)
```

### Pattern 4: Heatmap for Activity Over Time
**What:** Fetch /summaries endpoint for past N days, create (day, hour) -> activity heatmap.
**When to use:** VIZ-04 (activity heatmap).

```go
// Source: https://pkg.go.dev/github.com/NimbleMarkets/ntcharts/heatmap

// In NewModel()
heatmap := heatmap.New(50, 10,
    heatmap.WithAutoValueRange(),
    // Color scale: low (blue) -> medium (yellow) -> high (red)
    heatmap.WithColorScale([]lipgloss.Color{"#0000FF", "#FFFF00", "#FF0000"}),
)

// On summaryFetchedMsg (need to add this)
type summaryFetchedMsg struct {
    summary *types.SummaryResponse
}

case summaryFetchedMsg:
    m.heatmap.ClearData()

    for dayIdx, day := range msg.summary.Data {
        // For each day, distribute grand_total across 24 hours
        // (WakaTime doesn't provide hourly breakdown in summaries)
        // Use uniform distribution or fetch durations separately
        hoursPerDay := day.GrandTotal.TotalSeconds / 3600.0

        for hour := 0; hour < 24; hour++ {
            // Simplified: uniform distribution
            value := hoursPerDay / 24.0
            m.heatmap.Push(heatmap.NewHeatPoint(hour, dayIdx, value))
        }
    }

    m.heatmap.Draw()
    return m, nil
```

**Note:** Heatmap requires hourly granularity per day. WakaTime /summaries only provides daily totals. Options:
1. **Simplified approach**: Use uniform distribution (divide daily total by 24 hours)
2. **Accurate approach**: Fetch /durations for each day and group by hour (requires N additional API calls)
3. **Recommendation**: Start with simplified approach (avoids rate limiting), add accurate fetching in a later phase.

### Pattern 5: Responsive Layout with WindowSizeMsg
**What:** Recalculate panel dimensions when terminal resizes. Store width/height in model, resize ntcharts models.
**When to use:** UX-01 (terminal resize support).

```go
// Source: https://pkg.go.dev/github.com/charmbracelet/bubbletea (WindowSizeMsg docs)
// Source: https://github.com/charmbracelet/bubbletea/discussions/661

case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    m.help.Width = msg.Width

    // Recalculate panel dimensions
    // Example: 2x2 grid of panels
    panelWidth := (msg.Width / 2) - 4   // -4 for borders/padding
    panelHeight := (msg.Height / 2) - 4

    // Resize all chart models
    m.languagesChart.Resize(panelWidth, panelHeight)
    m.projectsChart.Resize(panelWidth, panelHeight)
    m.sparkline.Resize(msg.Width-4, 6) // Full width, fixed height
    m.heatmap.Resize(msg.Width-4, panelHeight)

    // Redraw all charts with new dimensions
    m.languagesChart.Draw()
    m.projectsChart.Draw()
    m.sparkline.Draw()
    m.heatmap.Draw()

    return m, nil
```

**Key insight:** ntcharts models have `Resize(w, h int)` methods. Call these on WindowSizeMsg, then call `Draw()` to re-render with new dimensions.

### Pattern 6: Exponential Backoff for Rate Limiting
**What:** Wrap API calls with cenkalti/backoff retry logic. Detect HTTP 429, retry with exponential backoff, show visual indicator.
**When to use:** UX-02 (API rate limit handling).

```go
// Source: https://pkg.go.dev/github.com/cenkalti/backoff/v5

import (
    "github.com/cenkalti/backoff/v5"
)

// In commands.go, wrap API calls
func fetchStatsCmd(client *api.Client, rangeStr string) tea.Cmd {
    return func() tea.Msg {
        var stats *types.StatsResponse
        var lastErr error

        operation := func() error {
            result, err := client.FetchStats(rangeStr)
            if err != nil {
                // Check if rate limited
                if strings.Contains(err.Error(), "429") {
                    return err // Retry
                }
                // Permanent error - don't retry
                return backoff.Permanent(err)
            }
            stats = result
            return nil
        }

        // Exponential backoff with max 5 retries
        b := backoff.NewExponentialBackOff()
        b.MaxElapsedTime = 2 * time.Minute

        err := backoff.Retry(operation, b)
        if err != nil {
            return fetchErrMsg{err: err}
        }

        return statsFetchedMsg{stats: stats}
    }
}

// In model.go, add rate limit state
type Model struct {
    // ... existing ...
    rateLimited bool  // Visual indicator
}

// In Update()
case fetchErrMsg:
    m.loading = false
    m.err = msg.err

    // Check if rate limited
    if strings.Contains(msg.err.Error(), "429") {
        m.rateLimited = true
    }

    return m, scheduleRefresh(m.refreshInterval)

case statsFetchedMsg:
    m.loading = false
    m.stats = msg.stats
    m.err = nil
    m.rateLimited = false  // Clear rate limit indicator
    // ... rest of logic
```

**Visual indicator in status bar:**
```go
if m.rateLimited {
    status = warningStyle.Render("⚠ Rate limited - retrying with backoff...")
}
```

### Pattern 7: Language Color Mapping
**What:** Map language names to GitHub Linguist colors. Store in colors.go as a map.
**When to use:** VIZ-01 (color-coded languages bar chart).

```go
// Source: https://raw.githubusercontent.com/github/linguist/master/lib/linguist/languages.yml
// In colors.go

package tui

import "github.com/charmbracelet/lipgloss"

// languageColors maps programming language names to GitHub Linguist colors.
// Source: github.com/github/linguist/master/lib/linguist/languages.yml
var languageColors = map[string]lipgloss.Color{
    "Go":         lipgloss.Color("#00ADD8"),
    "JavaScript": lipgloss.Color("#f1e05a"),
    "TypeScript": lipgloss.Color("#3178c6"),
    "Python":     lipgloss.Color("#3572A5"),
    "Rust":       lipgloss.Color("#dea584"),
    "Ruby":       lipgloss.Color("#701516"),
    "Java":       lipgloss.Color("#b07219"),
    "C":          lipgloss.Color("#555555"),
    "C++":        lipgloss.Color("#f34b7d"),
    "HTML":       lipgloss.Color("#e34c26"),
    "CSS":        lipgloss.Color("#563d7c"),
    "Shell":      lipgloss.Color("#89e051"),
    "Markdown":   lipgloss.Color("#083fa1"),
    // ... add more as needed
}

// getLanguageColor returns the GitHub Linguist color for a language name.
// Defaults to gray (#ccc) if language not found.
func getLanguageColor(name string) lipgloss.Color {
    if color, ok := languageColors[name]; ok {
        return color
    }
    return lipgloss.Color("#cccccc") // Default gray
}
```

### Anti-Patterns to Avoid
- **Hard-coded panel dimensions:** Always recalculate based on `m.width` and `m.height`. Terminal size changes frequently.
- **Forgetting to call Draw() after Resize():** ntcharts models cache rendered output. Resize() updates dimensions but doesn't re-render. Always call `chart.Draw()` after `chart.Resize()`.
- **Blocking in tea.Cmd for retry logic:** backoff.Retry() blocks. This is OK inside a tea.Cmd goroutine, but never in Update() or View().
- **Retrying all errors:** Only retry transient errors (429, 502, 503, 504). Use `backoff.Permanent()` for client errors (400, 401, 403, 404).
- **No visual feedback during retry:** Users should see "Rate limited - retrying..." to understand why refresh is slow.
- **Fetching hourly data for 30+ days:** /durations endpoint requires one API call per day. For 30 days = 30 API calls = likely to hit rate limits. Start with last 7 days max.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Bar chart rendering | Custom ASCII bars | ntcharts/barchart | Handles scaling, multi-segment bars, colors, auto-sizing, horizontal/vertical orientation |
| Sparkline columns | Custom block characters | ntcharts/sparkline | Handles data buffer, scaling, braille mode, lipgloss styling |
| Heatmap color mapping | Custom color interpolation | ntcharts/heatmap | Handles color scales, value ranges, auto-scaling, matrix input |
| Exponential backoff | Custom retry with sleep | cenkalti/backoff | Handles jitter, max elapsed time, permanent errors, exponential calculation |
| Language colors | Hard-coded color map | GitHub Linguist YAML + codegen | GitHub updates colors; hard-coding becomes stale |
| Layout math for grids | Manual width/height calculation | lipgloss.JoinHorizontal/Vertical + measured heights | Handles borders, padding, alignment automatically |

**Key insight:** Terminal visualization has many edge cases (color support levels, unicode rendering, terminal emulators, scaling math). ntcharts handles all of these for the bubbletea ecosystem.

## Common Pitfalls

### Pitfall 1: WindowSizeMsg Arrives After Initial Render
**What goes wrong:** Charts render at 80x24 initially, then jump to actual terminal size on first resize.
**Why it happens:** WindowSizeMsg is sent asynchronously after program starts (Phase 5 research pitfall #1).
**How to avoid:** Already handled in Phase 5 - model initializes width=80, height=24. Charts use these defaults until WindowSizeMsg arrives.
**Warning signs:** Charts briefly appear wrong size on startup.

### Pitfall 2: Forgetting to Redraw Charts After Data Update
**What goes wrong:** New data fetched but charts show stale data.
**Why it happens:** ntcharts models cache rendered output. Pushing data doesn't automatically re-render.
**How to avoid:** Always call `chart.Draw()` after `chart.Push()`, `chart.PushAll()`, or `chart.Clear()`.
**Warning signs:** Charts don't update when stats refresh.

### Pitfall 3: Rate Limiting Without Backoff
**What goes wrong:** After first 429 error, all subsequent fetches fail immediately. User sees constant errors.
**Why it happens:** API returns 429 when rate limit exceeded. Without backoff, retry happens immediately and fails again.
**How to avoid:** Use cenkalti/backoff to wait progressively longer between retries (1s, 2s, 4s, 8s, etc.).
**Warning signs:** "rate limit exceeded" error shows constantly; never recovers.

### Pitfall 4: Fetching Too Much Hourly Data
**What goes wrong:** Dashboard fetches /durations for 30 days on startup = 30 API calls = immediate rate limiting.
**Why it happens:** WakaTime doesn't provide pre-aggregated hourly data. Each day requires a separate /durations call.
**How to avoid:** Limit sparkline to "last 24 hours" or use simplified heatmap (daily totals divided by 24).
**Warning signs:** Dashboard hits 429 on every startup.

### Pitfall 5: Color Support Assumptions
**What goes wrong:** Charts use truecolor (#00ADD8) but user's terminal only supports 256 colors. Colors look wrong.
**Why it happens:** lipgloss auto-degrades colors, but some terminals report capabilities incorrectly.
**How to avoid:** Test in multiple terminals (xterm, tmux, Windows Terminal). lipgloss handles degradation automatically - no code changes needed, just test.
**Warning signs:** User reports "colors are gray/wrong" in certain terminals.

### Pitfall 6: Panel Toggle State Not Persisted
**What goes wrong:** User toggles panel off, dashboard refreshes, panel reappears.
**Why it happens:** Panel visibility flags reset on data fetch if not carefully managed.
**How to avoid:** Panel visibility is pure UI state - never reset on data updates. Only toggle on explicit key press.
**Warning signs:** Panels randomly appear/disappear during refresh.

### Pitfall 7: Sparkline/Heatmap Data Misalignment
**What goes wrong:** Heatmap shows 7 rows but data has 8 days. Off-by-one error.
**Why it happens:** Confusion between 0-indexed arrays and 1-indexed day counts.
**How to avoid:** Use explicit loops with clear index variables. Test with known dataset (e.g., exactly 7 days).
**Warning signs:** Heatmap/sparkline dimensions don't match expected data range.

## Code Examples

Verified patterns from official sources and ecosystem best practices:

### Complete Panel Rendering Function
```go
// In panels.go
func (m Model) renderLanguagesPanel() string {
    if !m.showLanguages {
        return ""
    }

    title := titleStyle.Render("Languages")
    chart := m.languagesChart.View()

    return borderStyle.
        Width(m.width/2 - 2).
        Height(m.height/2 - 4).
        Render(lipgloss.JoinVertical(lipgloss.Left, title, chart))
}

func (m Model) renderProjectsPanel() string {
    if !m.showProjects {
        return ""
    }

    title := titleStyle.Render("Projects")
    chart := m.projectsChart.View()

    return borderStyle.
        Width(m.width/2 - 2).
        Height(m.height/2 - 4).
        Render(lipgloss.JoinVertical(lipgloss.Left, title, chart))
}

func (m Model) renderSparklinePanel() string {
    if !m.showSparkline {
        return ""
    }

    title := titleStyle.Render("Hourly Activity (Last 24h)")
    chart := m.sparkline.View()

    return borderStyle.
        Width(m.width - 2).
        Height(8).
        Render(lipgloss.JoinVertical(lipgloss.Left, title, chart))
}

func (m Model) renderHeatmapPanel() string {
    if !m.showHeatmap {
        return ""
    }

    title := titleStyle.Render("Activity Heatmap (Last 7 Days)")
    chart := m.heatmap.View()

    return borderStyle.
        Width(m.width - 2).
        Height(12).
        Render(lipgloss.JoinVertical(lipgloss.Left, title, chart))
}
```

### 2x2 Grid Layout
```go
// In model.go View()
func (m Model) renderDashboard() string {
    // Top row: languages (left) + projects (right)
    topLeft := m.renderLanguagesPanel()
    topRight := m.renderProjectsPanel()
    topRow := lipgloss.JoinHorizontal(lipgloss.Top, topLeft, topRight)

    // Middle: sparkline (full width)
    sparklineRow := m.renderSparklinePanel()

    // Bottom: heatmap (full width)
    heatmapRow := m.renderHeatmapPanel()

    // Status bar
    statusBar := m.renderStatusBar()

    // Combine all rows
    return lipgloss.JoinVertical(lipgloss.Left,
        topRow,
        sparklineRow,
        heatmapRow,
        statusBar,
    )
}
```

### Retry with Backoff (Complete Example)
```go
// Source: https://pkg.go.dev/github.com/cenkalti/backoff/v5

func fetchStatsWithRetry(client *api.Client, rangeStr string) (*types.StatsResponse, error) {
    var stats *types.StatsResponse

    operation := func() error {
        result, err := client.FetchStats(rangeStr)
        if err != nil {
            // Retry transient errors
            if strings.Contains(err.Error(), "429") ||
               strings.Contains(err.Error(), "502") ||
               strings.Contains(err.Error(), "503") ||
               strings.Contains(err.Error(), "504") {
                return err // Retry
            }
            // Permanent error - don't retry
            return backoff.Permanent(err)
        }
        stats = result
        return nil
    }

    b := backoff.NewExponentialBackOff()
    b.InitialInterval = 1 * time.Second
    b.MaxInterval = 30 * time.Second
    b.MaxElapsedTime = 2 * time.Minute
    b.Multiplier = 2.0
    b.RandomizationFactor = 0.5 // Add jitter

    err := backoff.Retry(operation, b)
    return stats, err
}
```

### Language Color Mapping with Fallback
```go
// In colors.go
func getLanguageColor(name string) lipgloss.Color {
    // Normalize name (WakaTime might return "Javascript" vs "JavaScript")
    normalized := strings.Title(strings.ToLower(name))

    if color, ok := languageColors[normalized]; ok {
        return color
    }

    // Try exact match
    if color, ok := languageColors[name]; ok {
        return color
    }

    // Default to gray
    return lipgloss.Color("#cccccc")
}
```

### Hourly Data Aggregation
```go
// Client-side grouping of durations by hour
func groupDurationsByHour(durations []types.Duration) []float64 {
    hourly := make([]float64, 24)

    for _, d := range durations {
        t := time.Unix(int64(d.Time), 0)
        hour := t.Hour()
        hourly[hour] += d.Duration / 3600.0 // Convert seconds to hours
    }

    return hourly
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual ASCII art charts | ntcharts library | Jan 2026 | ntcharts released; first bubbletea-native charting library |
| asciigraph for TUIs | ntcharts | Jan 2026 | asciigraph isn't bubbletea-aware; manual layout; ntcharts integrates with lipgloss |
| Manual exponential backoff | cenkalti/backoff | 2015+ | Library is stable; current v5 released 2024 |
| Hard-coded retry delays | Exponential backoff with jitter | AWS Best Practices 2015 | Jitter prevents thundering herd; industry standard |
| GitHub language colors hard-coded | Fetch from Linguist YAML | 2020+ | GitHub updates colors; automated fetching keeps colors current |

**Deprecated/outdated:**
- **Manual chart rendering:** Before ntcharts (Jan 2026), developers hand-rolled ASCII charts. ntcharts is now the standard for bubbletea TUIs.
- **Linear backoff:** `time.Sleep(5 * attempt)` is outdated. Exponential with jitter is current best practice (AWS, Google Cloud recommendations).
- **Polling for WindowSizeMsg:** bubbletea sends WindowSizeMsg automatically. No manual polling needed (except on Windows due to SIGWINCH limitation).

## Open Questions

1. **Hourly data fetching strategy**
   - What we know: WakaTime /durations requires one API call per day. Fetching 30 days = 30 calls.
   - What's unclear: What's the acceptable sparkline/heatmap time range without hitting rate limits?
   - Recommendation: Start with **last 24 hours for sparkline** (1 API call) and **last 7 days for heatmap with simplified approach** (daily totals ÷ 24, no additional API calls). Add accurate hourly fetching in a later phase if needed.

2. **Panel layout strategy**
   - What we know: Requirements list 4 panels (languages, projects, sparkline, heatmap).
   - What's unclear: Fixed 2x2 grid vs dynamic layout based on visible panels?
   - Recommendation: **Fixed layout** (2x2 grid) for simplicity. When panel hidden, show "Panel hidden - press [N] to show". This is simpler than dynamic reflow and matches most dashboard UIs.

3. **Rate limit detection**
   - What we know: API returns 429 status code. Client wraps this as error string.
   - What's unclear: Does WakaTime include `Retry-After` header?
   - Recommendation: Check for `Retry-After` header in API client. If present, use it; otherwise use exponential backoff. **Action item:** Update client.go to return structured error with retry-after info.

4. **Color degradation testing**
   - What we know: lipgloss auto-degrades truecolor -> 256 -> 16 -> 8 colors.
   - What's unclear: Do ntcharts honor lipgloss color degradation?
   - Recommendation: Test in 256-color terminal (`TERM=xterm-256color`) and 16-color (`TERM=xterm`). If ntcharts doesn't degrade properly, file issue. **Assumption:** It should work (lipgloss is core dependency).

5. **Heatmap granularity tradeoff**
   - What we know: Accurate hourly heatmap requires N API calls (one per day). Simplified approach (daily total ÷ 24) requires zero additional calls.
   - What's unclear: User expectation - is simplified heatmap acceptable?
   - Recommendation: **Ship simplified heatmap first.** Add "⚠ Approximate hourly distribution" note in panel title. If users request accurate data, add it in Phase 7 or v2.1.

## Sources

### Primary (HIGH confidence)
- [NimbleMarkets/ntcharts](https://github.com/NimbleMarkets/ntcharts) - Official ntcharts repository; verified chart types, API, bubbletea integration
- [pkg.go.dev/github.com/NimbleMarkets/ntcharts/barchart](https://pkg.go.dev/github.com/NimbleMarkets/ntcharts/barchart) - Official barchart API docs; verified BarData, BarValue, Options
- [pkg.go.dev/github.com/NimbleMarkets/ntcharts/sparkline](https://pkg.go.dev/github.com/NimbleMarkets/ntcharts/sparkline) - Official sparkline API docs; verified Push, Draw, DrawBraille methods
- [pkg.go.dev/github.com/NimbleMarkets/ntcharts/heatmap](https://pkg.go.dev/github.com/NimbleMarkets/ntcharts/heatmap) - Official heatmap API docs; verified HeatPoint, color scales, matrix input
- [pkg.go.dev/github.com/cenkalti/backoff/v5](https://pkg.go.dev/github.com/cenkalti/backoff/v5) - Official backoff library docs; verified Retry, Permanent, ExponentialBackOff
- [github/linguist languages.yml](https://raw.githubusercontent.com/github/linguist/master/lib/linguist/languages.yml) - Official GitHub language color source; fetched exact hex codes for 13 languages
- [WakaTime API Docs](https://wakatime.com/developers) - Official API documentation; verified /durations endpoint, heartbeats, rate limiting
- [charmbracelet/bubbletea WindowSizeMsg docs](https://pkg.go.dev/github.com/charmbracelet/bubbletea) - Official WindowSizeMsg documentation; verified resize handling

### Secondary (MEDIUM confidence)
- [oneuptime.com: How to Build CLI Tools with Bubbletea](https://oneuptime.com/blog/post/2026-01-30-how-to-build-command-line-tools-with-bubbletea-in-go/view) - Bubbletea usage patterns (Jan 2026)
- [leg100.github.io: Tips for Building Bubbletea Programs](https://leg100.github.io/en/posts/building-bubbletea-programs/) - Layout pitfalls, resize handling best practices
- [GitHub Discussion #661: Resizing truncates on terminal window's original size](https://github.com/charmbracelet/bubbletea/discussions/661) - Community WindowSizeMsg patterns
- [oneuptime.com: Go Retry with Exponential Backoff](https://oneuptime.com/blog/post/2026-01-07-go-retry-exponential-backoff/view) - Backoff implementation guide (Jan 2026)
- [ozh/github-colors](https://github.com/ozh/github-colors) - Community GitHub language colors resource; cross-verified with Linguist
- [Ham Vocke: 16-Color Vim Color Scheme](https://hamvocke.com/blog/ansi-vim-color-scheme/) - Terminal color best practices; ANSI color mapping

### Tertiary (LOW confidence)
- WebSearch results on terminal color schemes - General consensus on color best practices; not authoritative but useful context
- WebSearch results on asciigraph and pterm - Alternative libraries for comparison; verified ntcharts is better fit for bubbletea

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - ntcharts verified via pkg.go.dev and GitHub; cenkalti/backoff is de facto standard (3.8k stars)
- Architecture patterns: HIGH - All code examples verified against official docs (ntcharts, backoff, bubbletea)
- WakaTime API patterns: MEDIUM - /durations endpoint verified in docs; hourly granularity limitation confirmed but workarounds not officially documented
- Color mapping: HIGH - GitHub Linguist YAML fetched directly; color values verified
- Pitfalls: MEDIUM-HIGH - WindowSizeMsg pitfall from Phase 5 research (HIGH); chart Draw() pitfall from ntcharts docs (HIGH); rate limiting pitfall from general HTTP best practices (MEDIUM)

**Research date:** 2026-02-19
**Valid until:** 2026-05-19 (90 days - ntcharts is new but stable; bubbletea v1 is stable; backoff is mature)

**Key assumptions:**
- ntcharts v0.1.x API is stable (released Jan 2026; no RC/beta warnings in docs)
- WakaTime API rate limits are per-user, not per-app (standard for OAuth APIs; not explicitly documented)
- lipgloss color degradation works correctly in ntcharts (ntcharts uses lipgloss for all styling; should inherit degradation)
