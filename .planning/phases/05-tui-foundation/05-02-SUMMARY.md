---
phase: 05-tui-foundation
plan: 02
subsystem: ui
tags: [bubbletea, bubbles/help, bubbles/key, auto-refresh, ticker, help-overlay, keybindings]

# Dependency graph
requires:
  - phase: 05-01
    provides: Async bubbletea TUI foundation with Model/Init/Update/View
provides:
  - Auto-refresh ticker with configurable interval and countdown display
  - Help overlay with keyboard shortcuts via bubbles/help
  - Typed keybindings with help.KeyMap interface implementation
  - Manual refresh via 'r' key
  - Self-loop ticker pattern avoiding drift
affects: [06-multi-panel-layout, 07-homebrew-release]

# Tech tracking
tech-stack:
  added:
    - github.com/charmbracelet/bubbles/help (help overlay)
    - github.com/charmbracelet/bubbles/key (typed keybindings)
  patterns:
    - Self-loop ticker pattern (refreshMsg → fetch → statsFetchedMsg → scheduleRefresh)
    - Countdown ticker via countdownTickMsg self-loop with tickEverySecond()
    - help.KeyMap interface implementation for auto-generated help text
    - Separate help overlay mode toggled by ? key

key-files:
  created:
    - wakadash/internal/tui/keymap.go
  modified:
    - wakadash/internal/tui/model.go
    - wakadash/internal/tui/messages.go
    - wakadash/internal/tui/commands.go
    - wakadash/cmd/wakadash/main.go

key-decisions:
  - "Use self-loop ticker pattern (scheduleRefresh fires once, statsFetchedMsg schedules next) to avoid ticker drift"
  - "Start countdown ticker in Init() and self-loop via countdownTickMsg handler for smooth countdown display"
  - "Schedule refresh only from statsFetchedMsg/fetchErrMsg handlers to prevent double-ticker bug"
  - "Implement help.KeyMap interface (ShortHelp/FullHelp) for bubbles/help auto-generation"

patterns-established:
  - "Auto-refresh: fetch → statsFetchedMsg → scheduleRefresh(interval) → refreshMsg → fetch (self-loop)"
  - "Countdown: tickEverySecond() → countdownTickMsg → tickEverySecond() (1-second self-loop)"
  - "Help overlay: separate view mode toggled by ? key, using bubbles/help.View(keymap)"
  - "Status bar shows: last update time + countdown to next refresh"

# Metrics
duration: 4min
completed: 2026-02-19
---

# Phase 5 Plan 02: Auto-Refresh and Help Overlay Summary

**Dashboard with auto-refresh ticker, countdown display, and discoverable keyboard navigation via help overlay**

## Performance

- **Duration:** 4 min 8 sec
- **Started:** 2026-02-19T19:10:44Z
- **Completed:** 2026-02-19T19:14:52Z
- **Tasks:** 3
- **Files created:** 1
- **Files modified:** 4

## Accomplishments

- Keymap with typed key bindings (Quit, Help, Refresh) implementing help.KeyMap interface
- Auto-refresh ticker with self-loop pattern avoiding drift (refreshMsg → fetch → scheduleRefresh)
- Countdown ticker updating every second for "Next: X" display in status bar
- Help overlay toggled by ? key showing auto-generated keyboard shortcuts
- Manual refresh via 'r' key triggering immediate stats fetch
- --refresh flag added to main.go (default 60 seconds, 0 to disable)

## Task Commits

Each task was committed atomically:

1. **Task 1: Add keymap with help integration** - `3d4170f` (feat)
   - Created keymap.go with Quit/Help/Refresh bindings
   - Implemented ShortHelp() and FullHelp() for bubbles/help integration

2. **Task 2: Add refresh ticker messages and commands** - `3a97afd` (feat)
   - Added refreshMsg and countdownTickMsg timer message types
   - Created scheduleRefresh() and tickEverySecond() command factories

3. **Task 3: Extend model with refresh timer, help toggle, and keymap** - `03a11de` (feat)
   - Extended Model with refreshInterval, nextRefresh, help, keys, showHelp fields
   - Updated NewModel to accept refreshInterval parameter
   - Init() starts countdown ticker
   - Update() handles ?, r keys, refreshMsg, countdownTickMsg
   - View() renders help overlay when showHelp is true
   - Status bar shows countdown to next refresh
   - Added --refresh flag to main.go

## Files Created/Modified

**Created:**
- `wakadash/internal/tui/keymap.go` - Typed key bindings with help.KeyMap interface

**Modified:**
- `wakadash/internal/tui/model.go` - Extended with refresh timer, help overlay, keymap handling
- `wakadash/internal/tui/messages.go` - Added refreshMsg and countdownTickMsg types
- `wakadash/internal/tui/commands.go` - Added scheduleRefresh() and tickEverySecond()
- `wakadash/cmd/wakadash/main.go` - Added --refresh flag, updated NewModel call

## Decisions Made

- **Self-loop ticker pattern:** scheduleRefresh fires once after interval; statsFetchedMsg handler schedules the next refresh. This avoids ticker drift per research pitfall #3.
- **Countdown ticker:** Separate 1-second ticker for countdown display, self-looping via countdownTickMsg handler.
- **Single refresh scheduler:** Only statsFetchedMsg and fetchErrMsg call scheduleRefresh() to prevent double-ticker bugs.
- **help.KeyMap interface:** Implemented ShortHelp/FullHelp methods allowing bubbles/help to auto-generate help text.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - all tasks completed without blockers or unexpected issues.

## User Setup Required

None - auto-refresh works out of the box with sensible 60-second default. Users can customize via --refresh flag.

## Next Phase Readiness

- Phase 5 TUI Foundation complete - all must-haves delivered:
  - ✅ Dashboard auto-refreshes at configurable interval (visible countdown)
  - ✅ User can view keybinding help with ? key
  - ✅ User can manually refresh with r key
- Ready for Phase 6 (Multi-Panel Layout) - model extension pattern established for adding new panels
- No blockers or concerns

## Self-Check: PASSED

All files created:
- FOUND: internal/tui/keymap.go

All files modified:
- FOUND: internal/tui/model.go
- FOUND: internal/tui/messages.go
- FOUND: internal/tui/commands.go
- FOUND: cmd/wakadash/main.go

All commits exist:
- FOUND: 3d4170f (Task 1)
- FOUND: 3a97afd (Task 2)
- FOUND: 03a11de (Task 3)

---
*Phase: 05-tui-foundation*
*Completed: 2026-02-19*
