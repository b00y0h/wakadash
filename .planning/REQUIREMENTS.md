# Requirements: wakadash

**Defined:** 2026-02-17
**Core Value:** A beautiful, live-updating terminal dashboard for coding stats — like htop for your coding activity

## v2.0 Requirements (Completed)

Requirements for the wakadash milestone. All shipped in v2.0.

### Repository Setup

- [x] **REPO-01**: Fresh wakadash repository created (b00y0h/wakadash, not a fork)
- [x] **REPO-02**: Wakafetch code ported to new repo with proper attribution
- [x] **REPO-03**: GoReleaser configured for multi-platform builds
- [x] **REPO-04**: GitHub Actions workflow for tag-triggered releases

### Dashboard Foundation

- [x] **DASH-01**: User can launch persistent full-screen dashboard with `wakadash` command
- [x] **DASH-02**: Dashboard auto-refreshes at configurable interval (default 60s)
- [x] **DASH-03**: User can quit dashboard with `q` key
- [x] **DASH-04**: User can view help overlay with `?` key
- [x] **DASH-05**: Dashboard uses async architecture (all API calls via tea.Cmd)

### Data Visualization

- [x] **VIZ-01**: Dashboard displays languages bar chart with color coding
- [x] **VIZ-02**: Dashboard displays projects bar chart with time breakdown
- [x] **VIZ-03**: Dashboard displays sparkline showing hourly activity (requires durations API)
- [x] **VIZ-04**: Dashboard displays heatmap panel showing activity over time

### User Experience

- [x] **UX-01**: Dashboard responds to terminal resize events
- [x] **UX-02**: Dashboard handles API rate limits with exponential backoff and visual indicator
- [x] **UX-03**: User can toggle panel visibility with number keys (1-4)

### Distribution

- [x] **DIST-01**: Personal Homebrew tap (b00y0h/homebrew-wakadash) with working cask
- [x] **DIST-02**: User can install via `brew tap b00y0h/wakadash && brew install wakadash`
- [x] **DIST-03**: homebrew-core formula submitted (builds from source)
- [ ] **DIST-04**: Formula accepted to homebrew-core (appears on brew.sh) — deferred pending popularity

## v2.1 Requirements

Requirements for v2.1 Visual Overhaul + Themes milestone. Each maps to roadmap phases.

### Stats Panels

- [ ] **STAT-01**: User sees Categories panel with top 10 categories as horizontal bars with time labels
- [ ] **STAT-02**: User sees Editors panel with top 10 editors as horizontal bars with time labels
- [ ] **STAT-03**: User sees Operating Systems panel with top 10 OS as horizontal bars with time labels
- [ ] **STAT-04**: User sees Machines panel with top 10 machines as horizontal bars with time labels
- [ ] **STAT-05**: User sees Summary panel showing Last 30d total, daily avg, top project/editor/category/OS, language count, project count

### Theme System

- [ ] **THEME-01**: User can choose from 6 theme presets (Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night)
- [ ] **THEME-02**: User sees visual theme preview on first run before selecting
- [ ] **THEME-03**: User's theme selection persists in ~/.wakatime.cfg
- [ ] **THEME-04**: All existing panels (Languages, Projects, Heatmap) use the selected theme colors

### Layout

- [ ] **LAYOUT-01**: Dashboard panels arrange in responsive 2-column layout on terminals ≥80 cols
- [ ] **LAYOUT-02**: User can toggle each panel's visibility with keyboard shortcuts

## Future Requirements

Deferred to v2.2+. Tracked but not in current roadmap.

### Advanced Theming

- **THEME-05**: User can switch themes at runtime without restart
- **THEME-06**: User can create custom themes via config file

### Layout Enhancements

- **LAYOUT-03**: Dashboard uses 3-column layout on ultra-wide terminals (≥120 cols)
- **LAYOUT-04**: User can rearrange panel order via config

### Advanced Features

- **ADV-02**: Time range switcher at runtime (today/week/month)
- **ADV-03**: Goals/streak display (if API supports)
- **ADV-04**: Multiple WakaTime/Wakapi account support

### Polish

- **POL-01**: Mouse support for panel interaction
- **POL-03**: Shell completions (bash, zsh, fish)

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Windows builds | Homebrew is macOS/Linux focused; Windows users use Go install |
| Real-time WebSocket | WakaTime doesn't offer WebSocket API; polling sufficient |
| Web interface | Terminal-only tool; web dashboard exists on wakatime.com |
| Per-file granularity | Complexity vs value; project-level sufficient |
| Custom theme editor in TUI | Complex UI, error-prone, config file editing simpler |
| Animated theme transitions | Terminal rendering limitations, flicker, poor UX |
| Auto-theme by time of day | Assumes user preferences, hard to debug |
| 100+ theme pack | Maintenance burden, decision paralysis; 6 curated themes optimal |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| STAT-01 | TBD | Pending |
| STAT-02 | TBD | Pending |
| STAT-03 | TBD | Pending |
| STAT-04 | TBD | Pending |
| STAT-05 | TBD | Pending |
| THEME-01 | TBD | Pending |
| THEME-02 | TBD | Pending |
| THEME-03 | TBD | Pending |
| THEME-04 | TBD | Pending |
| LAYOUT-01 | TBD | Pending |
| LAYOUT-02 | TBD | Pending |

**Coverage:**
- v2.1 requirements: 11 total
- Mapped to phases: 0
- Unmapped: 11

---
*Requirements defined: 2026-02-17*
*v2.1 requirements added: 2026-02-19*
