# Requirements: wakadash

**Defined:** 2026-02-17
**Core Value:** A beautiful, live-updating terminal dashboard for coding stats — like htop for your coding activity

## v2.0 Requirements

Requirements for the wakadash milestone. Each maps to roadmap phases.

### Dashboard Foundation

- [ ] **DASH-01**: User can launch persistent full-screen dashboard with `wakadash` command
- [ ] **DASH-02**: Dashboard auto-refreshes at configurable interval (default 60s)
- [ ] **DASH-03**: User can quit dashboard with `q` key
- [ ] **DASH-04**: User can view help overlay with `?` key
- [ ] **DASH-05**: Dashboard uses async architecture (all API calls via tea.Cmd)

### Data Visualization

- [ ] **VIZ-01**: Dashboard displays languages bar chart with color coding
- [ ] **VIZ-02**: Dashboard displays projects bar chart with time breakdown
- [ ] **VIZ-03**: Dashboard displays sparkline showing hourly activity (requires durations API)
- [ ] **VIZ-04**: Dashboard displays heatmap panel showing activity over time

### User Experience

- [ ] **UX-01**: Dashboard responds to terminal resize events
- [ ] **UX-02**: Dashboard handles API rate limits with exponential backoff and visual indicator
- [ ] **UX-03**: User can toggle panel visibility with number keys (1-4)

### Repository Setup

- [ ] **REPO-01**: Fresh wakadash repository created (b00y0h/wakadash, not a fork)
- [ ] **REPO-02**: Wakafetch code ported to new repo with proper attribution
- [ ] **REPO-03**: GoReleaser configured for multi-platform builds
- [ ] **REPO-04**: GitHub Actions workflow for tag-triggered releases

### Distribution

- [ ] **DIST-01**: Personal Homebrew tap (b00y0h/homebrew-wakadash) with working cask
- [ ] **DIST-02**: User can install via `brew tap b00y0h/wakadash && brew install wakadash`
- [ ] **DIST-03**: homebrew-core formula submitted (builds from source)
- [ ] **DIST-04**: Formula accepted to homebrew-core (appears on brew.sh)

## v3 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Advanced Features

- **ADV-01**: Color theme customization
- **ADV-02**: Time range switcher at runtime (today/week/month)
- **ADV-03**: Goals/streak display (if API supports)
- **ADV-04**: Multiple WakaTime/Wakapi account support

### Polish

- **POL-01**: Mouse support for panel interaction
- **POL-02**: Configuration file for persistent settings
- **POL-03**: Shell completions (bash, zsh, fish)

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Windows builds | Homebrew is macOS/Linux focused; Windows users use Go install |
| Real-time WebSocket | WakaTime doesn't offer WebSocket API; polling sufficient |
| Web interface | Terminal-only tool; web dashboard exists on wakatime.com |
| Per-file granularity | Complexity vs value; project-level sufficient |
| In-dashboard config editing | Defer to config file in v3 |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| REPO-01 | Phase 4 | Pending |
| REPO-02 | Phase 4 | Pending |
| REPO-03 | Phase 4 | Pending |
| REPO-04 | Phase 4 | Pending |
| DASH-01 | Phase 5 | Pending |
| DASH-02 | Phase 5 | Pending |
| DASH-03 | Phase 5 | Pending |
| DASH-04 | Phase 5 | Pending |
| DASH-05 | Phase 5 | Pending |
| VIZ-01 | Phase 6 | Pending |
| VIZ-02 | Phase 6 | Pending |
| VIZ-03 | Phase 6 | Pending |
| VIZ-04 | Phase 6 | Pending |
| UX-01 | Phase 6 | Pending |
| UX-02 | Phase 6 | Pending |
| UX-03 | Phase 6 | Pending |
| DIST-01 | Phase 7 | Pending |
| DIST-02 | Phase 7 | Pending |
| DIST-03 | Phase 7 | Pending |
| DIST-04 | Phase 7 | Pending |

**Coverage:**
- v2.0 requirements: 20 total
- Mapped to phases: 20
- Unmapped: 0 ✓

---
*Requirements defined: 2026-02-17*
*Last updated: 2026-02-17 after definition*
