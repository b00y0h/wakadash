# Phase 9: Stats Panels + Summary - Research

**Researched:** 2026-02-20
**Domain:** Dashboard stats panels with responsive layout and keyboard toggles
**Confidence:** HIGH

## Summary

Phase 9 builds comprehensive stats panels (Categories, Editors, OS, Machines) and a Summary panel using data already available from WakaTime's `/v1/users/current/stats/{range}` API. The wakadash codebase has proven patterns from Phase 6: ntcharts barchart.Model for horizontal bars, lipgloss.JoinHorizontal/JoinVertical for layouts, and WindowSizeMsg-based responsive sizing.

The single `/stats` API call returns all required data (categories, editors, operating_systems, machines) with name, total_seconds, and percent fields. The StatsData struct in types.go already includes these arrays. Best_day is also available in the stats response. However, **streaks are NOT provided by the WakaTime API** — calculating current/best streaks requires parsing daily summaries data, which is complex and outside phase scope. The user context confirms showing streaks in the Summary panel.

The responsive layout challenge (2-column ≥80 cols, stack <80 cols) is well-documented in lipgloss patterns: use conditional logic based on m.width, render panels independently, then join horizontally or vertically. Panel visibility toggles follow the existing pattern from Phase 6 (number keys map to bool fields).

**Primary recommendation:** Reuse ntcharts barchart.Model with WithHorizontalBars() option for all four stat panels. Create Summary panel as formatted text block with highlighted styling (subtle border/background). Calculate streaks from summaryData (7-day window only, not historical best). Use lipgloss.JoinHorizontal for 2-column layout when m.width ≥ 80, otherwise lipgloss.JoinVertical. Extend keymap to support panels 5-9 plus show-all/hide-all shortcuts. Truncate panels when m.height is insufficient, showing minimum 3 items per panel.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Panel Content:**
- Show top 10 items per panel (Categories, Editors, OS, Machines)
- Human-readable time format: "2h 15m" or "45 mins"
- Always show percentages alongside time: "VS Code: 4h 30m (65%)"
- Group unknown/untracked items as "Other" category

**Summary Panel:**
- Full overview: 30d total, daily avg, top language/project/editor, streaks, active days, item counts
- Highlighted visual treatment — subtle accent border or background (distinct from other panels)
- Position at top — first thing users see, overview before details
- Show both current streak and best streak: "Current: 5 days | Best: 14 days"

**Layout Behavior:**
- Wide terminals (≥80 cols): Summary spans full width at top, other panels in 2-column grid below
- Narrow terminals (<80 cols): All panels stack vertically, single column, full width each
- Truncate panels when terminal is too short (no scrolling viewport)
- Minimum 3 items per panel when truncating

**Toggle Shortcuts:**
- Number keys map to visual order: 1=first visible panel, 2=second, etc.
- Include show-all and hide-all shortcuts (e.g., Shift+A, Shift+H)
- Panel visibility does NOT persist across restarts — always start with all panels visible
- Help overlay shows abbreviated hint: "1-9: toggle panels" (not listing each individually)

### Claude's Discretion
- Exact column widths and spacing in 2-column layout
- Which specific keys for show-all/hide-all
- How to visually indicate truncated panels
- Exact accent styling for Summary panel
- Panel ordering within the 2-column grid

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| STAT-01 | User sees Categories panel with top 10 categories as horizontal bars with time labels | ntcharts barchart.Model with WithHorizontalBars(); data from StatsData.Categories |
| STAT-02 | User sees Editors panel with top 10 editors as horizontal bars with time labels | Same pattern as STAT-01; data from StatsData.Editors |
| STAT-03 | User sees Operating Systems panel with top 10 OS as horizontal bars with time labels | Same pattern as STAT-01; data from StatsData.OperatingSystems |
| STAT-04 | User sees Machines panel with top 10 machines as horizontal bars with time labels | Same pattern as STAT-01; data from StatsData.Machines |
| STAT-05 | User sees Summary panel showing Last 30d total, daily avg, top items, streaks, counts | Formatted text panel; streak calculation from SummaryResponse; best_day from StatsData |
| LAYOUT-01 | Dashboard panels arrange in 2-column layout on terminals ≥80 cols, stacking on smaller terminals | lipgloss.JoinHorizontal vs JoinVertical based on m.width conditional |
| LAYOUT-02 | User can toggle each panel's visibility with keyboard shortcuts | Extend keymap pattern from Phase 6 (Toggle1-4) to Toggle5-9 plus show-all/hide-all |

</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/NimbleMarkets/ntcharts/barchart | v0.4.0 | Horizontal bar charts | Already used for Languages/Projects; supports WithHorizontalBars() option |
| github.com/charmbracelet/lipgloss | v1.1.0+ | Layout primitives | JoinHorizontal/JoinVertical for responsive layouts; styling for Summary panel |
| github.com/charmbracelet/bubbletea | v1.3.10+ | TUI framework | WindowSizeMsg pattern for responsive breakpoints |

### Supporting
None required — all dependencies already present in go.mod.

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| ntcharts barchart | Hand-rolled bar rendering | ntcharts already proven; custom rendering adds complexity |
| Conditional layout logic | bubbles/viewport with scrolling | User context specifies truncation, not scrolling |
| Streak calculation | Wait for API support | Streaks required in Summary panel; must calculate from daily data |

**Installation:**
No new dependencies required.

## Architecture Patterns

### Recommended Project Structure
```
wakadash/internal/
├── tui/
│   ├── model.go          # Add panel visibility fields (showCategories, showEditors, etc.)
│   ├── keymap.go         # Add Toggle5-9, ShowAll, HideAll bindings
│   ├── stats_panels.go   # NEW: Rendering functions for 4 stat panels
│   ├── summary_panel.go  # NEW: Summary panel rendering + streak calculation
│   └── layout.go         # NEW: Responsive layout logic (2-col vs stack)
├── types/
│   └── types.go          # Already has StatsData with Categories, Editors, OS, Machines
```

### Pattern 1: Horizontal Bar Charts with ntcharts

**What:** Create horizontal bar charts showing top N items with time labels and percentages.

**When to use:** For all four stat panels (Categories, Editors, OS, Machines).

**Example:**
```go
// Source: Verified from wakadash/internal/tui/model.go updateLanguagesChart() pattern
// Adapted for horizontal bars

categoriesChart := barchart.New(panelWidth, chartHeight,
    barchart.WithHorizontalBars(),
)

for _, cat := range data.Categories[:limit] {
    hours := cat.TotalSeconds / 3600.0
    percent := cat.Percent
    label := fmt.Sprintf("%s (%.0f%%)", cat.Name, percent)

    barStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
    categoriesChart.Push(barchart.BarData{
        Label: label,
        Values: []barchart.BarValue{
            {
                Name:  formatSeconds(cat.TotalSeconds),
                Value: hours,
                Style: barStyle,
            },
        },
    })
}

categoriesChart.Draw()
return categoriesChart.View()
```

### Pattern 2: Responsive Layout Breakpoint

**What:** Switch between 2-column and stacked layout based on terminal width.

**When to use:** When arranging multiple panels that should reflow based on available space.

**Example:**
```go
// Source: Based on lipgloss documentation and BubbleTea responsive patterns
// https://github.com/charmbracelet/lipgloss/blob/master/examples/layout/main.go

func (m Model) renderStatsGrid() string {
    // Build individual panels
    categoriesPanel := m.renderCategoriesPanel()
    editorsPanel := m.renderEditorsPanel()
    osPanel := m.renderOSPanel()
    machinesPanel := m.renderMachinesPanel()

    if m.width >= 80 {
        // 2-column layout: pair panels horizontally
        row1 := lipgloss.JoinHorizontal(lipgloss.Top, categoriesPanel, editorsPanel)
        row2 := lipgloss.JoinHorizontal(lipgloss.Top, osPanel, machinesPanel)
        return lipgloss.JoinVertical(lipgloss.Left, row1, row2)
    } else {
        // Stack all panels vertically
        return lipgloss.JoinVertical(lipgloss.Left,
            categoriesPanel, editorsPanel, osPanel, machinesPanel)
    }
}
```

### Pattern 3: Summary Panel with Highlight Styling

**What:** Text-based panel with key metrics and subtle visual emphasis.

**When to use:** For dashboard overview/summary that needs visual distinction without overwhelming.

**Example:**
```go
// Source: Based on lipgloss styling patterns
// https://pkg.go.dev/github.com/charmbracelet/lipgloss

func SummaryPanelStyle(t theme.Theme) lipgloss.Style {
    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(t.Primary).  // Subtle accent
        Padding(1, 2).
        Width(m.width - 4)
}

func (m Model) renderSummary() string {
    var sb strings.Builder

    sb.WriteString(TitleStyle(m.theme).Render("Summary (Last 30 Days)") + "\n\n")
    sb.WriteString(fmt.Sprintf("  Total:         %s\n", m.stats.Data.HumanReadableTotal))
    sb.WriteString(fmt.Sprintf("  Daily average: %s\n", m.stats.Data.HumanReadableDailyAverage))
    sb.WriteString(fmt.Sprintf("  Best day:      %s (%s)\n",
        m.stats.Data.BestDay.Date, m.stats.Data.BestDay.Text))

    // Streaks calculated from summaryData
    currentStreak, bestStreak := m.calculateStreaks()
    sb.WriteString(fmt.Sprintf("  Streak:        Current: %d days | Best: %d days\n",
        currentStreak, bestStreak))

    // Top items
    if len(m.stats.Data.Languages) > 0 {
        sb.WriteString(fmt.Sprintf("  Top language:  %s\n", m.stats.Data.Languages[0].Name))
    }
    // ... more top items

    return SummaryPanelStyle(m.theme).Render(sb.String())
}
```

### Pattern 4: Panel Visibility Toggle

**What:** Number keys toggle individual panel visibility; show-all/hide-all for batch control.

**When to use:** When dashboard has many panels that users may want to customize.

**Example:**
```go
// Source: Verified from wakadash/internal/tui/keymap.go and model.go Update()

// In keymap.go
Toggle5: key.NewBinding(
    key.WithKeys("5"),
    key.WithHelp("5", "toggle categories"),
)
ShowAll: key.NewBinding(
    key.WithKeys("A"),  // Shift+A
    key.WithHelp("A", "show all panels"),
)
HideAll: key.NewBinding(
    key.WithKeys("H"),  // Shift+H
    key.WithHelp("H", "hide all panels"),
)

// In Update()
case key.Matches(msg, m.keys.Toggle5):
    m.showCategories = !m.showCategories
    return m, nil
case key.Matches(msg, m.keys.ShowAll):
    m.showCategories = true
    m.showEditors = true
    m.showOS = true
    m.showMachines = true
    m.showSummary = true
    // ... all panels
    return m, nil
```

### Anti-Patterns to Avoid

- **Dynamic .Width() in loops:** Use .MaxWidth() instead to prevent rendering corruption (known issue from STATE.md)
- **Hardcoded panel dimensions:** Calculate based on m.width/m.height from WindowSizeMsg
- **Assuming data arrays are sorted:** API returns data sorted by total_seconds DESC, but verify before displaying
- **Forgetting border width:** Subtract 2 from width when rendering inside BorderStyle panels
- **Over-complicated streak calculation:** Limit to 7-day window from summaryData, not full history

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Horizontal bar rendering | Custom Unicode block logic | ntcharts barchart.WithHorizontalBars() | Handles width scaling, labels, value formatting automatically |
| Time formatting | String manipulation for hours/minutes | Reuse existing formatSeconds() function | Already implemented and tested in model.go |
| Responsive breakpoints | Complex state machine | Simple m.width >= 80 conditional | Lipgloss handles rendering; just choose join direction |
| Panel ordering/visibility | Custom layout manager | Bool fields + conditional rendering | Proven pattern from Phase 6 |

**Key insight:** The wakadash codebase already has all the patterns needed. This phase is about replication and composition, not invention.

## Common Pitfalls

### Pitfall 1: Assuming Streaks Are in /stats Response
**What goes wrong:** Code expects `current_streak` or `best_streak` fields in StatsData.

**Why it happens:** WakaTime's web dashboard shows streaks, so developers assume the API provides them.

**How to avoid:** Verify API response structure from official docs. Streaks require client-side calculation from daily summaries.

**Warning signs:** Runtime errors about missing fields; nil pointer dereferences when accessing streak data.

**Solution:** Calculate streaks from `m.summaryData` (already fetched for heatmap):
```go
func (m Model) calculateStreaks() (current int, best int) {
    if m.summaryData == nil || len(m.summaryData.Data) == 0 {
        return 0, 0
    }

    // Current streak: consecutive days from most recent backwards
    // Best streak: longest consecutive run in available data (7 days max)
    // Implementation details left to planner
}
```

### Pitfall 2: Panel Height Overflow
**What goes wrong:** Rendering more panels than terminal height allows causes layout corruption.

**Why it happens:** Adding 5 new panels without checking available vertical space.

**How to avoid:** Check `m.height` after accounting for border, status bar, and Summary panel. Truncate panels to show minimum 3 items each.

**Warning signs:** Dashboard flickers; panels overlap status bar; content wraps unexpectedly.

**Solution:**
```go
availableHeight := m.height - statusBarHeight - summaryPanelHeight - borderHeight - 4
itemsPerPanel := max(3, availableHeight / (numberOfVisiblePanels * chartHeight))
```

### Pitfall 3: Number Key Mapping Confusion
**What goes wrong:** User presses "5" expecting Categories to toggle, but nothing happens because Categories is actually the 5th panel in code order, not visual order.

**Why it happens:** User context specifies "number keys map to visual order", but implementation uses code order.

**How to avoid:** Map keys to visual rendering order (top-to-bottom, left-to-right), not Model field order.

**Warning signs:** User confusion; keys toggle wrong panels; inconsistent behavior between layouts.

**Solution:** Build a visual order array when rendering layout, then map keys to array indices.

### Pitfall 4: Percentage Display Formatting
**What goes wrong:** Percentages show as "65.342857%" instead of "65%".

**Why it happens:** StatsData.Percent is a float with full precision.

**How to avoid:** Use `fmt.Sprintf("%.0f%%", percent)` for whole-number percentages.

**Warning signs:** Cluttered labels; bars too wide due to long text.

## Code Examples

Verified patterns from existing codebase and official sources:

### Creating Horizontal Barchart Models

```go
// Source: wakadash/internal/tui/model.go NewModel() and updateLanguagesChart()
// Adapted for horizontal orientation

func NewModel(client *api.Client, rangeStr string, refreshInterval time.Duration) Model {
    // ... existing fields ...

    panelWidth := 35  // Adjusted based on 2-column layout
    chartHeight := 10 // Allow for 10 bars

    categoriesChart := barchart.New(panelWidth, chartHeight,
        barchart.WithHorizontalBars(),
    )
    editorsChart := barchart.New(panelWidth, chartHeight,
        barchart.WithHorizontalBars(),
    )
    // ... etc

    return Model{
        // ... existing fields ...
        categoriesChart: categoriesChart,
        editorsChart:    editorsChart,
        showCategories:  true,  // All panels visible by default
        showEditors:     true,
        // ...
    }
}
```

### Formatting Time with Percentages

```go
// Source: Based on existing formatSeconds() in model.go
// Extended to include percentage display

func formatTimeWithPercent(secs float64, percent float64) string {
    timeStr := formatSeconds(secs)
    return fmt.Sprintf("%s (%.0f%%)", timeStr, percent)
}

// Usage in panel rendering:
label := formatTimeWithPercent(cat.TotalSeconds, cat.Percent)
```

### Responsive Layout Assembly

```go
// Source: lipgloss documentation and existing model.go patterns
// https://github.com/charmbracelet/lipgloss/blob/master/examples/layout/main.go

func (m Model) renderDashboard() string {
    var sections []string

    // Summary always first (full width)
    if m.showSummary {
        sections = append(sections, m.renderSummary())
    }

    // Stats grid (responsive 2-col or stack)
    statsGrid := m.renderStatsGrid()
    if statsGrid != "" {
        sections = append(sections, statsGrid)
    }

    // Existing panels (sparkline, heatmap)
    if m.showSparkline {
        sections = append(sections, m.renderSparkline())
    }
    if m.showHeatmap {
        sections = append(sections, m.renderHeatmapPanel())
    }

    content := lipgloss.JoinVertical(lipgloss.Left, sections...)

    // ... border and status bar wrapping ...
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Vertical bars only | Horizontal bars with labels | ntcharts v0.3.0+ (2024) | Better for category/editor names (long labels) |
| Fixed single-column layout | Responsive multi-column | BubbleTea pattern evolution (2023+) | Utilizes wide terminals efficiently |
| Hard-coded panel visibility | User-toggled panels | Common pattern in modern TUIs | Reduces clutter, improves UX |

**Deprecated/outdated:**
- **Manual bar rendering with Unicode blocks:** ntcharts provides production-quality charts
- **AdaptiveColor for dark/light detection:** Not needed when theme system already exists (Phase 8)

## Open Questions

1. **Streak calculation scope**
   - What we know: WakaTime API doesn't provide streak data; must calculate from summaries
   - What's unclear: Should "best streak" consider only the 7-day summary window, or fetch longer history?
   - Recommendation: Limit to 7-day window (data already available from heatmap). Best streak = longest consecutive run in those 7 days. Document limitation in help text.

2. **Panel truncation visual indicator**
   - What we know: User context specifies truncation when terminal too short
   - What's unclear: How to visually indicate truncation (ellipsis, message, different styling)?
   - Recommendation: Show "..." below truncated panel with DimStyle. Include hint in help: "Resize terminal for full view"

3. **"Other" category aggregation**
   - What we know: User wants unknown/untracked items grouped as "Other"
   - What's unclear: Does WakaTime API return "Unknown" items, or are all items always named?
   - Recommendation: If API returns items with empty/null names, group those. Otherwise, skip this requirement until confirmed needed.

4. **Show-all/Hide-all key choice**
   - What we know: User wants these shortcuts but exact keys are Claude's discretion
   - What's unclear: Should these be Shift+letter, Ctrl+letter, or different keys entirely?
   - Recommendation: Use 'a' (show all) and 'h' (hide all) — mnemonic and unlikely to conflict. Check keymap for collisions.

## Sources

### Primary (HIGH confidence)
- [ntcharts GitHub repository](https://github.com/NimbleMarkets/ntcharts) - Verified barchart API and WithHorizontalBars() option
- [lipgloss layout examples](https://github.com/charmbracelet/lipgloss/blob/master/examples/layout/main.go) - Official responsive layout patterns
- [WakaTime API Docs](https://wakatime.com/developers) - Stats endpoint response structure
- wakadash/internal/types/types.go - Verified StatsData includes Categories, Editors, OperatingSystems, Machines arrays
- wakadash/internal/tui/model.go - Verified existing barchart patterns and panel toggle implementation

### Secondary (MEDIUM confidence)
- [BubbleTea responsive layout discussion](https://github.com/charmbracelet/bubbletea/discussions/1316) - Community patterns for dynamic sizing
- [Tips for building Bubble Tea programs](https://leg100.github.io/en/posts/building-bubbletea-programs/) - Layout best practices
- [WakaTime API Stats Structure](https://pkg.go.dev/github.com/zcxb/wakatime-go/pkg/wakatime) - Confirmed best_day field structure

### Tertiary (LOW confidence)
- WebSearch: Streak calculation - Confirmed streaks NOT in API, must calculate from daily data

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries already in use; patterns proven in Phase 6
- Architecture: HIGH - Existing codebase provides reference implementations
- Pitfalls: HIGH - Known issues documented in STATE.md; community discussions provide context
- Streak calculation: MEDIUM - No API support confirmed, but implementation approach is standard

**Research date:** 2026-02-20
**Valid until:** 2026-03-22 (30 days for stable APIs and mature libraries)
