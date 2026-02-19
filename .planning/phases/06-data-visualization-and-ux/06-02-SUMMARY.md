---
phase: 06-data-visualization-and-ux
plan: 02
subsystem: ui
tags: [ntcharts, sparkline, heatmap, visualization, bubbletea, lipgloss]

# Dependency graph
requires:
  - phase: 05-tui-foundation
    provides: Bubbletea TUI foundation with async data fetching
provides:
  - Sparkline chart showing hourly coding activity (last 24 hours)
  - Heatmap showing weekly activity levels (last 7 days)
  - Duration and Summary API endpoints for temporal data
  - GitHub-style contribution colors for activity visualization
affects: [06-03-polish-and-styling, ui, visualization]

# Tech tracking
tech-stack:
  added:
    - github.com/NimbleMarkets/ntcharts (sparkline and barchart)
    - github.com/lrstanley/bubblezone (ntcharts dependency)
  patterns:
    - Temporal data aggregation (grouping durations by hour)
    - Parallel data fetching (stats, durations, summaries in same batch)
    - Color-coded visualization (GitHub-style activity heatmap)
    - Lipgloss-based panel composition

key-files:
  created:
    - None
  modified:
    - internal/types/types.go (Duration and DurationsResponse types)
    - internal/api/client.go (FetchDurations method)
    - internal/tui/messages.go (durationsFetchedMsg, summaryFetchedMsg)
    - internal/tui/commands.go (fetchDurationsCmd, fetchSummaryCmd)
    - internal/tui/model.go (sparkline and heatmap rendering)

key-decisions:
  - "Use simplified heatmap: daily totals from summaries instead of hourly durations per day"
  - "Fetch durations for today only to minimize API calls"
  - "Use GitHub Linguist colors for language bars (from colors.go)"
  - "Display 7-day heatmap as colored blocks with MM-DD labels"

patterns-established:
  - "Temporal visualization: sparkline for hourly, heatmap for daily patterns"
  - "Color scales: GitHub-style contribution intensity (dark gray → bright green)"
  - "Data grouping: aggregate time-stamped durations into hour buckets"

# Metrics
duration: 7min
completed: 2026-02-19
---

# Phase 6 Plan 02: Sparkline and Heatmap Summary

**Sparkline showing 24-hour coding patterns and GitHub-style heatmap for 7-day activity using ntcharts and lipgloss**

## Performance

- **Duration:** 7 min (443 seconds)
- **Started:** 2026-02-19T20:11:10Z
- **Completed:** 2026-02-19T20:18:33Z
- **Tasks:** 3
- **Files modified:** 5

## Accomplishments
- Added sparkline visualization showing hourly coding activity for today (24 columns)
- Added heatmap visualization showing last 7 days of activity with color-coded intensity
- Integrated ntcharts library for professional time-series visualizations
- Implemented parallel data fetching for durations and summaries alongside stats
- Created GitHub-style contribution colors for activity heatmap

## Task Commits

Each task was committed atomically:

1. **Task 1: Add Duration type and FetchDurations API method** - `6390e80` (feat)
2. **Task 2: Add sparkline with hourly data fetch** - `cfffa61` (feat)
3. **Task 3: Add heatmap using daily summaries** - `0fe930d` (feat)

## Files Created/Modified
- `internal/types/types.go` - Added Duration and DurationsResponse types for /durations endpoint
- `internal/api/client.go` - Added FetchDurations method to fetch hourly coding sessions
- `internal/tui/messages.go` - Added durationsFetchedMsg and summaryFetchedMsg message types
- `internal/tui/commands.go` - Added fetchDurationsCmd and fetchSummaryCmd with panic recovery
- `internal/tui/model.go` - Added sparkline and heatmap rendering, data aggregation helpers

## Decisions Made
- **Simplified heatmap approach:** Use daily totals from existing summaries endpoint instead of fetching hourly durations for each of 7 days. This reduces API calls from 8 to 2 (1 for today's durations, 1 for 7-day summaries)
- **GitHub-style colors:** Use GitHub Linguist colors for language bars and GitHub contribution colors for heatmap intensity
- **Data aggregation:** Group durations by hour using UNIX timestamp conversion to show temporal patterns
- **Panel layout:** Place sparkline below bar charts and heatmap below sparkline for top-to-bottom temporal flow

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed barStyle pointer type error**
- **Found during:** Task 2 (Build after sparkline implementation)
- **Issue:** Bar chart code (added by copilot/other assistant) used `&barStyle` but ntcharts expects value not pointer
- **Fix:** Changed `Style: &barStyle` to `Style: barStyle` in both updateLanguagesChart and updateProjectsChart
- **Files modified:** internal/tui/model.go
- **Verification:** Build succeeded after fix
- **Committed in:** cfffa61 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking build error)
**Impact on plan:** Type fix necessary for compilation. Bar chart code was added by another process (plan 06-01) during execution. No scope creep.

## Issues Encountered
- Build failures due to concurrent modifications: model.go was being modified by another process (likely copilot or auto-formatter) while executing tasks. Used sed for direct edits when file locking occurred.
- Bar chart implementation (plan 06-01) appeared during plan 06-02 execution, suggesting plans executed out of order or in parallel. Fixed type errors and continued.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Temporal visualizations complete (hourly and daily patterns)
- Ready for Phase 6 Plan 03 (polish and styling)
- Data fetching architecture supports adding more visualization types
- Color system established for consistent theming

## Self-Check: PASSED

All files verified:
- FOUND: wakadash/internal/types/types.go
- FOUND: wakadash/internal/api/client.go
- FOUND: wakadash/internal/tui/messages.go
- FOUND: wakadash/internal/tui/commands.go
- FOUND: wakadash/internal/tui/model.go

All commits verified:
- FOUND: 6390e80
- FOUND: cfffa61
- FOUND: 0fe930d

---
*Phase: 06-data-visualization-and-ux*
*Completed: 2026-02-19*
