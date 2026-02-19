# wakadash

## What This Is

A terminal dashboard for WakaTime/Wakapi coding stats. Displays live-updating visualizations of coding activity with color-coded charts, sparklines, and heatmaps — bringing the WakaTime web dashboard experience to the terminal.

## Core Value

A beautiful, live-updating terminal dashboard that shows your coding stats at a glance — like htop for your coding activity.

## Current Milestone: v2.1 Visual Overhaul + Themes

**Goal:** Enhance dashboard with comprehensive stats panels matching wakafetch visual style, plus configurable color themes.

**Target features:**
- Full stats panels: Languages, Projects, Categories, Editors, Operating Systems, Machines (top 10 each, horizontal bars with times)
- Stats summary panel: Last 30 days, Total Time, Daily Avg, Top Project/Editor/Category/OS, counts
- Built-in theme presets: Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night
- Theme selection via flag or config

## Requirements

### Validated

- ✓ Fork b00y0h/wakafetch from upstream — v1.0
- ✓ GoReleaser multi-platform builds (darwin/linux × amd64/arm64) — v1.0
- ✓ GitHub Actions workflow on version tags — v1.0
- ✓ Homebrew tap (b00y0h/homebrew-wakafetch) — v1.0
- ✓ Cross-repo PAT for formula publishing — v1.0
- ✓ End-to-end release with v0.1.0 — v1.0
- ✓ macOS quarantine removal (no Gatekeeper warnings) — v1.0

### Active

(To be defined — running v2.1 requirements phase)

### Out of Scope

- Windows builds — Homebrew is macOS/Linux focused
- Real-time WebSocket streaming from WakaTime — API polling sufficient
- Mobile/web interface — terminal only
- Multi-user/server mode — single-user CLI tool

## Context

**Origin:** This project started as release automation for a wakafetch fork. v2.0 pivots to creating a standalone, enhanced tool called wakadash.

**Technical foundation:**
- Port wakafetch Go code to new repo (WakaTime/Wakapi API integration)
- Add TUI/terminal graphics capabilities for live dashboard
- Fresh repo enables homebrew-core submission (no fork restrictions)

**Key insight from v1.0:** homebrew-core won't accept forks. Creating original work with unique value proposition solves this.

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
| Create standalone wakadash | Enables homebrew-core, adds unique value | — Pending |
| Port and enhance wakafetch | Preserve working API code, build on top | — Pending |
| Dashboard mode with live updates | Differentiates from simple fetch tools | — Pending |

---
*Last updated: 2026-02-19 after v2.1 milestone start*
