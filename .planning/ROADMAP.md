# Roadmap: wakadash

## Milestones

- ✅ **v1.0 Homebrew Distribution** - Phases 1-3 (shipped 2026-02-17)
- ✅ **v2.0 wakadash** - Phases 4-7 (shipped 2026-02-19)
- 🚧 **v2.1 Visual Overhaul + Themes** - Phases 8-10 (in progress)

## Phases

<details>
<summary>✅ v1.0 Homebrew Distribution (Phases 1-3) - SHIPPED 2026-02-17</summary>

See `.planning/milestones/v1.0-ROADMAP.md` for archived phase details.

**Delivered:** Automated Homebrew distribution for wakafetch — users install via `brew tap b00y0h/wakafetch && brew install wakafetch`

</details>

<details>
<summary>✅ v2.0 wakadash (Phases 4-7) - SHIPPED 2026-02-19</summary>

### Phase 4: Repository Setup
**Goal**: Fresh standalone repository with ported wakafetch code and working release automation
**Requirements**: REPO-01, REPO-02, REPO-03, REPO-04
**Success Criteria** (what must be TRUE):
  1. User can clone b00y0h/wakadash and build with `go build`
  2. Running `wakadash --help` shows usage (wakafetch functionality ported)
  3. Creating a version tag triggers GitHub Actions release workflow
  4. Release artifacts include darwin/linux x amd64/arm64 binaries with checksums
**Plans**: 2 plans

Plans:
- [x] 04-01: Create GitHub repository and port wakafetch code
- [x] 04-02: Configure GoReleaser and GitHub Actions

### Phase 5: TUI Foundation
**Goal**: Async bubbletea dashboard with basic stats display, keyboard navigation, and proper terminal handling
**Requirements**: DASH-01, DASH-02, DASH-03, DASH-04, DASH-05
**Success Criteria** (what must be TRUE):
  1. User can launch full-screen dashboard with `wakadash` command
  2. Dashboard fetches and displays coding stats without blocking the UI
  3. Dashboard auto-refreshes at configurable interval
  4. User can quit with `q` key and terminal restores cleanly
  5. User can view keybinding help with `?` key
**Plans**: 2 plans

Plans:
- [x] 05-01: TUI core with async fetch architecture
- [x] 05-02: Auto-refresh and help overlay

### Phase 6: Data Visualization and UX
**Goal**: Rich data visualization with color-coded charts and responsive terminal handling
**Requirements**: VIZ-01, VIZ-02, VIZ-03, VIZ-04, UX-01, UX-02, UX-03
**Success Criteria** (what must be TRUE):
  1. Dashboard displays languages bar chart with distinct colors per language
  2. Dashboard displays projects bar chart showing time breakdown
  3. Dashboard displays sparkline showing hourly coding activity pattern
  4. Dashboard displays heatmap panel showing activity over time
  5. Dashboard reflows layout correctly when terminal is resized
  6. Dashboard shows visual indicator and continues working when API rate-limited
  7. User can toggle panel visibility with number keys (1-4)
**Plans**: 3 plans

Plans:
- [x] 06-01: Bar charts for languages and projects
- [x] 06-02: Sparkline and heatmap
- [x] 06-03: Panel toggles and resize handling

### Phase 7: Distribution
**Goal**: Homebrew distribution via personal tap and homebrew-core submission
**Requirements**: DIST-01, DIST-02, DIST-03, DIST-04
**Success Criteria** (what must be TRUE):
  1. User can install via `brew tap b00y0h/wakadash && brew install wakadash`
  2. Installation works without macOS Gatekeeper warnings
  3. homebrew-core formula PR submitted with source build
  4. Formula appears on brew.sh after homebrew-core acceptance
**Plans**: 2 plans

Plans:
- [x] 07-01: Personal Homebrew tap
- [x] 07-02: homebrew-core submission

</details>

### 🚧 v2.1 Visual Overhaul + Themes (In Progress)

**Milestone Goal:** Enhance dashboard with comprehensive stats panels matching wakafetch visual style, plus configurable color themes

#### Phase 8: Theme Foundation
**Goal**: Users can select and persist color themes
**Depends on**: Phase 7
**Requirements**: THEME-01, THEME-02, THEME-03, THEME-04
**Success Criteria** (what must be TRUE):
  1. User can choose from 6 theme presets (Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night)
  2. User sees visual theme preview on first run before selecting
  3. User's theme selection persists across restarts in ~/.wakatime.cfg
  4. All existing panels (Languages, Projects, Heatmap) use the selected theme colors consistently
**Plans**: 3 plans

Plans:
- [ ] 08-01-PLAN.md - Theme package with struct, 6 presets, and config persistence
- [ ] 08-02-PLAN.md - Theme-aware styles migration and Model integration
- [ ] 08-03-PLAN.md - Full-screen theme picker with preview and first-run detection

#### Phase 9: Stats Panels + Summary
**Goal**: Dashboard displays comprehensive stats with responsive layout
**Depends on**: Phase 8
**Requirements**: STAT-01, STAT-02, STAT-03, STAT-04, STAT-05, LAYOUT-01, LAYOUT-02
**Success Criteria** (what must be TRUE):
  1. User sees Categories, Editors, Operating Systems, and Machines panels with top 10 items and time labels
  2. User sees Summary panel showing Last 30d total, daily avg, top items, and counts
  3. Dashboard panels arrange in 2-column layout on terminals ≥80 cols, stacking on smaller terminals
  4. User can toggle each panel's visibility with keyboard shortcuts
  5. All panels use consistent visual styling with selected theme
**Plans**: TBD

Plans:
- [ ] 09-01: TBD

#### Phase 10: Polish + Edge Cases
**Goal**: Dashboard handles edge cases gracefully
**Depends on**: Phase 9
**Requirements**: (polish phase - no specific requirements)
**Success Criteria** (what must be TRUE):
  1. User sees friendly error message when terminal is too small for dashboard
  2. Invalid theme names fallback to default theme gracefully
  3. Dashboard handles missing data categories without crashing
**Plans**: TBD

Plans:
- [ ] 10-01: TBD

## Progress

**Execution Order:** Phases execute in numeric order: 8 → 9 → 10

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1-3 | v1.0 | 6/6 | Complete | 2026-02-17 |
| 4. Repository Setup | v2.0 | 2/2 | Complete | 2026-02-19 |
| 5. TUI Foundation | v2.0 | 2/2 | Complete | 2026-02-19 |
| 6. Data Viz + UX | v2.0 | 3/3 | Complete | 2026-02-19 |
| 7. Distribution | v2.0 | 2/2 | Complete | 2026-02-19 |
| 8. Theme Foundation | v2.1 | 0/TBD | Not started | - |
| 9. Stats Panels + Summary | v2.1 | 0/TBD | Not started | - |
| 10. Polish + Edge Cases | v2.1 | 0/TBD | Not started | - |

## Distribution Notes

**Personal Tap (Working):** `brew tap b00y0h/wakadash && brew install wakadash`

**homebrew-core (Pending):** PR #268434 submitted and closed — requires ≥30 forks, ≥30 watchers, or ≥75 stars. Formula ready at b00y0h/homebrew-core:wakadash for resubmission when thresholds met.

---
*Roadmap created: 2026-02-18*
*Last updated: 2026-02-19 (v2.1 phases added)*
