# Requirements: wakadash

**Defined:** 2026-02-24
**Core Value:** A beautiful, live-updating terminal dashboard for coding stats — like htop for your coding activity

## v2.2 Requirements

Requirements for historical data support. Each maps to roadmap phases.

### Configuration

- [x] **CFG-01**: User can specify `history_repo` in ~/.wakatime.cfg

### Data Fetching

- [x] **DATA-01**: User can view stats from any date with archived data
- [x] **DATA-02**: Recent dates (last 7 days) fetch from WakaTime API
- [x] **DATA-03**: Older dates fetch from GitHub archive at `history_repo`
- [x] **DATA-04**: Missing archive data shows "no data" gracefully (no error)

### Navigation

- [ ] **NAV-01**: User can navigate to previous day with left arrow
- [ ] **NAV-02**: User can navigate to next day with right arrow
- [ ] **NAV-03**: User can return to today (e.g., 'Home' or '0' key)

### Display

- [ ] **DISP-01**: Date indicator appears when viewing historical data
- [ ] **DISP-02**: Auto-refresh pauses when viewing historical data
- [ ] **DISP-03**: Auto-refresh resumes when returning to today

## Future Requirements

### Navigation Enhancements

- **NAV-04**: User can jump by week (Shift+arrows)
- **NAV-05**: User can jump by month (Ctrl+arrows)
- **NAV-06**: User can open date picker modal

### Display Enhancements

- **DISP-04**: Show "auto-refresh paused" indicator when viewing history

### Other (from v2.1)

- **THEME-05**: Runtime theme switching without restart
- **THEME-06**: Custom themes via config file
- **ADV-02**: Time range switcher at runtime
- **POL-03**: Shell completions (bash, zsh, fish)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Date range selection | Single-day view is simpler; ranges add complexity |
| Comparison view (today vs history) | Requires dual data fetch and split UI |
| Caching historical data | Archive is already cached on GitHub |
| Offline mode | Requires local storage, out of scope for TUI |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| CFG-01 | Phase 11 | Complete |
| DATA-01 | Phase 13 | Complete |
| DATA-02 | Phase 13 | Complete |
| DATA-03 | Phase 12 | ✅ Complete |
| DATA-04 | Phase 12 | ✅ Complete |
| NAV-01 | Phase 14 | Pending |
| NAV-02 | Phase 14 | Pending |
| NAV-03 | Phase 14 | Pending |
| DISP-01 | Phase 15 | Pending |
| DISP-02 | Phase 15 | Pending |
| DISP-03 | Phase 15 | Pending |

**Coverage:**
- v2.2 requirements: 11 total
- Mapped to phases: 11
- Unmapped: 0

✓ All v2.2 requirements mapped to phases

---
*Requirements defined: 2026-02-24*
*Last updated: 2026-02-24 after roadmap creation*
