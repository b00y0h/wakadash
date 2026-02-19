# Technology Stack

**Project:** wakadash v2.1 — Visual Overhaul + Themes
**Researched:** 2026-02-19
**Confidence:** HIGH

---

## Context: What Already Exists (v2.0)

wakadash v2.0 established the live dashboard foundation with:
- **charmbracelet/bubbletea** v1.3.10 — Live-refresh TUI framework
- **charmbracelet/lipgloss** v1.1.0 — Terminal styling and layout
- **charmbracelet/bubbles** v1.0.0 — Spinner, help components
- **NimbleMarkets/ntcharts** v0.4.0 — Sparklines, bar charts, heatmaps
- **cenkalti/backoff/v5** v5.0.3 — Retry logic for API calls

**Current capabilities:**
- Languages panel (top 5, horizontal bars, GitHub Linguist colors)
- Projects panel (top 5, horizontal bars, cyan color)
- Hourly sparkline (24-hour activity)
- 7-day heatmap (GitHub contribution-style colors)
- Hardcoded purple/magenta color scheme

**API data already available but not displayed:**
- Categories, Editors, OperatingSystems, Machines (in `StatsData` struct)
- Summary statistics (cumulative total, daily average, date ranges)

---

## Recommended Stack Additions for v2.1

### Theme System

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| github.com/willyv3/gogh-themes/lipgloss | v1.2.0 | Pre-built theme presets | Provides 361 professional terminal color schemes including all required themes (Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night). Zero dependencies, themes compiled into binary. Colors pre-wrapped as `lipgloss.Color` — no manual conversion needed. Released Oct 2025. |

### Configuration (No New Dependencies)

| Component | Implementation | Why |
|-----------|----------------|-----|
| Theme selection | Standard library `flag` package | Already used for `--range` and `--refresh` flags. Add `--theme` string flag. Simple, zero dependencies. |
| Theme fallback | Existing hardcoded styles | If theme flag not provided or invalid, fall back to current purple/magenta scheme. No breaking changes. |

---

## Full Dependency List for v2.1

**Add:**
```bash
go get github.com/willyv3/gogh-themes/lipgloss@v1.2.0
```

**No updates needed:**
- charmbracelet/bubbletea v1.3.10 ✓ (v2 still beta)
- charmbracelet/lipgloss v1.1.0 ✓ (v2 still alpha)
- charmbracelet/bubbles v1.0.0 ✓ (v2 still beta)
- NimbleMarkets/ntcharts v0.4.0 ✓ (latest stable)

**Stats panels use existing stack:**
- Categories, Editors, OS, Machines → existing `barchart.Model` (ntcharts)
- Summary stats → existing lipgloss styling
- No new chart types required

---

## What NOT to Add

| Anti-Dependency | Why Avoid | Use Instead |
|-----------------|-----------|-------------|
| spf13/viper | Overkill for single string flag. Adds 30+ transitive dependencies for config file parsing we don't need. | Standard library `flag` (already in use) |
| gookit/config | Heavy config library (YAML, TOML, JSON parsing). We only need theme name selection, not multi-format config files. | Standard library `flag` |
| knadh/koanf | Another heavyweight config solution. Theme selection doesn't justify the complexity. | Standard library `flag` |
| Custom theme parser | Manually defining 6+ color schemes is error-prone and duplicates work. | gogh-themes/lipgloss (themes pre-built) |
| charmbracelet/huh | Form/input library for interactive prompts. Dashboard is display-only with keyboard shortcuts, no forms needed. | Existing bubbletea keyboard handling |
| lipgloss v2 | Still in alpha (v2.0.0-alpha.2). Breaking API changes, unstable. v1.1.0 meets all needs. | lipgloss v1.1.0 |
| bubbletea v2 | Still in beta (v2.0.0-beta.1). Breaking changes to Init/Update signatures. v1.3.10 stable and sufficient. | bubbletea v1.3.10 |
| bubbles v2 | Beta status. v1.0.0 provides all needed components (spinner, help). | bubbles v1.0.0 |
| Additional chart libraries | ntcharts already handles horizontal bars. Same API works for all stat types. | NimbleMarkets/ntcharts v0.4.0 |

---

## Theme Integration Architecture

### Available Themes (Confirmed in gogh-themes/lipgloss v1.2.0)

All 6 required themes verified:
- **Dracula** — Purple/pink/cyan dark theme by Zeno Rocha
- **Nord** — Arctic blue-tinted dark theme
- **Gruvbox** (Gruvbox Dark) — Retro warm dark theme by Pavel Pertsev
- **Monokai** (Monokai Pro) — Classic dark theme by Wimer Hazenberg
- **Solarized** (Solarized Dark, Solarized Light) — Scientific color theory by Ethan Schoonover
- **Tokyo Night** (Tokyo Night, Tokyo Night Storm, Tokyo Night Light) — Modern clean theme

Plus 355 additional themes for future expansion.

### Theme Structure

```go
type Theme struct {
    Name       string
    Background lipgloss.Color
    Foreground lipgloss.Color

    // Primary colors (ANSI 0-7)
    Black, Red, Green, Yellow, Blue, Magenta, Cyan, White lipgloss.Color

    // Bright colors (ANSI 8-15)
    BrightBlack, BrightRed, BrightGreen, BrightYellow,
    BrightBlue, BrightMagenta, BrightCyan, BrightWhite lipgloss.Color
}
```

### Integration Flow

```
1. CLI flag parsing:
   --theme "Dracula" (optional, defaults to current hardcoded scheme)

2. Theme loading:
   import lipglossthemes "github.com/willyv3/gogh-themes/lipgloss"

   theme, ok := lipglossthemes.Get("Dracula")
   if !ok {
       // Fallback to default theme
       theme = getDefaultTheme()
   }

3. Style initialization (replace current hardcoded colors):
   borderStyle := lipgloss.NewStyle().
       Border(lipgloss.RoundedBorder()).
       BorderForeground(theme.Magenta)  // was lipgloss.Color("62")

   titleStyle := lipgloss.NewStyle().
       Bold(true).
       Foreground(theme.Blue)           // was lipgloss.Color("205")

   errorStyle := lipgloss.NewStyle().
       Foreground(theme.Red)            // was lipgloss.Color("196")

4. Pass theme to TUI model:
   m := tui.NewModel(client, rangeStr, refreshInterval, theme)

5. Apply theme colors consistently across all panels
```

### Usage Pattern

```go
// Before (v2.0 — hardcoded)
borderStyle = lipgloss.NewStyle().
    BorderForeground(lipgloss.Color("62"))  // hardcoded purple

// After (v2.1 — themeable)
theme, _ := lipglossthemes.Get(themeName)
borderStyle = lipgloss.NewStyle().
    BorderForeground(theme.Magenta)  // theme-aware
```

---

## Stats Panels Implementation (No New Dependencies)

### Data Already Available

All additional stats exist in `types.StatsData`:
```go
type StatsData struct {
    Categories       []StatItem  // ✓ Already fetched
    Editors          []StatItem  // ✓ Already fetched
    OperatingSystems []StatItem  // ✓ Already fetched
    Machines         []StatItem  // ✓ Already fetched
    Languages        []StatItem  // ✓ Currently displayed
    Projects         []StatItem  // ✓ Currently displayed

    // Summary stats
    HumanReadableTotal        string  // ✓ Currently displayed
    HumanReadableDailyAverage string  // ✓ Currently displayed
    DaysIncludingHolidays     int     // ✓ Available
    Range                     string  // ✓ Available
}
```

### Rendering Pattern (Reuse Existing Code)

```go
// Pattern established in v2.0 for Languages panel
func (m *Model) updateLanguagesChart() {
    m.languagesChart.Clear()
    for _, lang := range m.stats.Data.Languages[:5] {
        hours := lang.TotalSeconds / 3600.0
        color := getLanguageColor(lang.Name)
        barStyle := lipgloss.NewStyle().Foreground(color)
        m.languagesChart.Push(barchart.BarData{
            Label: lang.Name,
            Values: []barchart.BarValue{{
                Name:  "",
                Value: hours,
                Style: barStyle,
            }},
        })
    }
    m.languagesChart.Draw()
}

// Same pattern works for Categories, Editors, OS, Machines
func (m *Model) updateCategoriesChart() {
    m.categoriesChart.Clear()
    for _, cat := range m.stats.Data.Categories[:5] {
        hours := cat.TotalSeconds / 3600.0
        barStyle := lipgloss.NewStyle().Foreground(theme.Green)
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

**No new libraries needed** — `barchart.Model` from ntcharts v0.4.0 handles all stat types.

### Summary Panel (Lipgloss Layout)

```go
func (m Model) renderSummary() string {
    summaryStyle := lipgloss.NewStyle().
        Foreground(theme.Foreground).
        Padding(1)

    return summaryStyle.Render(fmt.Sprintf(
        "Last 30 Days: %s | Daily Avg: %s | Days: %d",
        m.stats.Data.HumanReadableTotal,
        m.stats.Data.HumanReadableDailyAverage,
        m.stats.Data.DaysIncludingHolidays,
    ))
}
```

**Uses existing lipgloss v1.1.0** — no additional layout libraries.

---

## Configuration Pattern

### CLI Flag Definition (main.go)

```go
func main() {
    themeFlag := flag.String("theme", "",
        "Color theme (dracula, nord, gruvbox, monokai, solarized, tokyo-night)")
    // ... existing flags
    flag.Parse()

    theme := loadTheme(*themeFlag)  // handles "" default case
    m := tui.NewModel(client, *rangeFlag, refreshInterval, theme)
    // ...
}
```

### Theme Loading Helper

```go
func loadTheme(name string) lipglossthemes.Theme {
    if name == "" {
        return getDefaultTheme()  // current purple/magenta scheme
    }

    theme, ok := lipglossthemes.Get(name)
    if !ok {
        fmt.Fprintf(os.Stderr, "Warning: theme '%s' not found, using default\n", name)
        return getDefaultTheme()
    }
    return theme
}

func getDefaultTheme() lipglossthemes.Theme {
    // Recreate current hardcoded colors as a Theme struct
    return lipglossthemes.Theme{
        Name:       "Default",
        Magenta:    lipgloss.Color("205"),  // current titleStyle
        Blue:       lipgloss.Color("62"),   // current borderStyle
        Red:        lipgloss.Color("196"),  // current errorStyle
        Yellow:     lipgloss.Color("214"),  // current warningStyle
        Foreground: lipgloss.Color("252"),
        // ...
    }
}
```

**No config file needed** — simple flag-based selection.

---

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| Theme library | gogh-themes/lipgloss v1.2.0 | Manual theme definitions | Manual: error-prone, 6+ themes × 16 colors each = 96+ color values to maintain. gogh-themes: battle-tested, zero-dependency. |
| Theme library | gogh-themes/lipgloss v1.2.0 | charmbracelet/huh theme system | huh themes tied to form components, not general-purpose. Dashboard has no forms. |
| Config | Standard library flag | spf13/viper | Viper adds 30+ dependencies for features we don't use (YAML, TOML, remote config, env var binding). Simple string flag sufficient. |
| Config | Standard library flag | Config file (~/.wakadash.yaml) | Over-engineering. Theme name is the only new config. Flag pattern established. |
| Stats rendering | Existing ntcharts barchart | New table library | Bar charts more visual than tables. ntcharts already renders horizontal bars successfully. |
| Stats rendering | Existing ntcharts barchart | lipgloss table package | Table = text-heavy. Dashboard prioritizes visual data (charts). lipgloss table better for logs/data grids. |

---

## Migration Path from v2.0 to v2.1

### Phase 1: Add Theme Support (Non-Breaking)
1. Add `gogh-themes/lipgloss` dependency
2. Add `--theme` flag (optional, defaults to current behavior)
3. Refactor `internal/tui/styles.go` to accept theme parameter
4. Pass theme through `NewModel()` and apply to all panels
5. **Backward compatible** — no flag = current hardcoded colors

### Phase 2: Add Stats Panels (Pure Addition)
1. Add 4 new `barchart.Model` fields to `Model` struct (categories, editors, os, machines)
2. Implement `updateCategoriesChart()`, `updateEditorsChart()`, etc. (clone existing pattern)
3. Add panels to `renderStats()` layout
4. Add keyboard toggles (5-8 keys) for new panels
5. **No breaking changes** — additive only

### Phase 3: Add Summary Panel (Pure Addition)
1. Implement `renderSummary()` using existing `StatsData` fields
2. Add to dashboard layout above or below existing panels
3. **No API changes** — data already fetched

---

## Sources

**HIGH CONFIDENCE (Official Documentation):**
- [gogh-themes/lipgloss pkg.go.dev](https://pkg.go.dev/github.com/willyv3/gogh-themes/lipgloss) — v1.2.0, Oct 2025, API confirmed
- [gogh-themes GitHub](https://github.com/willyv3/gogh-themes) — 361 themes, Dracula/Nord/Gruvbox/Monokai/Solarized/Tokyo Night verified
- [Gogh Color Schemes](https://github.com/Gogh-Co/Gogh) — Source of themes, maintained by community
- [NimbleMarkets/ntcharts](https://github.com/NimbleMarkets/ntcharts) — v0.4.0, Jan 2026, barchart API stable
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) — v1.1.0, Mar 2025
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) — v1.3.10, Sep 2025

**MEDIUM CONFIDENCE (Multi-Source Verified):**
- [Go flag package documentation](https://pkg.go.dev/flag) — Standard library, stable API
- [spf13/viper alternatives discussion](https://blog.logrocket.com/handling-go-configuration-viper/) — Confirms viper is overkill for simple flags
- [Bubbletea v2 beta status](https://github.com/charmbracelet/bubbletea/discussions/1237) — v2.0.0-beta.1, breaking changes confirmed

**LOW CONFIDENCE (Needs Validation):**
- None — all stack decisions backed by official docs and stable releases.
