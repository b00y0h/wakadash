---
phase: 17-background-prefetch-and-improved-no-data-ux
plan: 01
subsystem: navigation
tags:
  - prefetch
  - performance
  - ux
  - cache
dependency_graph:
  requires:
    - "16-02: Auto-refresh pause during history view"
    - "14-03: Week navigation with auto-skip"
  provides:
    - "Instant backward navigation via prefetch cache"
    - "Silent background data loading"
  affects:
    - "Navigation UX (instant vs loading)"
    - "Data fetching patterns"
tech_stack:
  added:
    - "Prefetch cache (map[string]*types.DayData)"
  patterns:
    - "Predictive data loading"
    - "Silent background commands"
    - "Cache-first navigation"
key_files:
  created: []
  modified:
    - "wakadash/internal/tui/model.go"
    - "wakadash/internal/tui/commands.go"
    - "wakadash/internal/tui/messages.go"
decisions:
  - id: "prefetch-silent-failure"
    summary: "Prefetch errors are silently discarded - no UI feedback"
    rationale: "Background prefetch is an optimization; failures shouldn't interrupt user experience"
  - id: "prefetch-one-week-ahead"
    summary: "Always prefetch exactly one week ahead during navigation"
    rationale: "Users most commonly navigate one week back; continuous prefetch maintains instant feel"
  - id: "cache-stores-nil"
    summary: "Cache stores nil for weeks with no data"
    rationale: "Avoids re-fetching known empty weeks; error is discarded, nil indicates 'checked and empty'"
metrics:
  duration_seconds: 142
  tasks_completed: 3
  files_modified: 3
  commits: 3
  completed_date: 2026-02-25
---

# Phase 17 Plan 01: Background Prefetch Implementation Summary

**One-liner:** Instant backward navigation via silent background prefetch of previous week data

## What Was Built

Implemented background prefetch system that makes navigating to the previous week feel instant by pre-loading data after the dashboard loads.

**Key mechanism:**
1. After dashboard loads (dataFetchedMsg), quietly fetch previous week in background
2. Store result in prefetch cache (map[weekStart] -> data)
3. When user presses prev-week, check cache first - instant if found
4. Continue prefetching one week ahead as user navigates backward

**User experience:**
- First backward navigation: Instant (data already loaded)
- Continued backward navigation: Instant (continuous prefetch)
- No loading indicators for cached data
- No error messages for prefetch failures

## Tasks Completed

### Task 1: Add prefetch cache and message types
**Commit:** f216376

Added infrastructure for prefetch system:
- `prefetchResultMsg` message type for background prefetch results
- `prefetchedData map[string]*types.DayData` in Model struct
- Cache initialization in NewModel constructor

**Files modified:**
- `wakadash/internal/tui/messages.go`: Added prefetchResultMsg type
- `wakadash/internal/tui/model.go`: Added cache field and initialization

### Task 2: Create prefetch command and trigger after initial load
**Commit:** ecab989

Implemented background fetch command and trigger logic:
- `prefetchWeekCmd` function to fetch data in background (silent errors)
- `getPreviousWeekStart` helper to calculate previous week dates
- Trigger prefetch after dataFetchedMsg (after main UI loads)
- Cache check before prefetching (avoid duplicate fetches)

**Files modified:**
- `wakadash/internal/tui/commands.go`: Added prefetchWeekCmd
- `wakadash/internal/tui/model.go`: Added helper and trigger in dataFetchedMsg handler

### Task 3: Handle prefetch result and use cache on navigation
**Commit:** 85e0918

Wired up cache to navigation system:
- prefetchResultMsg handler stores results in cache
- PrevWeek handler checks cache before fetching (instant navigation)
- Continuous prefetch as user navigates (always one week ahead)
- Updated atOldestData flag when using cached data

**Files modified:**
- `wakadash/internal/tui/model.go`: Added handler and cache lookup in PrevWeek

## Deviations from Plan

None - plan executed exactly as written.

## Key Decisions

**1. Silent failure on prefetch errors (Decision ID: prefetch-silent-failure)**
- Prefetch errors are discarded without UI feedback
- Background optimization shouldn't interrupt user experience
- If prefetch fails, navigation falls through to normal fetch

**2. Prefetch one week ahead (Decision ID: prefetch-one-week-ahead)**
- Always prefetch exactly one week before current view
- Continuous prefetching during backward navigation
- Optimizes for common use case (navigating recent history)

**3. Cache stores nil for no-data weeks (Decision ID: cache-stores-nil)**
- Error is discarded, nil indicates "checked and empty"
- Avoids re-fetching known empty weeks
- Maintains cache effectiveness for sparse data

## Testing & Verification

All verification checks passed:
- ✅ gofmt clean on all modified files
- ✅ prefetchedData field exists in Model struct
- ✅ prefetchWeekCmd function exists in commands.go
- ✅ prefetchResultMsg type exists in messages.go
- ✅ Cache lookup in PrevWeek handler
- ✅ Prefetch triggers after dataFetchedMsg

## Performance Impact

**Expected improvements:**
- First backward navigation: 0ms (instant from cache)
- Continued backward navigation: 0ms (continuous prefetch)
- Network: Prefetch happens during idle time after initial load

**Memory impact:**
- Small cache overhead (map stores pointers to DayData)
- Cache grows as user navigates (one entry per visited week)
- Bounded by user navigation depth (typically < 10 weeks)

## Integration Points

**Integrates with:**
- Phase 14-03: Week navigation (auto-skip empty weeks)
- Phase 16-02: Auto-refresh pause during history
- DataSource: Hybrid API/archive fetching

**Enables future work:**
- Phase 17-02: Improved no-data UX (can detect nil cached data)
- Additional prefetch strategies (forward navigation, multiple weeks)

## Self-Check: PASSED

**Files verified:**
- ✅ `/workspace/wakadash/internal/tui/model.go` - prefetchedData field and cache logic exists
- ✅ `/workspace/wakadash/internal/tui/commands.go` - prefetchWeekCmd function exists
- ✅ `/workspace/wakadash/internal/tui/messages.go` - prefetchResultMsg type exists

**Commits verified:**
- ✅ f216376 - Add prefetch cache and message types
- ✅ ecab989 - Create prefetch command and trigger
- ✅ 85e0918 - Handle prefetch result and use cache

All artifacts exist and commits are in git history.
