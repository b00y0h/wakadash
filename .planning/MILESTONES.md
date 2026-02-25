# Milestones

## v1.0 Homebrew Distribution (Shipped: 2026-02-17)

**Phases completed:** 3 phases, 6 plans
**Timeline:** 5 days (2026-02-13 → 2026-02-17)

**Delivered:** Automated Homebrew distribution for wakafetch — users install via `brew tap b00y0h/wakafetch && brew install wakafetch`

**Key accomplishments:**
- Repository infrastructure: Forked sahaj-b/wakafetch, created tap b00y0h/homebrew-wakafetch, configured cross-repo PAT
- GoReleaser automation: Multi-platform builds (darwin/linux × amd64/arm64) with SHA256 checksums
- GitHub Actions workflow: Tag-triggered releases auto-publish to Homebrew tap
- Modern cask distribution: Migrated to homebrew_casks with macOS quarantine removal hook
- Verified end-to-end: v0.1.0 released, `brew tap && brew install` works without security warnings

**Archive:** `.planning/milestones/v1.0-ROADMAP.md`, `.planning/milestones/v1.0-REQUIREMENTS.md`

---

## v2.0 wakadash (Shipped: 2026-02-19)

**Phases completed:** 4 phases, 9 plans
**Timeline:** 2 days (2026-02-18 → 2026-02-19)

**Delivered:** Full-featured terminal dashboard for WakaTime/Wakapi with async data fetching, visualization panels, and Homebrew distribution

**Key accomplishments:**
- Fresh standalone repository: Created b00y0h/wakadash, ported wakafetch code with enhancements
- Full-screen TUI dashboard: BubbleTea-based with async API calls and auto-refresh
- Data visualization: Languages and projects bar charts, hourly activity sparkline, weekly heatmap
- User experience: Keyboard navigation, panel toggles (1-4), help overlay, terminal resize handling
- Rate limit handling: Exponential backoff with visual indicator during API throttling
- Distribution: Personal Homebrew tap with cask (b00y0h/homebrew-wakadash)

**Archive:** `.planning/milestones/v2.0-ROADMAP.md`, `.planning/milestones/v2.0-REQUIREMENTS.md`

---

## v2.1 Visual Overhaul + Themes (Shipped: 2026-02-23)

**Phases completed:** 3 phases, 7 plans
**Timeline:** 1 day (2026-02-20)

**Delivered:** Comprehensive theme system with 6 presets, expanded stats panels, and responsive layout

**Key accomplishments:**
- Complete theme system: 6 official presets (Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night) with config persistence
- Theme picker UI: Full-screen preview with first-run flow and runtime 't' key switching
- Stats panels: Categories, Editors, OS, Machines with top-10 horizontal bars and "Other" aggregation
- Summary panel: 30-day overview with streak calculation and accent styling
- Responsive layout: 2-column grid at >= 80 cols with extended keyboard controls (5-9, a, h)
- Edge case hardening: Terminal size validation, case-insensitive theme lookup, division-by-zero protection

**Archive:** `.planning/milestones/v2.1-ROADMAP.md`, `.planning/milestones/v2.1-REQUIREMENTS.md`, `.planning/milestones/v2.1-MILESTONE-AUDIT.md`

---


## v2.2 Historical Data (Shipped: 2026-02-25)

**Phases completed:** 6 phases, 12 plans
**Timeline:** 2 days (2026-02-24 → 2026-02-25)

**Delivered:** Historical data viewing via GitHub archive with week-based navigation, instant prefetch, and polished end-of-history UX

**Key accomplishments:**
- Configuration system: Section-aware INI parsing with `history_repo` field and auto-template creation for discoverability
- GitHub archive fetcher: Retrieves historical WakaTime data with graceful 404 handling (missing data = nil, not error)
- Hybrid data source: Automatic routing to API for recent dates (≤7 days) or archive for historical data
- Week-based navigation: Left/right arrows navigate weeks with auto-skip blank weeks and 52-week search limit
- Archive display wiring: All 9 stat panels show historical data via getActiveStatsData helper with [HISTORICAL] indicator
- Auto-refresh management: Pauses during historical view, resumes when returning to today
- Background prefetch: Instant backward navigation via silent prefetch cache for one week ahead
- End-of-history UX: Full-screen modal banner when reaching oldest data with navigation hints

**Archive:** `.planning/milestones/v2.2-ROADMAP.md`, `.planning/milestones/v2.2-REQUIREMENTS.md`, `.planning/milestones/v2.2-MILESTONE-AUDIT.md`

---

