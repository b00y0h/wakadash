---
phase: 13-hybrid-data-fetching
plan: 01
subsystem: data-access
tags: [hybrid, routing, tdd, api, archive]
dependency_graph:
  requires: [api-client, archive-fetcher, types]
  provides: [datasource, date-routing]
  affects: [data-fetching]
tech_stack:
  added: []
  patterns: [date-based-routing, graceful-nil-handling]
key_files:
  created:
    - wakadash/internal/datasource/source.go
    - wakadash/internal/datasource/source_test.go
  modified: []
decisions:
  - "7-day boundary for recent vs archive (7 days ago = recent, 8 days ago = archive)"
  - "API returns SummaryResponse with multiple days - extract single matching day"
  - "Nil archive fetcher returns (nil, nil) not error - enables zero-config operation"
metrics:
  duration_seconds: 197
  duration_human: "3 minutes 17 seconds"
  tasks_completed: 3
  files_created: 2
  commits: 4
  test_lines: 167
  completed_date: "2026-02-25"
---

# Phase 13 Plan 01: Date-based Data Source Routing

**One-liner:** Date-based hybrid routing using API for recent dates (≤7 days) and GitHub archive for historical data (>7 days)

## Overview

Created the `datasource` package that provides a unified interface for fetching WakaTime data from either the API (recent dates) or GitHub archive (older dates). The routing decision is based purely on date recency, not data availability.

## Implementation Summary

### TDD Workflow

Followed strict red-green-refactor cycle:

1. **RED**: Created failing test for `IsRecent` date logic
2. **GREEN**: Implemented `IsRecent` to make tests pass
3. **RED**: Added failing tests for `Fetch` routing and `extractDay` helper
4. **GREEN**: Implemented `Fetch` and `extractDay` to make all tests pass

### Key Components

**DataSource struct:**
- Holds references to `api.Client` and `archive.Fetcher`
- Routes requests based on date recency
- Handles nil archive fetcher gracefully

**IsRecent(date string) bool:**
- Returns true if date is within 7 days of today (inclusive)
- Boundary: 7 days ago = recent, 8 days ago = not recent
- Handles invalid date formats (returns false)

**Fetch(date string) (*types.DayData, error):**
- Routes to API for recent dates (calls `FetchSummary(7)`)
- Routes to archive for old dates (calls `FetchArchive(date)`)
- Returns `(nil, nil)` for old dates when archive fetcher is nil

**extractDay(summary, date) *types.DayData:**
- Filters API's `SummaryResponse.Data[]` to single matching day
- Returns nil if date not found in response

## Test Coverage

Comprehensive behavioral tests (167 lines):

- **IsRecent tests:** Today, 1 day ago, 7 days ago (boundary), 8 days ago (boundary), 30 days ago
- **Fetch routing tests:** Recent date uses API, old date uses archive
- **Nil handling test:** Old date with nil fetcher returns (nil, nil)
- **extractDay tests:** Finds matching date, returns nil when no match

All tests verify behavior, not implementation details.

## Deviations from Plan

None - plan executed exactly as written. TDD workflow followed rigorously.

## Verification

✅ All verification criteria met:
- `gofmt -l` returns empty (no formatting issues)
- `go build ./wakadash/internal/datasource/...` compiles without errors
- `go test ./wakadash/internal/datasource/... -v` passes all tests

## Success Criteria

- [x] `wakadash/internal/datasource/source.go` exists with DataSource struct
- [x] IsRecent correctly identifies dates within 7-day window
- [x] Fetch routes to API for recent dates, archive for old dates
- [x] All tests pass demonstrating routing behavior
- [x] Code follows existing patterns (same timeout, error handling style)

## Files Created

### wakadash/internal/datasource/source.go
- DataSource struct with api/archive routing
- IsRecent date calculation (7-day window)
- Fetch routing implementation
- extractDay helper for API response filtering

### wakadash/internal/datasource/source_test.go
- Comprehensive tests for IsRecent boundaries
- Behavioral tests for Fetch routing
- Tests for nil fetcher handling
- Tests for extractDay filtering logic

## Commits

| Hash    | Type | Description |
|---------|------|-------------|
| 28b88ec | test | Add failing test for IsRecent date logic |
| eea894d | feat | Implement IsRecent date logic |
| 8b84e83 | test | Add failing tests for Fetch routing logic |
| acd20ce | feat | Implement Fetch routing and extractDay helper |

## Next Steps

This datasource abstraction is ready to be integrated into the dashboard UI to enable historical data viewing. Next plans will:
1. Integrate datasource into the TUI for date navigation
2. Add UI controls for date selection
3. Update dashboard to use hybrid data fetching

## Self-Check: PASSED

**Files created:**
✅ wakadash/internal/datasource/source.go exists
✅ wakadash/internal/datasource/source_test.go exists

**Commits verified:**
✅ 28b88ec - test(13-01): add failing test for IsRecent date logic
✅ eea894d - feat(13-01): implement IsRecent date logic
✅ 8b84e83 - test(13-01): add failing tests for Fetch routing logic
✅ acd20ce - feat(13-01): implement Fetch routing and extractDay helper

**Tests verified:**
✅ All tests pass (6 test functions, 11 test cases)
✅ Package builds without errors
✅ Code formatting is correct
