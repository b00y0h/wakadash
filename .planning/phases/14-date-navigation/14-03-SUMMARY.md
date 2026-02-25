---
phase: 14-date-navigation
plan: 03
subsystem: tui-navigation
tags: [week-navigation, auto-skip, end-of-history, ux]
dependency_graph:
  requires: [14-02]
  provides: [auto-skip-blank-weeks, oldest-data-indicator]
  affects: [tui-model, datasource-api]
tech_stack:
  added: []
  patterns: [async-search, data-availability-checking]
key_files:
  created: []
  modified:
    - wakadash/internal/datasource/source.go
    - wakadash/internal/tui/commands.go
    - wakadash/internal/tui/messages.go
    - wakadash/internal/tui/model.go
decisions:
  - summary: "Search limit of 52 weeks (1 year) prevents infinite search loops"
    rationale: "Balances thoroughness with performance - most users won't have gaps longer than a year"
  - summary: "HasOlderData checks 4 weeks back to avoid false negatives from sparse data"
    rationale: "Single-week check could miss data in archives with gaps, multiple weeks increases confidence"
  - summary: "Auto-skip only on backward navigation, not forward navigation"
    rationale: "Forward navigation has clear boundary (current week), backward needs smart detection"
metrics:
  duration_seconds: 219
  tasks_completed: 3
  files_modified: 4
  commits: 3
  completed_date: 2026-02-25
---

# Phase 14 Plan 03: Auto-Skip Blank Weeks and End-of-History Indicator Summary

Auto-skip logic for week navigation with visual end-of-history indicator using async search commands and data availability checking.

## What Was Built

### Data Availability Detection
**DataSource.FindNonEmptyWeek**: Searches backward/forward through weeks to find next week with data, checking API for recent dates and archive for historical dates. Respects 52-week search limit to prevent unbounded searches.

**DataSource.HasOlderData**: Multi-week check (up to 4 weeks back) to determine if older data exists beyond current week, avoiding false negatives from sparse archive data.

### Async Week Search
**findNonEmptyWeekCmd**: BubbleTea command that performs async week search without blocking UI, with panic recovery and proper error handling following established command patterns.

**weekSearchResultMsg**: Message type communicating search results including found week, success flag, and end-of-history detection to Model for state updates.

### Navigation Integration
**Auto-skip on PrevDay**: Previous week navigation now triggers async search for next week with data instead of blindly jumping to previous week, improving UX for archives with gaps.

**End-of-history tracking**: Model.atOldestData flag tracks when viewing oldest available data, automatically cleared when navigating forward or returning to today.

**Status bar indicator**: Visual "[oldest data]" warning displayed in status bar when at end of history, using theme's warning styling for clear visibility.

## Success Criteria Met

- [x] DataSource.FindNonEmptyWeek searches backward for non-empty weeks
- [x] DataSource.HasOlderData checks if older data exists
- [x] findNonEmptyWeekCmd executes search asynchronously
- [x] weekSearchResultMsg communicates search results
- [x] atOldestData field tracks end-of-history state
- [x] PrevDay handler triggers async week search
- [x] weekSearchResultMsg handler updates state and fetches data
- [x] Status bar shows "[oldest data]" when atOldestData is true
- [x] Code compiles without errors
- [x] Code formats cleanly with gofmt

## Task Breakdown

| Task | Name | Commit | Files Modified |
|------|------|--------|----------------|
| 1 | Add data availability check to DataSource | 3298574 | source.go |
| 2 | Add async week search command and message | 2d6bdfa | commands.go, messages.go |
| 3 | Update navigation handlers with auto-skip and end-of-history | be54293 | model.go |

## Technical Details

### Auto-Skip Search Pattern
```go
// Search starts from week before current position
prevWeekStart := searchStart.AddDate(0, 0, -7)
m.loading = true
return m, findNonEmptyWeekCmd(m.dataSource, prevWeekStart.Format("2006-01-02"), -1)
```

Direction parameter (-1 for backward, 1 for forward) allows extensibility for future forward auto-skip if needed.

### Multi-Week Availability Check
```go
// Try up to 4 weeks back to avoid false negatives
for i := 2; i <= 4; i++ {
    checkWeek := parsed.AddDate(0, 0, -7*i).Format("2006-01-02")
    data, err := ds.archive.FetchArchive(checkWeek)
    if err == nil && data != nil && data.GrandTotal.TotalSeconds > 0 {
        return true
    }
}
```

Prevents incorrectly showing "oldest data" when there are sparse gaps in archive.

### State Management
atOldestData flag managed through:
- Set to true when week search returns no results or finds oldest week
- Cleared when navigating forward (NextDay)
- Cleared when returning to current week (Today)
- Persists when navigating backward through historical data

## Deviations from Plan

None - plan executed exactly as written.

## Testing Notes

**Manual verification recommended:**
1. Navigate backward through weeks with archive gaps to verify auto-skip behavior
2. Continue backward until reaching oldest data to verify indicator appears
3. Navigate forward to verify indicator clears appropriately
4. Press 't' (today) from oldest data to verify return to live view

**Edge cases to test:**
- Archives with large gaps (should skip to next week with data)
- Reaching 52-week search limit (should show oldest indicator)
- Sparse data at boundary (HasOlderData checks 4 weeks back)
- Navigating forward from oldest week (indicator should clear)

## Self-Check

Verifying all claimed files and commits exist:

**Files:**
- ✓ wakadash/internal/datasource/source.go
- ✓ wakadash/internal/tui/commands.go
- ✓ wakadash/internal/tui/messages.go
- ✓ wakadash/internal/tui/model.go

**Commits:**
- ✓ 3298574 (Task 1: DataSource methods)
- ✓ 2d6bdfa (Task 2: Async week search)
- ✓ be54293 (Task 3: Navigation integration)

## Self-Check: PASSED

All files and commits verified.
