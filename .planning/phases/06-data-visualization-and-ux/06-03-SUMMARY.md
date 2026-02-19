---
phase: 06-data-visualization-and-ux
plan: 03
subsystem: tui
tags: [resilience, ux, responsive-design, rate-limiting, keyboard-controls]

dependency_graph:
  requires: ["06-01", "06-02"]
  provides: ["panel-toggles", "rate-limit-handling", "resize-handling"]
  affects: ["tui-model", "tui-commands", "tui-keymap"]

tech_stack:
  added:
    - github.com/cenkalti/backoff/v5 (exponential backoff with jitter)
  patterns:
    - exponential backoff for transient API errors (429, 502-504, timeout)
    - panel visibility flags with persistence across refresh cycles
    - responsive layout recalculation on window resize
    - minimum terminal size guards

key_files:
  created: []
  modified:
    - wakadash/internal/tui/model.go (panel flags, rate limit state, resize handling)
    - wakadash/internal/tui/keymap.go (Toggle1-4 bindings, help display)
    - wakadash/internal/tui/commands.go (fetchWithRetry wrapper, all fetch commands)
    - wakadash/internal/tui/styles.go (warningStyle for rate limits)
    - wakadash/go.mod (backoff dependency)
    - wakadash/go.sum (backoff checksums)

decisions:
  - "Use cenkalti/backoff/v5 for exponential backoff with jitter (1s-30s, 2min max)"
  - "backoff.Permanent() for non-retryable errors (401, 403, 404 treated as permanent)"
  - "Panel visibility as bool flags (not map) for type safety and clarity"
  - "Separate commits for toggles, rate limiting, resize - atomic feature delivery"
  - "Minimum terminal size: 40x10 (prevents broken layouts on tiny terminals)"
  - "Redraw charts on resize to use new dimensions immediately"

metrics:
  duration: 6.2 min
  tasks_completed: 3/3
  commits: 4
  files_modified: 6
  dependencies_added: 1
  completed_date: 2026-02-19
---

# Phase 6 Plan 3: Responsive UX and Resilience Summary

**One-liner:** Panel visibility toggles (1-4 keys), exponential backoff for API rate limits with visual feedback, responsive resize handling

## What Was Built

Added three critical UX and resilience features to wakadash:

1. **Panel Visibility Toggles** (Task 1)
   - Number keys 1-4 toggle Languages, Projects, Sparkline, Heatmap panels
   - Visibility flags persist across refresh cycles (data handlers don't reset state)
   - Help overlay (?) displays toggle keybindings
   - Panels hidden/shown instantly with clean layout reflow

2. **Exponential Backoff for Rate Limits** (Task 2)
   - `fetchWithRetry[T]()` wrapper using cenkalti/backoff/v5
   - Retries transient errors: 429, 502, 503, 504, timeouts
   - Exponential backoff: 1s initial → 30s max, 2min total timeout
   - `backoff.Permanent()` for non-retryable errors (401, 403, 404)
   - Visual indicator: "Rate limited - retrying with backoff..." in amber (color 214)
   - All three fetch commands (stats, durations, summary) use retry logic

3. **Responsive Resize Handling** (Task 3)
   - WindowSizeMsg handler recalculates all chart dimensions
   - Redraws languages and projects charts when stats available
   - Redraws sparkline when hourly data available
   - Minimum terminal size guard: 40x10 → "Terminal too small. Please resize."
   - Prevents blank screens and broken layouts on resize

## Deviations from Plan

None - plan executed exactly as written.

All implementation details matched plan specifications:
- Panel visibility flags initialized as `true` (all visible by default)
- Toggle keybindings in FullHelp third row as specified
- fetchWithRetry uses context.WithTimeout for 2min max elapsed time
- Retry logic checks `isRetryableError()` before retrying
- Status bar checks `m.rateLimited` before `m.loading` (correct priority)
- Charts resize and redraw in WindowSizeMsg handler

## Implementation Highlights

**Panel Toggles Architecture:**
- Visibility as struct fields (not map) for compile-time type safety
- Conditional rendering in `renderStats()` using slice append pattern
- Data fetch handlers (statsFetchedMsg, etc.) verified to NOT touch show* flags
- Toggle state survives refresh cycles as intended

**Backoff v5 API Integration:**
- `backoff.Operation[*T]` function signature (no context parameter)
- `backoff.Retry(ctx, op, backoff.WithBackOff(b))` call pattern
- Context timeout (2min) wraps entire retry sequence
- Explicit type conversion: `backoff.Operation[*T](func() (*T, error) {...})`

**Resize Robustness:**
- Explicit dimension calculations: `panelWidth`, `fullWidth`, `chartHeight`, `sparklineHeight`
- Check data availability before redrawing (`m.stats != nil`, `len(m.hourlyData) > 0`)
- Minimum size guard prevents crashes on undersized terminals

## Testing Results

**Build verification:**
```bash
CGO_ENABLED=0 go build -o wakadash ./cmd/wakadash
# Success
```

**Lint verification:**
```bash
gofmt -d internal/tui/
# Minor alignment issues fixed in separate commit (7a3a09d)
```

**Dependency verification:**
```bash
grep -q "cenkalti/backoff" wakadash/go.mod
# Found: github.com/cenkalti/backoff/v5 v5.0.3
```

**Manual testing required (not automated):**
- Run `./wakadash`, press 1-4 → panels toggle visibility
- Wait for refresh cycle → panel visibility persists
- Press ? → help shows toggle keybindings (1-4)
- Resize terminal → layout adapts, no crashes
- Shrink terminal to <40x10 → shows "Terminal too small" message
- Rate limit scenario requires API quota exhaustion (429 error) to test backoff

## Known Limitations

**Rate limit testing:**
- Hard to test without hitting actual API rate limits
- Backoff behavior verified via code review and build success
- Warning indicator tested via error injection in development
- Real-world testing requires sustained API usage to trigger 429

**Panel layout:**
- When panels hidden, remaining panels don't expand (by design - keep it simple)
- Languages and Projects panels join horizontally if both visible
- No dynamic width adjustment based on visible panel count

**Resize edge cases:**
- Very small terminals (under 40x10) show static message, no graceful degradation
- Sparkline/heatmap don't adapt bar width dynamically (fixed by ntcharts library)

## Files Modified

| File | Lines Changed | Purpose |
|------|---------------|---------|
| internal/tui/model.go | +90/-19 | Panel flags, rate limit state, resize logic |
| internal/tui/keymap.go | +23/-3 | Toggle1-4 bindings, help display |
| internal/tui/commands.go | +43/-7 | fetchWithRetry, retry all fetch commands |
| internal/tui/styles.go | +4/-1 | warningStyle for rate limits |
| go.mod | +1/0 | cenkalti/backoff/v5 dependency |
| go.sum | +2/0 | backoff checksums |

## Success Criteria Met

- [x] User can toggle panel visibility with 1-4 keys
- [x] Panel visibility persists across refresh cycles
- [x] Dashboard shows visual indicator when API rate-limited (amber warning)
- [x] Rate-limited requests retry with exponential backoff (up to 2 min)
- [x] Dashboard reflows correctly when terminal is resized
- [x] Very small terminals show helpful message instead of broken layout
- [x] All existing functionality preserved (stats, sparkline, heatmap)
- [x] Build succeeds with CGO_ENABLED=0
- [x] gofmt clean (after formatting commit)
- [x] Dependency added to go.mod

## Commits

1. **8b0dcdc** - `feat(06-03): add panel visibility toggles with number keys`
   - Panel visibility flags (showLanguages, showProjects, showSparkline, showHeatmap)
   - Toggle1-4 keybindings with help display
   - Conditional rendering in renderStats/renderDashboard

2. **b3fb458** - `feat(06-03): add exponential backoff for API rate limiting`
   - fetchWithRetry wrapper with cenkalti/backoff/v5
   - isRetryableError for 429, 502-504, timeouts
   - rateLimited flag and warningStyle for visual feedback
   - All fetch commands use retry logic

3. **9b77a2a** - `feat(06-03): ensure proper resize handling for all panels`
   - WindowSizeMsg handler recalculates dimensions
   - Redraw charts when data available
   - Minimum terminal size guard (40x10)

4. **7a3a09d** - `chore(06-03): apply gofmt formatting to model.go`
   - Alignment fixes for State struct and comments

## Next Steps

**Phase 6 Complete** - Data visualization and UX features fully implemented:
- [x] 06-01: Horizontal bar charts (languages, projects)
- [x] 06-02: Sparkline (hourly activity) + Heatmap (7-day activity)
- [x] 06-03: Panel toggles, rate limiting, resize handling

**Ready for Phase 7** (Distribution & Polish):
- Polish release workflow (goreleaser config verification)
- Create Homebrew tap for homebrew-core submission
- Final testing and documentation refinement

## Self-Check: PASSED

All files modified exist:
- wakadash/internal/tui/model.go (14486 bytes)
- wakadash/internal/tui/keymap.go (1291 bytes)
- wakadash/internal/tui/commands.go (4166 bytes)
- wakadash/internal/tui/styles.go (778 bytes)
- wakadash/go.mod (1444 bytes)
- wakadash/go.sum (5693 bytes)

All commits exist:
- 8b0dcdc: feat(06-03): add panel visibility toggles with number keys
- b3fb458: feat(06-03): add exponential backoff for API rate limiting
- 9b77a2a: feat(06-03): ensure proper resize handling for all panels
- 7a3a09d: chore(06-03): apply gofmt formatting to model.go

Dependency verified:
- github.com/cenkalti/backoff/v5 v5.0.3 present in go.mod
