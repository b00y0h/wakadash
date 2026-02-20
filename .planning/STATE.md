# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-19)

**Core value:** A beautiful, live-updating terminal dashboard for coding stats — like htop for your coding activity
**Current focus:** v2.1 Visual Overhaul + Themes - Phase 8 (Theme Foundation)

## Current Position

Phase: 10 of 10 (Polish and Edge Cases)
Plan: 1 of 1
Status: Complete
Last activity: 2026-02-20 — Completed 10-01 (Edge Case Hardening)

Progress: [██████████████████] 100% (15 of 15 total plans complete)

## Performance Metrics

**Velocity (v1.0):**
- Total plans completed: 6
- Average duration: 2.5 min
- Total execution time: 0.25 hours

**Velocity (v2.0):**
- Total plans completed: 15
- Average duration: 7.0 min
- Total execution time: 1.74 hours

**By Phase (v2.0):**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 4 | 2/2 | 20min | 10min |
| 5 | 2/2 | 8min | 4min |
| 6 | 3/3 | 18min | 6min |
| 7 | 2/2 | 44min | 22min |
| 8 | 3/3 | 9min | 3min |
| 9 | 3/4 | 9min | 3min |
| 10 | 1/1 | 2min | 2min |

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

**v2.0 foundation decisions:**
- Create standalone wakadash: Enables homebrew-core, adds unique value
- Dashboard mode with live updates: Differentiates from simple fetch tools
- wakadash repo at /workspace/wakadash (root not writable)
- Use tea.WithAltScreen() ProgramOption (not EnterAltScreen command) to avoid race conditions
- Self-loop ticker pattern to avoid ticker drift
- Use GitHub Linguist colors for language bars and contribution colors for heatmap
- Use cenkalti/backoff/v5 for exponential backoff with jitter (1s-30s, 2min max)
- Minimum terminal size: 40x10 to prevent broken layouts
- Version at v0.2.0 representing Homebrew distribution capability

**v2.1 context:**
- Theme system must address hardcoded colors in existing code ✓ (completed in 08-02)
- Research flags AdaptiveColor startup hang risk (fixed in BubbleTea v0.27.1+)
- Single /stats API request returns all data (Categories, Editors, OS, Machines)
- ntcharts barchart.Model pattern proven for Languages/Projects, reuse for new panels
- Use hex colors for themes (lipgloss auto-handles terminal downsampling)
- 5-level heatmap gradient per theme (None, Low, Medium, High, VeryHigh)
- Persist theme to ~/.wakatime.cfg (reuses existing config file)
- Theme-aware style functions pattern: func StyleName(t theme.Theme) lipgloss.Style
- Theme loaded from config in NewModel(), falls back to DefaultTheme if not set
- Theme picker pattern: isFirstRun parameter controls cancel behavior (first-run ignores Esc/Q)
- BubbleTea model composition: Model contains ThemePickerModel for modal UI pattern
- Phase 8 (Theme Foundation) complete ✓ — full theme system with picker, persistence, and live preview
- Summary panel uses accent border (theme.Primary) for visual distinction
- Streak calculation limited to 7-day window from heatmap data
- Panel style functions accept (theme, width) for consistent styling
- Stats panels (Categories, Editors, OS, Machines) limit to top 10 items with "Other" aggregation
- formatTimeWithPercent helper for consistent "2h 15m (65%)" label formatting
- All stat panel bars use theme.Primary color for consistent styling
- Responsive layout: 2-column grid at >= 80 cols, vertical stack at 40-79 cols
- Toggle keys 5-9 control new panels, 'a' shows all, 'h' hides all
- Summary panel positioned at top of layout spanning full width
- Terminal size errors show current and required dimensions for clear user guidance
- Theme lookup normalized to lowercase with whitespace trimming for forgiving config parsing
- Division by zero protection in all stats panel percentage calculations

### Pending Todos

None.

### Blockers/Concerns

**homebrew-core resubmission:** Formula ready at b00y0h/homebrew-core:wakadash. Resubmit when project reaches ≥30 forks, ≥30 watchers, or ≥75 stars.

**v2.1 integration risks (from research):**
- Hardcoded colors in styles.go must be migrated to theme system ✓ (resolved in 08-02)
- AdaptiveColor terminal detection needs early call in main() to prevent hangs
- Dynamic .Width() styles can cause rendering corruption (use .MaxWidth() instead)

## Session Continuity

Last session: 2026-02-20
Stopped at: Completed 10-01-PLAN.md (Edge Case Hardening) — Phase 10 complete
Resume file: None
