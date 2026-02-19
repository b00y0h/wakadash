# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-17)

**Core value:** A beautiful, live-updating terminal dashboard for coding stats — like htop for your coding activity
**Current focus:** Phase 6 - Data Visualization and UX

## Current Position

Phase: 6 of 7 (Data Visualization and UX)
Plan: 3 of 3 in current phase
Status: Complete
Last activity: 2026-02-19 — 06-03 complete, panel toggles, rate limiting, and resize handling implemented

Progress: [███████░░░] 63% (v2.0)

## Performance Metrics

**Velocity (v1.0):**
- Total plans completed: 6
- Average duration: 2.5 min
- Total execution time: 0.25 hours

**Velocity (v2.0):**
- Total plans completed: 7
- Average duration: 6.4 min
- Total execution time: 0.75 hours

**By Phase (v2.0):**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 4 | 2/2 | 20min | 10min |
| 5 | 2/2 | 8min | 4min |
| 6 | 3/3 | 18min | 6min |
| 7 | 0/2 | - | - |

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Create standalone wakadash: Enables homebrew-core, adds unique value
- Port and enhance wakafetch: Preserve working API code, build on top
- Dashboard mode with live updates: Differentiates from simple fetch tools
- wakadash repo at /workspace/wakadash (root not writable)
- CGO_ENABLED=0 for fully static binaries (no libc dependency on target systems)
- No brews: in goreleaser config — Homebrew tap is Phase 7 scope
- ldflags target main.version/commit/date (NOT full module path)
- GITHUB_TOKEN sufficient for Phase 4 release workflow
- Use tea.WithAltScreen() ProgramOption (not EnterAltScreen command) to avoid race conditions
- Initialize width=80, height=24 in NewModel() to prevent blank/panicking first render before WindowSizeMsg
- Include recover() in fetchStatsCmd to prevent terminal corruption if API client panics
- Self-loop ticker pattern (scheduleRefresh fires once, statsFetchedMsg schedules next) to avoid ticker drift
- Schedule refresh only from statsFetchedMsg/fetchErrMsg handlers to prevent double-ticker bug
- Implement help.KeyMap interface for bubbles/help auto-generation
- Use simplified heatmap: daily totals from summaries instead of hourly durations per day
- Fetch durations for today only to minimize API calls
- Use GitHub Linguist colors for language bars and contribution colors for heatmap
- Display 7-day heatmap as colored blocks with MM-DD labels
- Use cenkalti/backoff/v5 for exponential backoff with jitter (1s-30s, 2min max)
- Panel visibility as bool flags (not map) for type safety
- Minimum terminal size: 40x10 to prevent broken layouts

### Pending Todos

None.

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-02-19
Stopped at: Completed 06-03-PLAN.md - responsive UX and resilience features (Phase 6 complete)
Resume file: None
