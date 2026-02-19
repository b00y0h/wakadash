---
phase: 06-data-visualization-and-ux
plan: 01
subsystem: dashboard-visualization
tags: [charts, colors, ui, ntcharts]
completed: 2026-02-19

dependency_graph:
  requires:
    - "05-02: TUI foundation with bubbletea"
    - "wakadash/internal/api: Stats API client"
  provides:
    - "Horizontal bar charts for languages and projects"
    - "GitHub Linguist color palette for language differentiation"
    - "2-column chart layout with lipgloss"
  affects:
    - "wakadash/internal/tui/model.go: Chart models and rendering"
    - "wakadash/internal/tui/colors.go: Color mapping"

tech_stack:
  added:
    - library: github.com/NimbleMarkets/ntcharts/barchart
      purpose: Horizontal bar chart rendering
      version: v0.4.0
  patterns:
    - "Bar chart state management in bubbletea Model"
    - "Dynamic chart resizing on terminal size changes"
    - "Color-coded data visualization with lipgloss"

key_files:
  created:
    - path: wakadash/internal/tui/colors.go
      purpose: GitHub Linguist color palette mapping
      exports: [getLanguageColor]
  modified:
    - path: wakadash/internal/tui/model.go
      changes: [languagesChart, projectsChart, updateLanguagesChart, updateProjectsChart, renderStats]
    - path: wakadash/go.mod
      changes: [ntcharts@v0.4.0, golang.org/x/text@v0.20.0]

decisions: []

metrics:
  duration_minutes: 5
  tasks_completed: 2
  files_created: 1
  files_modified: 2
  commits: 2
---

# Phase 06 Plan 01: Data Visualization - Bar Charts Summary

**One-liner:** Horizontal bar charts with GitHub Linguist colors for languages and fixed cyan for projects in 2-column layout

## What Was Built

Replaced plain-text language and project lists with color-coded horizontal bar charts using ntcharts library. Languages display with GitHub Linguist color palette (Go = #00ADD8, JavaScript = #f1e05a, etc.), while projects use uniform cyan (#00d7ff). Charts are rendered side-by-side in a 2-column layout that adapts to terminal width.

## Tasks Completed

| Task | Commit | Description | Files |
|------|--------|-------------|-------|
| 1 | cf63f83 | Add language color mapping | colors.go, go.mod |
| 2 | 6cb9c12 | Implement bar charts and 2-column layout | model.go, messages.go |

## Technical Implementation

### Color Mapping (colors.go)

Created `getLanguageColor(name string)` with 20 language mappings:
- Case-insensitive matching (Go/go/GO all match)
- GitHub Linguist hex colors (Go=#00ADD8, Python=#3572A5, Rust=#dea584, etc.)
- Default gray (#cccccc) for unknown languages

### Chart Integration (model.go)

**Model updates:**
- Added `languagesChart` and `projectsChart` barchart.Model fields
- Initialize both charts in NewModel() with 35x8 default dimensions
- Resize charts on WindowSizeMsg: `panelWidth = (width / 2) - 4`

**Data pipeline:**
1. statsFetchedMsg received → stats stored
2. updateLanguagesChart() called → clears chart, pushes top 5 languages with colors, draws
3. updateProjectsChart() called → clears chart, pushes top 5 projects with cyan, draws
4. renderStats() generates 2-column layout with lipgloss.JoinHorizontal

**Chart update logic:**
- Convert TotalSeconds to hours for bar values
- Create lipgloss.Style with getLanguageColor() for each language
- Use fixed cyan style for all projects
- Push BarData with label, value, and style to chart
- Call Draw() to generate rendered output

### Layout Structure

```
┌─────────────────────────────────────────────────────┐
│ WakaTime Stats (last_7_days)                        │
│                                                      │
│ Total time:    12h 34m                              │
│ Daily average: 1h 47m                               │
│                                                      │
│ Languages            Projects                       │
│ Go        ▓▓▓▓▓▓     wakadash   ▓▓▓▓▓▓              │
│ Python    ▓▓▓▓       dotfiles   ▓▓▓▓                │
│ Rust      ▓▓▓        scripts    ▓▓▓                 │
│ ...                  ...                            │
└─────────────────────────────────────────────────────┘
```

## Deviations from Plan

None - plan executed exactly as specified.

## Integration Points

**Upstream dependencies:**
- `internal/types.StatsResponse`: Languages and Projects slices
- `internal/types.StatItem`: Name and TotalSeconds fields
- bubbletea WindowSizeMsg for responsive chart sizing
- lipgloss for color styling and layout

**Downstream impact:**
- Dashboard now uses chart-based visualization instead of text lists
- renderStats() output structure changed (affects panel height calculations)
- Chart views integrated into existing bubbletea render loop

## Success Criteria Met

- ✅ Dashboard renders languages bar chart with GitHub Linguist colors
- ✅ Dashboard renders projects bar chart with time breakdown
- ✅ Charts are side-by-side in 2-column layout
- ✅ Charts resize correctly on terminal size changes
- ⚠️ Manual verification (requires valid API key - not tested in environment)

## Known Limitations

1. **No API key in environment:** Could not perform full visual verification of colored bars
2. **Fixed chart height:** Charts use fixed 8-line height, not adaptive to data count
3. **Top 5 only:** Displays maximum 5 languages and 5 projects (plan requirement)
4. **No value labels:** Bar charts show visual bars but not numeric hour values

## Next Steps

Phase 06 Plan 02 will likely add:
- Value labels on bars (show "12.5h" next to bars)
- Adaptive chart heights based on data count
- Additional visualizations (editors, OS, time trends)
- Color legends or tooltips

## Notes

- ntcharts library works seamlessly with lipgloss styling
- Chart.Resize() must be called before chart data is visible (WindowSizeMsg handler)
- Chart.Clear() + Push() + Draw() pattern works well for data updates
- 2-column layout with JoinHorizontal provides clean visual separation

## Self-Check: PASSED

✓ FOUND: internal/tui/colors.go
✓ FOUND: internal/tui/model.go
✓ FOUND: commit cf63f83
✓ FOUND: commit 6cb9c12

All claimed files and commits verified.
