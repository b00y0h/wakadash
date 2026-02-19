# Roadmap: wakadash

## Milestones

- [x] **v1.0 Homebrew Distribution** - Phases 1-3 (shipped 2026-02-17)
- [ ] **v2.0 wakadash** - Phases 4-7 (in progress)

## Phases

<details>
<summary>v1.0 Homebrew Distribution (Phases 1-3) - SHIPPED 2026-02-17</summary>

See `.planning/milestones/v1.0-ROADMAP.md` for archived phase details.

**Delivered:** Automated Homebrew distribution for wakafetch — users install via `brew tap b00y0h/wakafetch && brew install wakafetch`

</details>

### v2.0 wakadash (In Progress)

**Milestone Goal:** Create a standalone terminal dashboard tool with live updates and rich data visualization, ready for homebrew-core submission.

- [x] **Phase 4: Repository Setup** - Fresh wakadash repo with ported code and release automation (completed 2026-02-19)
- [ ] **Phase 5: TUI Foundation** - Async bubbletea architecture with basic dashboard
- [ ] **Phase 6: Data Visualization and UX** - Charts, resize handling, panel toggles
- [ ] **Phase 7: Distribution** - Homebrew tap and homebrew-core submission

## Phase Details

### Phase 4: Repository Setup
**Goal**: Fresh standalone repository with ported wakafetch code and working release automation
**Depends on**: Nothing (first phase of v2.0)
**Requirements**: REPO-01, REPO-02, REPO-03, REPO-04
**Success Criteria** (what must be TRUE):
  1. User can clone b00y0h/wakadash and build with `go build`
  2. Running `wakadash --help` shows usage (wakafetch functionality ported)
  3. Creating a version tag triggers GitHub Actions release workflow
  4. Release artifacts include darwin/linux x amd64/arm64 binaries with checksums
**Plans:** 2 plans

Plans:
- [x] 04-01-PLAN.md — Create GitHub repository and port wakafetch code via clean rewrite
- [x] 04-02-PLAN.md — Configure GoReleaser and GitHub Actions for tag-triggered releases

### Phase 5: TUI Foundation
**Goal**: Async bubbletea dashboard with basic stats display, keyboard navigation, and proper terminal handling
**Depends on**: Phase 4
**Requirements**: DASH-01, DASH-02, DASH-03, DASH-04, DASH-05
**Success Criteria** (what must be TRUE):
  1. User can launch full-screen dashboard with `wakadash` command (AltScreen mode)
  2. Dashboard fetches and displays coding stats without blocking the UI
  3. Dashboard auto-refreshes at configurable interval (visible countdown or last-updated timestamp)
  4. User can quit with `q` key and terminal restores cleanly
  5. User can view keybinding help with `?` key
**Plans**: TBD

Plans:
- [ ] 05-01: Implement bubbletea model with async fetch architecture
- [ ] 05-02: Add auto-refresh ticker and help overlay

### Phase 6: Data Visualization and UX
**Goal**: Rich data visualization with color-coded charts and responsive terminal handling
**Depends on**: Phase 5
**Requirements**: VIZ-01, VIZ-02, VIZ-03, VIZ-04, UX-01, UX-02, UX-03
**Success Criteria** (what must be TRUE):
  1. Dashboard displays languages bar chart with distinct colors per language
  2. Dashboard displays projects bar chart showing time breakdown
  3. Dashboard displays sparkline showing hourly coding activity pattern
  4. Dashboard displays heatmap panel showing activity over time
  5. Dashboard reflows layout correctly when terminal is resized
  6. Dashboard shows visual indicator and continues working when API rate-limited
  7. User can toggle panel visibility with number keys (1-4)
**Plans**: TBD

Plans:
- [ ] 06-01: Implement bar charts for languages and projects
- [ ] 06-02: Implement sparkline and heatmap panels
- [ ] 06-03: Add resize handling, rate-limit backoff, and panel toggles

### Phase 7: Distribution
**Goal**: Homebrew distribution via personal tap and homebrew-core submission
**Depends on**: Phase 6
**Requirements**: DIST-01, DIST-02, DIST-03, DIST-04
**Success Criteria** (what must be TRUE):
  1. User can install via `brew tap b00y0h/wakadash && brew install wakadash`
  2. Installation works without macOS Gatekeeper warnings
  3. homebrew-core formula PR submitted with source build
  4. Formula appears on brew.sh after homebrew-core acceptance
**Plans**: TBD

Plans:
- [ ] 07-01: Create personal Homebrew tap with working cask
- [ ] 07-02: Submit homebrew-core formula

## Progress

**Execution Order:** Phases execute in numeric order: 4 -> 5 -> 6 -> 7

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1-3 | v1.0 | 6/6 | Complete | 2026-02-17 |
| 4. Repository Setup | v2.0 | 2/2 | Complete | 2026-02-19 |
| 5. TUI Foundation | v2.0 | 0/2 | Not started | - |
| 6. Data Viz + UX | v2.0 | 0/3 | Not started | - |
| 7. Distribution | v2.0 | 0/2 | Not started | - |

---
*Roadmap created: 2026-02-18*
*Last updated: 2026-02-19*
