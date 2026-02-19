# Architecture Patterns for Dashboard Visual Overhaul + Themes

**Project:** wakadash v2.1
**Domain:** Terminal dashboard enhancement with stats panels and theme system
**Researched:** 2026-02-19
**Confidence:** HIGH

## Executive Summary

The architecture for adding comprehensive stats panels and a theme system to wakadash builds on the existing Bubble Tea Model-View-Update pattern with Lipgloss styling. The recommended approach uses a **centralized theme abstraction layer** with adaptive color support, **modular panel components** integrated into the existing view composition, and **config-based theme persistence** via ~/.wakatime.cfg extension.

Key architectural decisions:
- **Theme system**: Style registry pattern with adaptive colors, not runtime theme switching
- **Panel organization**: Grid-based layout with responsive column arrangement
- **Data aggregation**: Summary stats computed in Update, not View (maintain pure render)
- **Integration points**: Extend existing Model fields, add theme field to config

## Current Architecture Analysis

### Existing Structure (Bubble Tea + Lipgloss)

```
wakadash/
├── cmd/wakadash/
│   └── main.go              # Entry point, config loading, program initialization
├── internal/
│   ├── api/
│   │   └── client.go        # WakaTime API client
│   ├── config/
│   │   └── config.go        # ~/.wakatime.cfg parser (APIURL, APIKey)
│   ├── tui/
│   │   ├── model.go         # Bubble Tea Model (state + MUV methods)
│   │   ├── styles.go        # Lipgloss styles (5 global vars, inline colors)
│   │   ├── colors.go        # Language color mapping (GitHub Linguist colors)
│   │   ├── commands.go      # tea.Cmd factories for async ops
│   │   ├── messages.go      # Custom message types
│   │   └── keymap.go        # Keyboard bindings
│   └── types/
│       └── types.go         # API response structs
```

### Current Styling Approach

**styles.go** (5 global variables with hardcoded colors):
```go
borderStyle   = lipgloss.NewStyle().Border(...).BorderForeground("#62")
titleStyle    = lipgloss.NewStyle().Bold(true).Foreground("#205")
dimStyle      = lipgloss.NewStyle().Foreground("#241")
errorStyle    = lipgloss.NewStyle().Foreground("#196")
warningStyle  = lipgloss.NewStyle().Foreground("#214")
```

**colors.go** (language-specific colors):
```go
languageColors = map[string]string{
    "go": "#00ADD8",
    "javascript": "#f1e05a",
    // ... GitHub Linguist colors
}
```

**Inline styles in model.go**:
```go
// Spinner color
s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

// Bar chart colors
projectColor := lipgloss.Color("#00d7ff")
barStyle := lipgloss.NewStyle().Foreground(projectColor)
```

### Current Panel Rendering

**Layout strategy**: Vertical stacking with horizontal pairing
- Title + totals (top)
- Languages + Projects (2-column, lipgloss.JoinHorizontal)
- Sparkline (full-width)
- Heatmap (full-width)
- Status bar (bottom)

**Panel visibility**: Model fields control toggles (showLanguages, showProjects, etc.)

**Chart dimensions**: Calculated in WindowSizeMsg handler, stored as chart model state

## Recommended Architecture for v2.1

### Theme System Architecture

#### Pattern: Style Registry with Adaptive Colors

**Rationale**: Based on charmbracelet/huh theme.go and purpleclay/lipgloss-theme patterns:
- Lipgloss adaptive colors provide automatic light/dark terminal detection
- Theme struct centralizes all styling concerns
- Presets are static definitions, not runtime-switchable (avoids state complexity)

**Structure**:

```go
// internal/tui/theme.go

package tui

import "github.com/charmbracelet/lipgloss"

// Theme contains all styling for the dashboard.
type Theme struct {
    Name string

    // UI Elements
    Border      lipgloss.Style
    Title       lipgloss.Style
    Dim         lipgloss.Style
    Error       lipgloss.Style
    Warning     lipgloss.Style

    // Chart Colors (semantic)
    Primary     lipgloss.Color  // Main data bars
    Secondary   lipgloss.Color  // Background bars
    Accent      lipgloss.Color  // Highlights, spinners
    Success     lipgloss.Color  // Positive indicators

    // Heatmap gradient (5 levels)
    HeatmapLow  lipgloss.Color
    HeatmapMed1 lipgloss.Color
    HeatmapMed2 lipgloss.Color
    HeatmapHigh1 lipgloss.Color
    HeatmapHigh2 lipgloss.Color

    // Stats panel colors (for new panels)
    CategoryColor lipgloss.Color
    EditorColor   lipgloss.Color
    OSColor       lipgloss.Color
    MachineColor  lipgloss.Color
}

// Preset themes
var (
    ThemeDracula    = newDraculaTheme()
    ThemeNord       = newNordTheme()
    ThemeGruvbox    = newGruvboxTheme()
    ThemeMonokai    = newMonokaiTheme()
    ThemeSolarized  = newSolarizedTheme()
    ThemeTokyoNight = newTokyoNightTheme()
    ThemeDefault    = ThemeDracula  // Default to Dracula
)

// GetTheme returns the theme by name.
func GetTheme(name string) Theme {
    switch strings.ToLower(name) {
    case "dracula":
        return ThemeDracula
    case "nord":
        return ThemeNord
    case "gruvbox":
        return ThemeGruvbox
    case "monokai":
        return ThemeMonokai
    case "solarized":
        return ThemeSolarized
    case "tokyonight", "tokyo-night":
        return ThemeTokyoNight
    default:
        return ThemeDefault
    }
}
```

#### Theme Definition Pattern (Example: Dracula)

```go
func newDraculaTheme() Theme {
    // Dracula color palette
    background := lipgloss.Color("#282A36")
    foreground := lipgloss.Color("#F8F8F2")
    comment := lipgloss.Color("#6272A4")
    cyan := lipgloss.Color("#8BE9FD")
    green := lipgloss.Color("#50FA7B")
    orange := lipgloss.Color("#FFB86C")
    pink := lipgloss.Color("#FF79C6")
    purple := lipgloss.Color("#BD93F9")
    red := lipgloss.Color("#FF5555")
    yellow := lipgloss.Color("#F1FA8C")

    return Theme{
        Name: "Dracula",

        // UI Elements use adaptive colors for light terminal support
        Border: lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.AdaptiveColor{
                Light: string(comment),
                Dark:  string(purple),
            }),

        Title: lipgloss.NewStyle().
            Bold(true).
            Foreground(lipgloss.AdaptiveColor{
                Light: string(purple),
                Dark:  string(pink),
            }),

        Dim: lipgloss.NewStyle().
            Foreground(lipgloss.AdaptiveColor{
                Light: "#666666",
                Dark:  string(comment),
            }),

        Error: lipgloss.NewStyle().
            Foreground(red),

        Warning: lipgloss.NewStyle().
            Foreground(orange),

        // Chart colors (non-adaptive, work on dark backgrounds)
        Primary:   green,
        Secondary: comment,
        Accent:    pink,
        Success:   green,

        // Heatmap gradient (GitHub-style, green progression)
        HeatmapLow:   lipgloss.Color("#2d2d2d"),
        HeatmapMed1:  lipgloss.Color("#0e4429"),
        HeatmapMed2:  lipgloss.Color("#006d32"),
        HeatmapHigh1: lipgloss.Color("#26a641"),
        HeatmapHigh2: lipgloss.Color("#39d353"),

        // New panel colors
        CategoryColor: cyan,
        EditorColor:   yellow,
        OSColor:       purple,
        MachineColor:  orange,
    }
}
```

#### Other Theme Color Palettes

**Nord**:
```go
// Polar Night (backgrounds): nord0-3
nord0 := "#2e3440"
nord1 := "#3b4252"
nord2 := "#434c5e"
nord3 := "#4c566a"

// Snow Storm (foregrounds): nord4-6
nord4 := "#d8dee9"
nord5 := "#e5e9f0"
nord6 := "#eceff4"

// Frost (accents): nord7-10
nord7 := "#8fbcbb"   // cyan
nord8 := "#88c0d0"   // bright cyan
nord9 := "#81a1c1"   // blue
nord10 := "#5e81ac"  // dark blue

// Aurora (syntax): nord11-15
nord11 := "#bf616a"  // red
nord12 := "#d08770"  // orange
nord13 := "#ebcb8b"  // yellow
nord14 := "#a3be8c"  // green
nord15 := "#b48ead"  // purple
```

**Gruvbox Dark**:
```go
// Backgrounds
dark0 := "#282828"
dark1 := "#3c3836"
dark2 := "#504945"
dark3 := "#665c54"

// Foregrounds
light0 := "#fbf1c7"
light1 := "#ebdbb2"

// Accents
red := "#cc241d"
green := "#98971a"
yellow := "#d79921"
blue := "#458588"
purple := "#b16286"
aqua := "#689d6a"
orange := "#d65d0e"
```

**Tokyo Night**:
```go
// Common colors from folke/tokyonight.nvim
background := "#1a1b26"
foreground := "#c0caf5"
comment := "#565f89"

blue := "#7aa2f7"
cyan := "#7dcfff"
green := "#9ece6a"
magenta := "#bb9af7"
orange := "#ff9e64"
red := "#f7768e"
yellow := "#e0af68"
```

**Solarized Dark**:
```go
base03 := "#002b36"  // background
base02 := "#073642"  // background highlights
base01 := "#586e75"  // comments
base00 := "#657b83"
base0 := "#839496"   // body text
base1 := "#93a1a1"   // emphasis
base2 := "#eee8d5"
base3 := "#fdf6e3"

// Accents
yellow := "#b58900"
orange := "#cb4b16"
red := "#dc322f"
magenta := "#d33682"
violet := "#6c71c4"
blue := "#268bd2"
cyan := "#2aa198"
green := "#859900"
```

**Monokai**:
```go
background := "#272822"
foreground := "#f8f8f2"
comment := "#75715e"

red := "#f92672"
orange := "#fd971f"
yellow := "#e6db74"
green := "#a6e22e"
cyan := "#66d9ef"
purple := "#ae81ff"
```

#### Why Not Runtime Theme Switching?

**Avoided pattern**: Mutable theme state in Model with switch command

**Rationale**:
1. **Complexity**: Requires theme field in Model, switch key binding, theme mutation logic
2. **Lipgloss design**: Adaptive colors handle light/dark automatically, no switching needed
3. **Config-based simpler**: Set once at startup, no state management
4. **Use case mismatch**: Users rarely switch themes mid-session in terminals

**Chosen pattern**: Theme selected at startup via config/flag, immutable during session

### Theme Persistence Architecture

#### Extend ~/.wakatime.cfg

**Rationale**:
- Users already have ~/.wakatime.cfg (api_url, api_key)
- No new config file proliferation
- Simple key=value format

**Extended config structure**:
```
# ~/.wakatime.cfg
api_url = https://api.wakatime.com
api_key = waka_xxx

# wakadash settings (optional, new in v2.1)
theme = dracula
```

**Implementation**:

```go
// internal/config/config.go (modified)

type Config struct {
    APIURL string
    APIKey string
    Theme  string  // NEW: theme name, defaults to "dracula"
}

func Load() (*Config, error) {
    // ... existing api_url and api_key parsing ...

    // NEW: Parse theme (optional, defaults to dracula)
    switch key {
    case "api_url":
        cfg.APIURL = value
    case "api_key":
        cfg.APIKey = value
    case "theme":  // NEW
        cfg.Theme = value
    }

    // Default theme if not specified
    if cfg.Theme == "" {
        cfg.Theme = "dracula"
    }

    return cfg, nil
}
```

**Alternative: CLI flag override**:
```go
// cmd/wakadash/main.go
themeFlag := flag.String("theme", "", "Color theme (dracula, nord, gruvbox, monokai, solarized, tokyonight)")

// Config precedence: CLI flag > config file > default
themeName := cfg.Theme
if *themeFlag != "" {
    themeName = *themeFlag
}
```

### New Panels Architecture

#### Data Structure Extensions

**New Model fields** (add to existing Model struct):

```go
// internal/tui/model.go

type Model struct {
    // ... existing fields ...

    // NEW: Theme
    theme Theme

    // NEW: Chart components for new panels
    categoriesChart barchart.Model
    editorsChart    barchart.Model
    osChart         barchart.Model
    machinesChart   barchart.Model

    // NEW: Summary stats (computed from StatsResponse)
    summaryStats *SummaryStats

    // NEW: Panel visibility toggles
    showCategories bool  // 5 key
    showEditors    bool  // 6 key
    showOS         bool  // 7 key
    showMachines   bool  // 8 key
    showSummary    bool  // 9 key
}

// SummaryStats holds aggregated stats for the summary panel.
type SummaryStats struct {
    Last30Days      string  // e.g., "42h 15m"
    TotalTime       string  // Same as stats.HumanReadableTotal
    DailyAvg        string  // Same as stats.HumanReadableDailyAverage
    TopProject      string
    TopEditor       string
    TopCategory     string
    TopOS           string
    LanguageCount   int
    ProjectCount    int
    CategoryCount   int
    EditorCount     int
    OSCount         int
    MachineCount    int
}
```

#### Summary Stats Computation

**Where**: In Update, when statsFetchedMsg arrives
**Why**: Maintain View as pure render function

```go
// internal/tui/model.go (Update method)

case statsFetchedMsg:
    m.loading = false
    m.stats = msg.stats
    m.err = nil

    // NEW: Compute summary stats
    m.summaryStats = computeSummaryStats(msg.stats)

    // Update all charts (existing + new)
    m.updateLanguagesChart()
    m.updateProjectsChart()
    m.updateCategoriesChart()  // NEW
    m.updateEditorsChart()     // NEW
    m.updateOSChart()          // NEW
    m.updateMachinesChart()    // NEW

    return m, scheduleRefresh(m.refreshInterval)
```

```go
// Compute summary stats from StatsResponse.
func computeSummaryStats(stats *types.StatsResponse) *SummaryStats {
    data := stats.Data

    // For "Last 30 Days", need to fetch summaries endpoint separately
    // OR use existing stats.HumanReadableTotal if range is last_30_days

    return &SummaryStats{
        Last30Days:    data.HumanReadableTotal,  // Approximate if range matches
        TotalTime:     data.HumanReadableTotal,
        DailyAvg:      data.HumanReadableDailyAverage,
        TopProject:    topItemName(data.Projects),
        TopEditor:     topItemName(data.Editors),
        TopCategory:   topItemName(data.Categories),
        TopOS:         topItemName(data.OperatingSystems),
        LanguageCount: len(data.Languages),
        ProjectCount:  len(data.Projects),
        CategoryCount: len(data.Categories),
        EditorCount:   len(data.Editors),
        OSCount:       len(data.OperatingSystems),
        MachineCount:  len(data.Machines),
    }
}

func topItemName(items []types.StatItem) string {
    if len(items) == 0 {
        return "None"
    }
    return items[0].Name
}
```

#### Panel Layout Strategy

**Current layout** (2-column for Languages/Projects):
```
┌────────────────────────────────────────┐
│ Title + Totals (full-width)            │
│ Languages (left) │ Projects (right)    │
│ Sparkline (full-width)                 │
│ Heatmap (full-width)                   │
└────────────────────────────────────────┘
```

**New layout** (responsive grid):

**Wide terminal (≥120 cols)**: 3-column grid
```
┌──────────────────────────────────────────────────────────────┐
│ Summary Panel (full-width, 8 lines)                          │
├──────────────────────┬──────────────────────┬────────────────┤
│ Languages            │ Projects             │ Categories     │
│ (top 10, 8 lines)    │ (top 10, 8 lines)    │ (top 10)       │
├──────────────────────┼──────────────────────┼────────────────┤
│ Editors              │ Operating Systems    │ Machines       │
│ (top 10, 8 lines)    │ (top 10, 8 lines)    │ (top 10)       │
└──────────────────────┴──────────────────────┴────────────────┘
```

**Medium terminal (80-119 cols)**: 2-column grid
```
┌──────────────────────────────────────────┐
│ Summary Panel (full-width)               │
├────────────────────┬─────────────────────┤
│ Languages          │ Projects            │
├────────────────────┼─────────────────────┤
│ Categories         │ Editors             │
├────────────────────┼─────────────────────┤
│ Operating Systems  │ Machines            │
└────────────────────┴─────────────────────┘
```

**Narrow terminal (<80 cols)**: Single column
```
┌────────────────────┐
│ Summary Panel      │
├────────────────────┤
│ Languages          │
├────────────────────┤
│ Projects           │
├────────────────────┤
│ Categories         │
├────────────────────┤
│ Editors            │
├────────────────────┤
│ Operating Systems  │
├────────────────────┤
│ Machines           │
└────────────────────┘
```

**Implementation**:

```go
// internal/tui/model.go

func (m Model) renderDashboard() string {
    if m.width >= 120 {
        return m.renderThreeColumnLayout()
    } else if m.width >= 80 {
        return m.renderTwoColumnLayout()
    }
    return m.renderSingleColumnLayout()
}

func (m Model) renderThreeColumnLayout() string {
    // Summary panel (full-width)
    summaryPanel := m.renderSummaryPanel()

    // Build 3-column grid
    col1 := m.renderPanelColumn([]string{"languages", "editors"})
    col2 := m.renderPanelColumn([]string{"projects", "os"})
    col3 := m.renderPanelColumn([]string{"categories", "machines"})

    grid := lipgloss.JoinHorizontal(lipgloss.Top, col1, col2, col3)

    return lipgloss.JoinVertical(lipgloss.Left, summaryPanel, grid)
}
```

#### Panel Rendering Pattern

**Reference**: wakafetch ui/render.go CardSection pattern

**Unified panel renderer**:

```go
// internal/tui/model.go

// renderStatsPanel renders a single stats category as a panel with title and bars.
func (m Model) renderStatsPanel(title string, chart barchart.Model, visible bool) string {
    if !visible {
        return ""
    }

    titleLine := m.theme.Title.Render(title)
    chartContent := chart.View()

    return lipgloss.JoinVertical(lipgloss.Left, titleLine, chartContent)
}

// Update chart with theme colors.
func (m *Model) updateCategoriesChart() {
    if m.stats == nil {
        return
    }

    m.categoriesChart.Clear()
    data := m.stats.Data

    limit := min(10, len(data.Categories))
    barStyle := lipgloss.NewStyle().Foreground(m.theme.CategoryColor)

    for _, cat := range data.Categories[:limit] {
        hours := cat.TotalSeconds / 3600.0
        m.categoriesChart.Push(barchart.BarData{
            Label: cat.Name,
            Values: []barchart.BarValue{{
                Name:  "",
                Value: hours,
                Style: barStyle,
            }},
        })
    }

    m.categoriesChart.Draw()
}
```

#### Summary Panel Rendering

**Layout**: 2-column key-value display, similar to wakafetch stats panel

```go
func (m Model) renderSummaryPanel() string {
    if m.summaryStats == nil {
        return ""
    }

    s := m.summaryStats

    // Left column
    leftFields := []string{
        fmt.Sprintf("Last 30 Days:  %s", s.Last30Days),
        fmt.Sprintf("Total Time:    %s", s.TotalTime),
        fmt.Sprintf("Daily Avg:     %s", s.DailyAvg),
        fmt.Sprintf("Top Project:   %s", s.TopProject),
    }

    // Right column
    rightFields := []string{
        fmt.Sprintf("Top Editor:    %s", s.TopEditor),
        fmt.Sprintf("Top Category:  %s", s.TopCategory),
        fmt.Sprintf("Top OS:        %s", s.TopOS),
        fmt.Sprintf("Languages: %d  Projects: %d  Categories: %d",
            s.LanguageCount, s.ProjectCount, s.CategoryCount),
    }

    // Style fields
    leftPanel := m.theme.Dim.Render(strings.Join(leftFields, "\n"))
    rightPanel := m.theme.Dim.Render(strings.Join(rightFields, "\n"))

    content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)

    title := m.theme.Title.Render("Summary")
    return lipgloss.JoinVertical(lipgloss.Left, title, content)
}
```

### Integration Points Summary

#### Files to Modify

| File | Changes | Reason |
|------|---------|--------|
| `internal/config/config.go` | Add `Theme string` field, parse `theme=` key | Theme persistence |
| `cmd/wakadash/main.go` | Add `-theme` flag, pass theme to Model | CLI override |
| `internal/tui/model.go` | Add theme, new charts, summaryStats, show* flags | New state |
| `internal/tui/styles.go` | **DELETE**, move to theme.go | Centralize styling |
| `internal/tui/colors.go` | Keep, but reference theme colors | Language colors |

#### Files to Create

| File | Purpose |
|------|---------|
| `internal/tui/theme.go` | Theme struct, presets, GetTheme function |
| `internal/tui/themes_dracula.go` | Dracula theme definition |
| `internal/tui/themes_nord.go` | Nord theme definition |
| `internal/tui/themes_gruvbox.go` | Gruvbox theme definition |
| `internal/tui/themes_monokai.go` | Monokai theme definition |
| `internal/tui/themes_solarized.go` | Solarized theme definition |
| `internal/tui/themes_tokyonight.go` | Tokyo Night theme definition |

#### Data Flow Changes

**Before** (v2.0):
```
main.go → config.Load() → (APIURL, APIKey)
       → api.New(key, url)
       → tui.NewModel(client, range, refresh)
       → Model with hardcoded styles
```

**After** (v2.1):
```
main.go → config.Load() → (APIURL, APIKey, Theme)
       → api.New(key, url)
       → tui.GetTheme(cfg.Theme) → Theme struct
       → tui.NewModel(client, range, refresh, theme)
       → Model with theme-based styles
```

**NewModel signature change**:
```go
// Before
func NewModel(client *api.Client, rangeStr string, refreshInterval time.Duration) Model

// After
func NewModel(client *api.Client, rangeStr string, refreshInterval time.Duration, theme Theme) Model
```

## Component Boundaries

| Component | Responsibility | Dependencies | Modified/New |
|-----------|---------------|--------------|--------------|
| `theme.go` | Theme definitions, GetTheme | lipgloss | NEW |
| `themes_*.go` | Individual theme constructors | lipgloss, theme.go | NEW |
| `config.go` | Parse theme from config | os, strings | MODIFIED |
| `model.go` | Store theme, new charts, summary stats | theme.go, barchart | MODIFIED |
| `colors.go` | Language color mapping | theme (for fallback) | MODIFIED |
| `styles.go` | Global style vars | - | DELETE |

## Build Order and Dependencies

### Phase 1: Theme Foundation (No UI Changes)
1. Create `theme.go` with Theme struct and GetTheme
2. Create `themes_dracula.go` (single theme for testing)
3. Modify `config.go` to parse theme field
4. Modify `main.go` to pass theme to NewModel
5. Modify `model.go` NewModel signature to accept theme
6. **Test**: Verify theme loads, no visual changes yet

### Phase 2: Convert Existing Styles to Theme
1. Replace `borderStyle` with `m.theme.Border` in renderDashboard
2. Replace `titleStyle` with `m.theme.Title` throughout
3. Replace `dimStyle`, `errorStyle`, `warningStyle` with theme equivalents
4. Update chart colors to use theme.Primary/Secondary/Accent
5. Delete `styles.go`
6. **Test**: Existing UI renders with Dracula theme

### Phase 3: Add Remaining Theme Presets
1. Create `themes_nord.go`
2. Create `themes_gruvbox.go`
3. Create `themes_monokai.go`
4. Create `themes_solarized.go`
5. Create `themes_tokyonight.go`
6. **Test**: Switch theme via config, verify colors change

### Phase 4: Add New Panel Components
1. Add new barchart fields to Model (categoriesChart, editorsChart, etc.)
2. Add chart initialization in NewModel
3. Add update* methods for new charts
4. **Test**: Charts initialize, no rendering yet

### Phase 5: Add Summary Stats
1. Add SummaryStats struct to model.go
2. Add computeSummaryStats function
3. Compute in statsFetchedMsg handler
4. **Test**: Stats computed correctly, logged

### Phase 6: Render New Panels
1. Implement renderSummaryPanel
2. Implement responsive layout renderers (1/2/3 column)
3. Add new panel toggles (keys 5-9)
4. Update help text
5. **Test**: All panels render, toggles work

### Dependency Graph

```
theme.go
  ├─ themes_dracula.go
  ├─ themes_nord.go
  ├─ themes_gruvbox.go
  ├─ themes_monokai.go
  ├─ themes_solarized.go
  └─ themes_tokyonight.go

config.go (modified) → theme name string

main.go (modified)
  ├─ config.go
  └─ theme.go (GetTheme)

model.go (modified)
  ├─ theme.go (Theme struct)
  ├─ colors.go (language colors, unchanged)
  └─ barchart (ntcharts, existing)

colors.go (modified)
  └─ theme.go (optional fallback to theme.Primary)
```

## Architectural Patterns to Follow

### Pattern 1: Centralized Theme Registry

**What**: All styles accessed via `m.theme.Field`
**When**: Every lipgloss.NewStyle() call in rendering
**Example**:
```go
// Before (scattered styles)
titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

// After (centralized theme)
title := m.theme.Title.Render("Languages")
```

### Pattern 2: Pure View Functions

**What**: View never computes data, only renders from Model
**When**: All rendering methods
**Example**:
```go
// WRONG: Computation in View
func (m Model) renderSummaryPanel() string {
    topProject := findTopProject(m.stats.Data.Projects)  // ❌ computation
    return fmt.Sprintf("Top Project: %s", topProject)
}

// RIGHT: Render from pre-computed state
func (m Model) renderSummaryPanel() string {
    return fmt.Sprintf("Top Project: %s", m.summaryStats.TopProject)  // ✓ just render
}
```

### Pattern 3: Responsive Layout Helpers

**What**: Layout adapts to terminal width via helper methods
**When**: WindowSizeMsg and renderDashboard
**Example**:
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height

        // Resize all charts based on layout
        panelWidth := m.calculatePanelWidth()
        m.categoriesChart.Resize(panelWidth, 8)
        // ...
}

func (m Model) calculatePanelWidth() int {
    if m.width >= 120 {
        return (m.width - 8) / 3  // 3-column grid
    } else if m.width >= 80 {
        return (m.width - 6) / 2  // 2-column grid
    }
    return m.width - 4  // single column
}
```

### Pattern 4: Adaptive Colors for UI, Fixed for Charts

**What**: UI elements (borders, titles) use adaptive colors, charts use fixed colors
**When**: Theme definition
**Why**: Charts need consistent visual identity, UI adapts to terminal

**Example**:
```go
Border: lipgloss.NewStyle().
    BorderForeground(lipgloss.AdaptiveColor{
        Light: "#6272A4",  // Dracula comment (lighter)
        Dark:  "#BD93F9",  // Dracula purple (brighter)
    }),

// Charts: fixed colors (assume dark terminal)
Primary: lipgloss.Color("#50FA7B"),  // Dracula green, no adaptation
```

## Anti-Patterns to Avoid

### Anti-Pattern 1: Runtime Theme Mutation

**What goes wrong**: Adding theme switcher key binding
**Why it happens**: Seems like a nice feature
**Consequences**:
- Requires theme state in Model
- All styles must be recomputed on switch
- Breaks Lipgloss immutable style pattern
- Adds complexity for rare use case

**Prevention**: Theme selection at startup only (config/flag)

### Anti-Pattern 2: Inline Style Creation

**What goes wrong**: Creating lipgloss.NewStyle() directly in View
**Why it happens**: Quick and easy for one-off styles
**Consequences**:
- Styles not themeable
- Scattered styling logic
- Performance (styles recreated every render)

**Prevention**: All styles via theme struct, computed once

**Example**:
```go
// ❌ WRONG: Inline style in View
func (m Model) renderTitle() string {
    style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6"))
    return style.Render("WakaTime Stats")
}

// ✓ RIGHT: Use theme
func (m Model) renderTitle() string {
    return m.theme.Title.Render("WakaTime Stats")
}
```

### Anti-Pattern 3: Hard-Coding Panel Dimensions

**What goes wrong**: Fixed panel widths/heights in View
**Why it happens**: Easier than calculating responsive dimensions
**Consequences**:
- Breaks on terminal resize
- Overflow or gaps in layout
- Poor UX on varied terminal sizes

**Prevention**: Calculate dimensions in WindowSizeMsg handler, use lipgloss.Width/Height helpers

**Example**:
```go
// ❌ WRONG: Hard-coded width
panel := lipgloss.NewStyle().Width(40).Render(content)

// ✓ RIGHT: Responsive width
panelWidth := m.calculatePanelWidth()
panel := lipgloss.NewStyle().Width(panelWidth).Render(content)
```

### Anti-Pattern 4: Duplicating Data Structures

**What goes wrong**: Storing same data in multiple Model fields
**Why it happens**: Convenience for different views
**Consequences**:
- Synchronization bugs
- Wasted memory
- Confusing state management

**Prevention**: Single source of truth (StatsResponse), derive views in Update

**Example**:
```go
// ❌ WRONG: Duplicate storage
type Model struct {
    stats          *types.StatsResponse
    topProject     string  // ❌ duplicate
    topEditor      string  // ❌ duplicate
    summaryStats   *SummaryStats  // ✓ OK, derived once
}

// ✓ RIGHT: Compute in Update, store derivation
case statsFetchedMsg:
    m.stats = msg.stats
    m.summaryStats = computeSummaryStats(msg.stats)  // Single derivation
```

## Scalability Considerations

### At Current Scale (Single User, ~10 Stats Categories)

**Approach**: Direct struct storage, synchronous rendering
- All 6 panels fit in memory easily
- Rendering is instant (<10ms)
- No optimization needed

### If Scaling to More Panels (20+ Stats Categories)

**Approach**: Lazy rendering, pagination
- Only render visible panels
- Add panel pagination (PgUp/PgDn to scroll)
- Keep data fetch unchanged (API provides all)

**Example**:
```go
type Model struct {
    allPanels     []PanelConfig
    visibleStart  int  // Index of first visible panel
    visibleEnd    int  // Index of last visible panel
}

func (m Model) renderDashboard() string {
    visiblePanels := m.allPanels[m.visibleStart:m.visibleEnd]
    // Render only visible panels
}
```

### If Scaling to Historical Data (Charts with Time Series)

**Approach**: Add time range navigation, cache API responses
- Store historical data in Model (map[string]*StatsResponse)
- Add date range picker
- Cache to avoid re-fetching same range

**Not needed for v2.1**: Current scope is single time range snapshot

## Research Confidence Assessment

| Area | Confidence | Evidence |
|------|------------|----------|
| Theme struct pattern | HIGH | charmbracelet/huh theme.go, purpleclay/lipgloss-theme |
| Adaptive colors | HIGH | Official Lipgloss docs, multiple sources |
| Config extension | HIGH | Existing ~/.wakatime.cfg parser in codebase |
| Panel layout strategy | MEDIUM | Extrapolated from wakafetch CardSection, Bubble Tea examples |
| Theme color palettes | HIGH | Official Dracula spec, Nord docs, Gruvbox contrib |
| Responsive grid | MEDIUM | Based on Bubble Tea best practices, not project-specific |

## Open Questions for Implementation

1. **Heatmap gradient**: Should each theme define custom heatmap colors, or use a universal GitHub-style green gradient?
   - **Recommendation**: Theme-specific for brand consistency (Dracula = purple gradient, Nord = blue gradient)

2. **Summary stats "Last 30 Days"**: Should we fetch summaries endpoint separately, or approximate from current range?
   - **Recommendation**: If range is last_30_days, use TotalTime directly. Otherwise, show "Range: [actual range]" instead of "Last 30 Days"

3. **Panel toggle persistence**: Should panel visibility be saved to config, or reset each session?
   - **Recommendation**: Reset each session (simpler, avoids config pollution)

4. **Chart bar limit**: wakafetch shows top 10. Should we stick with 10 or make configurable?
   - **Recommendation**: Fixed at 10 (matches requirements, avoids config creep)

5. **Language colors vs theme**: Should language bars use GitHub Linguist colors or theme colors?
   - **Recommendation**: Keep GitHub Linguist colors (semantic, recognizable). Other panels use theme colors.

## Sources

- [Lipgloss GitHub Repository](https://github.com/charmbracelet/lipgloss)
- [purpleclay lipgloss-theme Package](https://pkg.go.dev/github.com/purpleclay/lipgloss-theme)
- [Bubble Tea GitHub Repository](https://github.com/charmbracelet/bubbletea)
- [Shifoo: Multi View Interfaces in Bubble Tea](https://shi.foo/weblog/multi-view-interfaces-in-bubble-tea)
- [Tips for building Bubble Tea programs](https://leg100.github.io/en/posts/building-bubbletea-programs/)
- [charmbracelet/huh theme.go](https://github.com/charmbracelet/huh/blob/main/theme.go)
- [Dracula Theme Specification](https://draculatheme.com/spec)
- [Nord Theme Colors and Palettes](https://www.nordtheme.com/docs/colors-and-palettes/)
- [Gruvbox GitHub Repository](https://github.com/morhetz/gruvbox)
- [Gruvbox Color Guide](https://github.com/vanzsh/gruvbox-color-guide)
- [Tokyo Night GitHub Repository](https://github.com/folke/tokyonight.nvim)
- [Gogh Terminal Color Schemes](https://gogh-co.github.io/Gogh/)
