# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-19)

**Core value:** A beautiful, live-updating terminal dashboard for coding stats — like htop for your coding activity
**Current focus:** v2.1 Visual Overhaul + Themes

## Current Position

Phase: Not started (defining requirements)
Plan: —
Status: Defining requirements
Last activity: 2026-02-19 — Milestone v2.1 started

## Performance Metrics

**Velocity (v1.0):**
- Total plans completed: 6
- Average duration: 2.5 min
- Total execution time: 0.25 hours

**Velocity (v2.0):**
- Total plans completed: 9
- Average duration: 9.8 min
- Total execution time: 1.5 hours

**By Phase (v2.0):**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 4 | 2/2 | 20min | 10min |
| 5 | 2/2 | 8min | 4min |
| 6 | 3/3 | 18min | 6min |
| 7 | 2/2 | 44min | 22min |

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
- Use personal tap (b00y0h/wakadash) before homebrew-core submission
- Add quarantine removal hook for unsigned binaries to prevent Gatekeeper warnings
- Use fine-grained PAT with Contents permission for tap repository writes
- Version bump to v0.2.0 to represent Homebrew distribution capability
- Submit early to homebrew-core with transparency note about popularity thresholds
- Use std_go_args helper for Go source builds in homebrew-core formula
- Target main branch (not master) for homebrew-core PRs

### Pending Todos

None.

### Blockers/Concerns

**homebrew-core resubmission:** Formula ready at b00y0h/homebrew-core:wakadash. Resubmit when project reaches ≥30 forks, ≥30 watchers, or ≥75 stars.

## Session Continuity

Last session: 2026-02-19
Stopped at: v2.0 MILESTONE COMPLETE - All 4 phases shipped
Resume file: None
