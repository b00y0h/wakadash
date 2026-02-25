# wakadash

## What This Is

A live terminal dashboard for WakaTime coding stats. Like htop for your coding activity — shows projects, languages, editors, and time breakdown live in your terminal.

## Core Value

Real-time visibility into coding activity without leaving the terminal.

## Current Milestone: v2.2 Version Update Check

**Goal:** Add gh-style version update checking on startup

**Target features:**
- Background check for new versions via GitHub releases
- Multi-line notice in status bar when update available
- Display version diff and `brew upgrade wakadash` command

## Requirements

### Validated

- Dashboard displays live WakaTime stats
- Supports WakaTime and Wakapi backends
- Reads `~/.wakatime.cfg` automatically
- GitHub archive integration for historical data
- Themed UI with bordered panels
- Horizontal bar charts for stats visualization

### Active

- [ ] Version update check on startup (background/async)
- [ ] Display update notice in status bar
- [ ] Show upgrade command for Homebrew users

### Out of Scope

- Auto-update functionality — keep it manual like gh CLI
- Update checks for non-Homebrew installs — Homebrew only for v1

## Context

- Built with Go + Bubble Tea TUI framework
- Distributed via Homebrew tap (b00y0h/wakadash)
- Inspired by wakafetch
- GitHub releases used for version distribution

## Constraints

- **Performance**: Update check must not block dashboard startup
- **Network**: Graceful handling when GitHub unreachable
- **UX**: Non-intrusive notice, doesn't interrupt workflow

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Async check | Don't delay dashboard load | — Pending |
| Status bar placement | Non-intrusive, visible | — Pending |
| Homebrew-only | Simplifies command suggestion | — Pending |

---
*Last updated: 2026-02-25 after milestone v2.2 start*
