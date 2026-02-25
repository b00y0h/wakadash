# Roadmap: wakadash

## Milestones

- **v1.0 Homebrew Distribution** - Phases 1-3 (shipped 2026-02-17)
- **v2.0 wakadash** - Phases 4-7 (shipped 2026-02-19)
- **v2.1 Visual Overhaul + Themes** - Phases 8-10 (shipped 2026-02-23)
- **v2.2 Historical Data** - Phases 11-15 (in progress)

## Phases

<details>
<summary>v1.0 Homebrew Distribution (Phases 1-3) - SHIPPED 2026-02-17</summary>

See `.planning/milestones/v1.0-ROADMAP.md` for archived phase details.

**Delivered:** Automated Homebrew distribution for wakafetch — users install via `brew tap b00y0h/wakafetch && brew install wakafetch`

</details>

<details>
<summary>v2.0 wakadash (Phases 4-7) - SHIPPED 2026-02-19</summary>

- [x] Phase 4: Repository Setup (2/2 plans) — completed 2026-02-19
- [x] Phase 5: TUI Foundation (2/2 plans) — completed 2026-02-19
- [x] Phase 6: Data Viz + UX (3/3 plans) — completed 2026-02-19
- [x] Phase 7: Distribution (2/2 plans) — completed 2026-02-19

**Delivered:** Full-featured terminal dashboard for WakaTime/Wakapi with async data fetching, visualization panels, and Homebrew distribution

</details>

<details>
<summary>v2.1 Visual Overhaul + Themes (Phases 8-10) - SHIPPED 2026-02-23</summary>

- [x] Phase 8: Theme Foundation (3/3 plans) — completed 2026-02-20
- [x] Phase 9: Stats Panels + Summary (3/3 plans) — completed 2026-02-20
- [x] Phase 10: Polish + Edge Cases (1/1 plan) — completed 2026-02-20

**Delivered:** Comprehensive theme system with 6 presets, expanded stats panels, and responsive layout

</details>

### v2.2 Historical Data (In Progress)

**Milestone Goal:** Enable viewing coding stats from any historical date by reading archived data from a GitHub repository.

**Phase Numbering:**
- Integer phases (11-15): Planned milestone work
- Decimal phases (11.1, 11.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 11: Configuration & Validation** - Add history_repo config with graceful fallback (completed 2026-02-24)
- [x] **Phase 12: GitHub Archive Integration** - Read archived stats from GitHub repo
- [x] **Phase 13: Hybrid Data Fetching** - Combine API and archive for seamless experience (completed 2026-02-25)
- [ ] **Phase 14: Date Navigation** - Week-based navigation with auto-skip blank weeks (gap closure in progress)
- [ ] **Phase 15: Historical Display** - Date indicator and auto-refresh control

## Phase Details

### Phase 11: Configuration & Validation
**Goal**: User can specify archive location in config
**Depends on**: Phase 10 (v2.1 complete)
**Requirements**: CFG-01
**Success Criteria** (what must be TRUE):
  1. User can add `history_repo` key to ~/.wakatime.cfg
  2. Dashboard starts successfully when `history_repo` is not configured
  3. Dashboard starts successfully when `history_repo` is configured but invalid
**Plans**: 1 plan

Plans:
- [x] 11-01-PLAN.md — Add history_repo config with section-aware parsing and auto-template

### Phase 12: GitHub Archive Integration
**Goal**: Read archived WakaTime data from GitHub
**Depends on**: Phase 11
**Requirements**: DATA-03, DATA-04
**Success Criteria** (what must be TRUE):
  1. Dashboard fetches archived data from GitHub when `history_repo` is configured
  2. Dashboard shows "no data available" when archive file missing (no crash)
  3. Archive data parses correctly and populates all panels
**Plans**: 2 plans

Plans:
- [x] 12-01-PLAN.md — Create GitHub archive fetcher with graceful 404 handling
- [x] 12-02-PLAN.md — Integrate archive fetcher into dashboard

### Phase 13: Hybrid Data Fetching
**Goal**: Seamlessly combine API and archive data
**Depends on**: Phase 12
**Requirements**: DATA-01, DATA-02
**Success Criteria** (what must be TRUE):
  1. Recent dates (last 7 days) fetch from WakaTime API
  2. Older dates fetch from GitHub archive when available
  3. User sees stats from any date without knowing the source
**Plans**: 2 plans

Plans:
- [x] 13-01-PLAN.md — Create DataSource with date-based routing logic (TDD)
- [x] 13-02-PLAN.md — Integrate DataSource into dashboard initialization

### Phase 14: Date Navigation
**Goal**: User can browse historical weeks with auto-skip for blank data
**Depends on**: Phase 13
**Requirements**: NAV-01, NAV-02, NAV-03
**Success Criteria** (what must be TRUE):
  1. Left arrow key navigates to previous week (Sunday-Saturday boundary)
  2. Right arrow key navigates to next week (capped at current week)
  3. Pressing '0' or Home key returns to current week
  4. Navigation auto-skips blank weeks to next week with data
  5. Status bar shows week range and end-of-history indicator
**Plans**: 3 plans

Plans:
- [x] 14-01-PLAN.md — Add day-based navigation controls (superseded by gap closure)
- [ ] 14-02-PLAN.md — Convert to week-based navigation with week range display (gap closure)
- [ ] 14-03-PLAN.md — Add auto-skip blank weeks and end-of-history indicator (gap closure)

### Phase 15: Historical Display
**Goal**: User knows when viewing historical data
**Depends on**: Phase 14
**Requirements**: DISP-01, DISP-02, DISP-03
**Success Criteria** (what must be TRUE):
  1. Date indicator appears when viewing non-today date
  2. Auto-refresh pauses when viewing historical data
  3. Auto-refresh resumes when returning to today
**Plans**: TBD

Plans:
- TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 11 → 11.1 → 11.2 → 12 → 12.1 → 13 → ...

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1-3 | v1.0 | 6/6 | Complete | 2026-02-17 |
| 4. Repository Setup | v2.0 | 2/2 | Complete | 2026-02-19 |
| 5. TUI Foundation | v2.0 | 2/2 | Complete | 2026-02-19 |
| 6. Data Viz + UX | v2.0 | 3/3 | Complete | 2026-02-19 |
| 7. Distribution | v2.0 | 2/2 | Complete | 2026-02-19 |
| 8. Theme Foundation | v2.1 | 3/3 | Complete | 2026-02-20 |
| 9. Stats Panels + Summary | v2.1 | 3/3 | Complete | 2026-02-20 |
| 10. Polish + Edge Cases | v2.1 | 1/1 | Complete | 2026-02-20 |
| 11. Configuration & Validation | v2.2 | 1/1 | Complete | 2026-02-24 |
| 12. GitHub Archive Integration | v2.2 | 2/2 | Complete | 2026-02-25 |
| 13. Hybrid Data Fetching | v2.2 | 2/2 | Complete | 2026-02-25 |
| 14. Date Navigation | 2/3 | In Progress|  | - |
| 15. Historical Display | v2.2 | 0/? | Not started | - |

## Distribution Notes

**Personal Tap (Working):** `brew tap b00y0h/wakadash && brew install wakadash`

**homebrew-core (Pending):** PR #268434 submitted and closed — requires >= 30 forks, >= 30 watchers, or >= 75 stars. Formula ready at b00y0h/homebrew-core:wakadash for resubmission when thresholds met.

---
*Roadmap created: 2026-02-18*
*Last updated: 2026-02-25 (Phase 14 gap closure plans created)*
