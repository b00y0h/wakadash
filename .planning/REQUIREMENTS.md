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

- [x] **NAV-01**: User can navigate to previous day with left arrow
- [x] **NAV-02**: User can navigate to next day with right arrow
- [x] **NAV-03**: User can return to today (e.g., 'Home' or '0' key)

### Display

- [x] **DISP-01**: Date indicator appears when viewing historical data
- [x] **DISP-02**: Auto-refresh pauses when viewing historical data
- [x] **DISP-03**: Auto-refresh resumes when returning to today

### Prefetch (Phase 17)

- [x] **PREFETCH-01**: Previous week data prefetches silently after dashboard loads
- [x] **PREFETCH-02**: Backward navigation is instant when data is prefetched

### No-Data UX (Phase 17)

- [x] **NODATA-01**: Full-screen "End of history" banner appears when no data exists
- [x] **NODATA-02**: Banner shows navigation hints to return to current week

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
| DATA-01 | Phase 16 | Complete |
| DATA-02 | Phase 13 | Complete |
| DATA-03 | Phase 12 | Complete |
| DATA-04 | Phase 12 | Complete |
| NAV-01 | Phase 14 | Complete |
| NAV-02 | Phase 14 | Complete |
| NAV-03 | Phase 14 | Complete |
| DISP-01 | Phase 16 | Complete |
| DISP-02 | Phase 16 | Complete |
| DISP-03 | Phase 16 | Complete |
| PREFETCH-01 | Phase 17 | Complete |
| PREFETCH-02 | Phase 17 | Complete |
| NODATA-01 | Phase 17 | Complete |
| NODATA-02 | Phase 17 | Complete |

**Coverage:**
- v2.2 requirements: 15 total
- Complete: 11 (CFG-01, DATA-01-04, NAV-01-03, DISP-01-03)
- Pending (Phase 17): 4 (PREFETCH-01-02, NODATA-01-02)
- Unmapped: 0

All v2.2 requirements mapped to phases

---
*Requirements defined: 2026-02-24*
*Last updated: 2026-02-25 (Phase 17 requirements added)*
