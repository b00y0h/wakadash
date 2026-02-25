---
phase: quick-weekly-browser
plan: "01"
subsystem: tui
tags: [keyboard-shortcut, weekly-browser, navigation, historical-data]
dependency-graph:
  requires: [datasource.DataSource, archive.Fetcher, tui.ThemePickerModel pattern]
  provides: [WeeklyBrowserModel, fetchWeeklySummariesCmd, weeklyDataFetchedMsg, WeeklySummary]
  affects: [internal/tui/model.go, internal/tui/keymap.go]
tech-stack:
  added: []
  patterns: [BubbleTea sub-model delegation, ThemePickerModel pattern, tea.Cmd async scanning]
key-files:
  created:
    - internal/tui/weekly_browser.go
  modified:
    - internal/tui/messages.go
    - internal/tui/commands.go
    - internal/tui/keymap.go
    - internal/tui/model.go
    - internal/datasource/source.go
decisions:
  - "WeeklyBrowserModel receives theme at construction (NewWeeklyBrowser(t theme.Theme)) to avoid field access to parent model"
  - "Current week always included in browser list with 'current week' label instead of time, since live TotalSeconds may be 0 at scan time"
  - "Async scan skips weeks with zero TotalSeconds or nil data — resulting in a compact list of weeks with actual coding"
metrics:
  duration: "~4m 18s"
  completed: "2026-02-25"
  tasks-completed: 2
  files-changed: 6
---

# Quick Task 1: Weekly Browser ('w' key) Summary

**One-liner:** 'w' keyboard shortcut opens scrollable WeeklyBrowserModel overlay scanning up to 52 weeks of coding history with async data loading, arrow-key navigation, and Enter-to-navigate selection.

## What Was Built

A weekly history browser overlay triggered by pressing 'w' in the dashboard. It follows the same BubbleTea sub-model delegation pattern as the existing ThemePickerModel.

### New File: internal/tui/weekly_browser.go (257 lines)

Complete `WeeklyBrowserModel` with:
- Constructor `NewWeeklyBrowser(t theme.Theme)` — loading state by default
- `Update()` handling `weeklyDataFetchedMsg`, `tea.WindowSizeMsg`, and `tea.KeyMsg` (up/down/home/end/enter/esc/q)
- `View()` rendering: loading state, error state, scrollable week list with cursor marker
- Scrolling via `scrollOffset` clamped to keep selection visible in any terminal height
- Accessor methods: `IsConfirmed()`, `IsCancelled()`, `SelectedWeek()`
- Helper `formatWeekRangeFromStrings()` for "Jan 5-11" style display

### Updated: internal/tui/messages.go

Added:
- `WeeklySummary` struct: WeekStart, WeekEnd, TotalSeconds, TopLanguage, ProjectCount, HasData
- `weeklyDataFetchedMsg` struct: weeks []WeeklySummary, err error

### Updated: internal/tui/commands.go

Added:
- `fetchWeeklySummariesCmd(ds *datasource.DataSource, maxWeeks int) tea.Cmd` — scans backwards from current week, includes current week always, collects weeks with data
- `getWeekStartTime(t time.Time) time.Time` — helper for Sunday calculation (avoids import cycle with model.go's getWeekStart)

### Updated: internal/tui/keymap.go

Added:
- `WeeklyBrowser key.Binding` to keymap struct
- `WeeklyBrowser` binding mapped to `"w"` key in defaultKeymap
- `k.WeeklyBrowser` added to FullHelp navigation row

### Updated: internal/tui/model.go

Added:
- `showWeeklyBrowser bool` and `weeklyBrowser WeeklyBrowserModel` fields to Model struct
- Weekly browser delegation block in `Update()` (after picker delegation) — forwards WindowSizeMsg, weeklyDataFetchedMsg, and KeyMsg
- On confirm: navigates to selected week via `selectedWeekStart` and `fetchDataCmd` (current week clears selectedWeekStart)
- On cancel: closes browser, returns to dashboard
- `case key.Matches(msg, m.keys.WeeklyBrowser)` in main key handler — opens browser and fires `fetchWeeklySummariesCmd`
- `showWeeklyBrowser` guard in `View()` after picker guard
- Status bar hint updated to include "w weeks"

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed timezone mismatch in datasource.IsRecent causing boundary test failure**
- **Found during:** Task 2 verification (go test ./...)
- **Issue:** `time.Parse("2006-01-02", date)` returns UTC midnight while `today` used `now.Location()` (local timezone). When comparing "exactly 7 days ago" these were in different locations and compared unequal even though the dates were the same calendar date
- **Fix:** Changed to `time.ParseInLocation("2006-01-02", date, time.Local)` and `time.Local` for today — consistent local timezone comparison
- **Files modified:** internal/datasource/source.go
- **Commit:** 188fb18

**2. [Rule 3 - Blocking] Added theme parameter to NewWeeklyBrowser constructor**
- **Found during:** Task 1 implementation
- **Issue:** WeeklyBrowserModel needs theme colors for View() rendering. Plan specified `NewWeeklyBrowser() WeeklyBrowserModel` with no parameters, but the model has no access to the parent model's theme without being passed it
- **Fix:** Changed constructor signature to `NewWeeklyBrowser(t theme.Theme) WeeklyBrowserModel` and updated the call site in model.go to pass `m.theme`
- **Files modified:** internal/tui/weekly_browser.go, internal/tui/model.go

## Tasks Completed

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | Add weekly data fetching infrastructure and WeeklyBrowserModel | 770660a | internal/tui/weekly_browser.go (created), messages.go, commands.go |
| 2 | Wire weekly browser into main Model and add 'w' keybinding | 188fb18 | internal/tui/keymap.go, model.go, datasource/source.go (bug fix) |

## Self-Check: PASSED

All created/modified files exist. Both commits (770660a, 188fb18) verified in git log. Build passes. All tests pass.
