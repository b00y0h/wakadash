---
phase: quick-4
plan: 01
subsystem: ui
tags: [bubbletea, tui, theme-picker, weekly-browser, terminal-dimensions]

# Dependency graph
requires:
  - phase: quick-1
    provides: WeeklyBrowserModel with NewWeeklyBrowser constructor
  - phase: quick-2
    provides: ThemePickerModel rendering
provides:
  - ThemePickerModel initialized with correct terminal dimensions at creation
  - WeeklyBrowserModel initialized with correct terminal dimensions at creation
affects: [any future sub-model additions following the same delegation pattern]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Sub-model constructors receive parent dimensions so View() works correctly before WindowSizeMsg arrives"

key-files:
  created: []
  modified:
    - internal/tui/picker.go
    - internal/tui/weekly_browser.go
    - internal/tui/model.go
    - cmd/wakadash/main.go

key-decisions:
  - "Pass current terminal dimensions (m.width, m.height) to sub-model constructors at creation time — avoids 0x0 initial state"
  - "First-run picker in main.go passes 0,0 — acceptable because BubbleTea sends WindowSizeMsg before first render in standalone program mode"

patterns-established:
  - "Sub-model constructor pattern: always pass parent width/height so View() is correct immediately on opening"

# Metrics
duration: 8min
completed: 2026-02-25
---

# Quick Task 4: Fix Terminal Too Small Error When Opening Theme Picker

**ThemePickerModel and WeeklyBrowserModel now receive parent terminal dimensions at construction, eliminating the false "Terminal too small" error on first render**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-02-25T20:00:00Z
- **Completed:** 2026-02-25T20:08:55Z
- **Tasks:** 1
- **Files modified:** 4

## Accomplishments

- Fixed root cause: `NewThemePicker` and `NewWeeklyBrowser` previously created structs with `width=0, height=0` (Go zero values), causing `View()` to always return "Terminal too small" on first render
- Updated both constructors to accept `width, height int` parameters and initialize those fields
- Updated all three call sites: runtime picker in `model.go` (`ChangeTheme` handler), runtime browser in `model.go` (`WeeklyBrowser` handler), and first-run picker in `main.go`
- All existing tests pass; `go build ./...` and `go vet ./...` succeed with no issues

## Task Commits

1. **Task 1: Pass terminal dimensions to sub-models on creation** - `d1f74ea` (fix)

## Files Created/Modified

- `internal/tui/picker.go` - `NewThemePicker(isFirstRun bool, width, height int)` — sets width/height on construction
- `internal/tui/weekly_browser.go` - `NewWeeklyBrowser(t theme.Theme, width, height int)` — sets width/height on construction
- `internal/tui/model.go` - Updated ChangeTheme and WeeklyBrowser handlers to pass `m.width, m.height`
- `cmd/wakadash/main.go` - Updated first-run call to `NewThemePicker(true, 0, 0)`

## Decisions Made

- First-run picker in `main.go` passes `0, 0` — the first-run picker is run as a standalone BubbleTea program with `tea.WithAltScreen()`, which sends a `WindowSizeMsg` before the first render, so `0, 0` is immediately overridden before View() is ever called
- Belt-and-suspenders approach not needed: since the constructor now sets the dimensions, no need to also set `m.picker.width = m.width` after construction

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Theme picker now opens correctly when 't' is pressed, showing the full preview at any reasonable terminal size
- Weekly browser also benefits from the same fix for consistency
- Pattern established: future sub-models added to the dashboard should follow the same constructor pattern

---
*Phase: quick-4*
*Completed: 2026-02-25*

## Self-Check: PASSED

- `internal/tui/picker.go` - FOUND (modified, NewThemePicker accepts width/height)
- `internal/tui/weekly_browser.go` - FOUND (modified, NewWeeklyBrowser accepts width/height)
- `internal/tui/model.go` - FOUND (modified, both call sites updated)
- `cmd/wakadash/main.go` - FOUND (modified, first-run call site updated)
- Commit `d1f74ea` - FOUND (verified via git log)
