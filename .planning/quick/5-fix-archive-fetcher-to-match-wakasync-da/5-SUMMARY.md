---
phase: quick-5
plan: 01
subsystem: archive
tags: [github-archive, wakasync, http, json-parsing]

# Dependency graph
requires:
  - phase: 12-github-archive
    provides: "Original archive fetcher implementation"
provides:
  - "Archive fetcher with correct wakasync URL pattern (data/YYYY/MM/DD/summary.json)"
  - "SummaryResponse envelope unwrapping for archive JSON"
  - "CheckAccess() method for startup repo accessibility validation"
affects: [datasource, tui]

# Tech tracking
tech-stack:
  added: []
  patterns: ["SummaryResponse envelope unwrapping for archive data"]

key-files:
  created: []
  modified:
    - internal/archive/fetcher.go
    - internal/archive/fetcher_test.go

key-decisions:
  - "Keep 404 returning (nil, nil) for backward compatibility with week-scanning logic"
  - "Add separate CheckAccess() method instead of erroring on every 404"

patterns-established:
  - "Archive data uses wakasync URL structure: data/YYYY/MM/DD/summary.json"
  - "Archive JSON is SummaryResponse wrapper, not raw DayData"

# Metrics
duration: 2min
completed: 2026-03-04
---

# Quick Task 5: Fix Archive Fetcher to Match Wakasync Data Structure

**Archive fetcher URL changed to data/YYYY/MM/DD/summary.json with SummaryResponse envelope unwrapping and CheckAccess startup validation**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-04T16:23:22Z
- **Completed:** 2026-03-04T16:25:39Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Fixed URL pattern from `data/YYYY-MM-DD.json` to `data/YYYY/MM/DD/summary.json` matching wakasync's actual structure
- Added SummaryResponse envelope unwrapping (decode into wrapper, extract Data[0])
- Added CheckAccess() method for one-time startup repo accessibility check with private-repo guidance
- Added comprehensive tests: URL pattern validation, empty data array, invalid date, and CheckAccess scenarios

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix URL pattern, unwrap SummaryResponse, and add private-repo 404 warning** - `142b362` (fix)
2. **Task 2: Update tests for new URL pattern, SummaryResponse wrapper, and CheckAccess** - `1556a2d` (test)

## Files Created/Modified
- `internal/archive/fetcher.go` - Fixed URL pattern, SummaryResponse decoding, date validation, CheckAccess method
- `internal/archive/fetcher_test.go` - Updated success test to use wrapped JSON, added 6 new test functions

## Decisions Made
- Kept 404 returning `(nil, nil)` instead of an error to maintain backward compatibility with `FindNonEmptyWeek` and `HasOlderData` scanning logic in datasource
- Added `CheckAccess()` as a separate startup-time method rather than making every 404 noisy, since missing dates are normal for public repos

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Archive fetcher now correctly reads wakasync data
- CheckAccess() available for callers to validate repo accessibility at startup (not yet wired into startup flow)

---
*Quick Task: 5-fix-archive-fetcher-to-match-wakasync-da*
*Completed: 2026-03-04*
