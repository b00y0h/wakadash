# wakadash

## What This Is

A terminal dashboard for WakaTime/Wakapi coding stats. Displays live-updating visualizations of coding activity with color-coded charts, sparklines, heatmaps, and comprehensive stats panels — bringing the WakaTime web dashboard experience to the terminal with 6 beautiful theme presets.

## Core Value

A beautiful, live-updating terminal dashboard that shows your coding stats at a glance — like htop for your coding activity.

## Current State

**Shipped:** v2.1 Visual Overhaul + Themes (2026-02-23)

**Codebase:** 2,638 LOC Go, BubbleTea TUI framework

**Features shipped:**
- Full-screen dashboard with auto-refresh (configurable interval)
- 9 visualization panels: Languages, Projects, Categories, Editors, OS, Machines, Sparkline, Heatmap, Summary
- 6 theme presets: Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night
- Theme picker with live preview (first-run and runtime 't' key)
- Responsive layout: 2-column grid >= 80 cols, vertical stack at 40-79 cols
- Keyboard controls: 1-9 panel toggles, a/h show/hide all, ? help, q quit
- Rate limiting: Exponential backoff with visual indicator
- Edge case handling: Terminal size validation, case-insensitive config, zero-division protection

**Distribution:**
- Personal tap: `brew tap b00y0h/wakadash && brew install wakadash`
- homebrew-core: Pending (requires >= 30 forks, >= 30 watchers, or >= 75 stars)

## Requirements

### Validated

**v1.0:**
- ✓ Fork b00y0h/wakafetch from upstream — v1.0
- ✓ GoReleaser multi-platform builds (darwin/linux × amd64/arm64) — v1.0
- ✓ GitHub Actions workflow on version tags — v1.0
- ✓ Homebrew tap (b00y0h/homebrew-wakafetch) — v1.0
- ✓ Cross-repo PAT for formula publishing — v1.0
- ✓ End-to-end release with v0.1.0 — v1.0
- ✓ macOS quarantine removal (no Gatekeeper warnings) — v1.0

**v2.0:**
- ✓ Fresh wakadash repository created (b00y0h/wakadash) — v2.0
- ✓ Full-screen TUI dashboard with async API calls — v2.0
- ✓ Languages and projects bar charts with colors — v2.0
- ✓ Hourly activity sparkline and weekly heatmap — v2.0
- ✓ Keyboard navigation and panel toggles — v2.0
- ✓ Personal Homebrew tap (b00y0h/homebrew-wakadash) — v2.0

**v2.1:**
- ✓ 6 theme presets (Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night) — v2.1
- ✓ Visual theme preview on first run — v2.1
- ✓ Theme persistence in ~/.wakatime.cfg — v2.1
- ✓ All panels use selected theme colors — v2.1
- ✓ Categories, Editors, OS, Machines panels with top-10 bars — v2.1
- ✓ Summary panel with 30-day stats and streak calculation — v2.1
- ✓ Responsive 2-column layout at >= 80 cols — v2.1
- ✓ Extended keyboard toggles (5-9, a, h) — v2.1

### Active

(Define in next milestone)

### Out of Scope

- Windows builds — Homebrew is macOS/Linux focused
- Real-time WebSocket streaming — API polling sufficient
- Mobile/web interface — terminal only
- Multi-user/server mode — single-user CLI tool
- Custom theme editor in TUI — config file editing simpler
- Light theme variants — dark themes match terminal aesthetic
- 100+ theme pack — 6 curated themes optimal for decision simplicity

## Context

**Origin:** This project started as release automation for a wakafetch fork. v2.0 created a standalone, enhanced tool called wakadash.

**Technical stack:**
- Go 1.21+ with BubbleTea TUI framework
- ntcharts for bar chart visualization
- lipgloss for terminal styling
- WakaTime/Wakapi REST API integration

**Key insight from v1.0:** homebrew-core won't accept forks. Creating original work with unique value proposition solved this.

**Key insight from v2.1:** Theme system with hex colors works well; lipgloss handles terminal color downsampling automatically.

## Constraints

- **Go language**: Continuing with Go for consistency and performance
- **WakaTime API**: Must work with existing WakaTime/Wakapi API (no custom backend)
- **Terminal-native**: Pure terminal output, no external dependencies
- **homebrew-core eligible**: Must meet Homebrew's criteria for core inclusion

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Fork instead of tap-only | Full control over releases; upstream has no releases | ✓ Good |
| GoReleaser over manual | Industry standard for Go release automation | ✓ Good |
| Fine-grained PAT | More secure than classic PAT with broad repo scope | ✓ Good |
| homebrew_casks over brews | Modern GoReleaser v2.10+ syntax, v3-ready | ✓ Good |
| xattr quarantine removal | Prevents macOS Gatekeeper warnings | ✓ Good |
| Keep personal tap only | brew.sh requires homebrew-core + upstream cooperation | — Accepted |
| Create standalone wakadash | Enables homebrew-core, adds unique value | ✓ Good |
| Port and enhance wakafetch | Preserve working API code, build on top | ✓ Good |
| Dashboard mode with live updates | Differentiates from simple fetch tools | ✓ Good |
| BubbleTea with tea.WithAltScreen() | Avoids race conditions vs EnterAltScreen command | ✓ Good |
| Self-loop ticker pattern | Avoids ticker drift over long dashboard sessions | ✓ Good |
| GitHub Linguist colors for languages | Familiar colors for developers | ✓ Good |
| cenkalti/backoff/v5 for rate limiting | Exponential backoff with jitter (1s-30s, 2min max) | ✓ Good |
| Hex colors for themes | Lipgloss auto-handles terminal downsampling | ✓ Good |
| Persist theme to ~/.wakatime.cfg | Reuses existing config file; single source of truth | ✓ Good |
| Top-10 limit with "Other" aggregation | Prevents UI overflow; maintains readability | ✓ Good |
| Summary panel with accent border | Visual distinction from stat panels | ✓ Good |
| Case-insensitive theme lookup | Forgiving config parsing for user-friendliness | ✓ Good |

---
*Last updated: 2026-02-23 after v2.1 milestone*
