# Feature Landscape

**Domain:** Terminal dashboard for WakaTime/Wakapi coding statistics
**Researched:** 2026-02-17
**Scope:** New capabilities only — wakafetch already provides static fetch, --full, --daily, --heatmap, --range/--days

---

## Context: What wakafetch Already Provides

These features exist and must NOT be rebuilt. The dashboard extends them.

- Static stat fetch (languages, projects, editors, OSs, categories)
- --full breakdown
- --daily table
- --heatmap (GitHub-style coding frequency)
- --range / --days time range selection
- WakaTime + Wakapi API support
- ~/.wakatime.cfg auto-config

---

## Table Stakes

Features users expect from a "live dashboard." Missing any of these and it feels like a slightly fancier wakafetch, not a real dashboard.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Auto-refresh loop | Defining characteristic of "live dashboard" (htop, btop, glances pattern) | Low | Ticker-based, configurable interval (default 60s). Users expect +/- keys to adjust like btop |
| Full-screen TUI layout | Dashboard implies a dedicated view occupying the terminal, not inline scrolling output | Low-Med | Bubbletea handles this; challenge is layout composition with lipgloss |
| Keyboard quit (q / Ctrl-C) | Universal TUI convention — absence is jarring and unprofessional | Low | Must restore terminal state cleanly; Bubbletea handles cleanup |
| Visible refresh indicator | Users need to know if data is live or stale | Low | Last-updated timestamp in header or footer, ideally showing seconds since last fetch |
| Graceful terminal resize | Dashboard must reflow on window resize without crashing | Med | Bubbletea's WindowSizeMsg handles events; layout math for panel sizing is nontrivial |
| Color-coded bar chart for top languages | Core "more visual than wakafetch" requirement; every WakaTime web dashboard shows this | Low-Med | Horizontal bar chart with percentage and time labels; ntcharts BarChart or lipgloss manual render |
| Summary header panel | Today total, week total, daily average — always visible at a glance | Low | Static panel at top; data from stats API endpoint already used in wakafetch |
| Top N projects panel | Second most-watched metric after languages in every developer stats tool | Low | Bar chart matching language panel style; same API data already fetched |
| Error state display | API failures must not crash or show Go stack traces to users | Low | Friendly error message in panel area; retry on next refresh tick |
| Help overlay (? key) | All serious TUI tools (lazygit, btop, k9s) provide in-app keybinding reference | Low | Static overlay showing all keybindings; dismiss with Esc or ? again |

---

## Differentiators

Features that make wakadash worth using over opening a browser tab or running wakafetch repeatedly.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Sparkline for today's hourly activity | Shows coding rhythm across the day — not exposed anywhere in wakafetch; answers "when am I most productive?" | Med | Requires durations API endpoint (separate from stats endpoint); ntcharts Streamline or Sparkline component |
| Multi-panel layout (languages + projects side-by-side) | Information density similar to btop; glanceable at a single terminal view | Med | Lipgloss grid/join layout; must degrade gracefully to stacked single-column on narrow terminals |
| Configurable refresh interval at runtime | btop's +/- pattern; user controls data staleness without restarting | Low | Keybinding adjusts ticker interval; display current interval in footer |
| Panel toggle (show/hide editors, OSs, machines panels) | btop region-toggle UX pattern; users focus on what they care about | Med | Track panel visibility in model state; dedicated keybinding per panel (e, o, m); persist in session only |
| Time range switcher (today / week / month) at runtime | Switch context without restarting the dashboard or re-running CLI flags | Med | r key (or tab) cycles ranges; triggers full API re-fetch; updates all panels simultaneously |
| Color themes (via --theme flag or config) | Differentiates from all existing WakaTime CLI tools; lipgloss makes this low-effort once layout works | Low-Med | 2-3 preset themes: dark (default), light, high-contrast; applied globally via lipgloss style config |
| Best day callout | "Your best day was X hours on Y" — motivating highlight; available directly from stats API best_day field | Low | Single-line display in header panel; no additional API call required |
| Editors panel | Surfaces editor usage visually — valuable for polyglot developers switching between vim/vscode/etc | Low | Same API data as languages; just another bar chart panel that can be toggled |

---

## Anti-Features

Features to explicitly NOT build in this milestone.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Mouse interaction / clickable panels | btop implements it at significant complexity cost; marginal gain in a keyboard-first tool; adds event handling complexity to every panel | Keyboard-only navigation; document all bindings in ? overlay |
| Per-file / per-branch granularity | WakaTime exposes this via heartbeats API but it is noisy, slow, and rarely the first thing users want on a dashboard view | Expose as a future --files or --branch flag in wakafetch if requested; keep dashboard at project level |
| Real-time heartbeat streaming | WakaTime API is summary/polling only, not an event stream; simulating real-time would be misleading about data freshness | Honest polling with a visible last-updated indicator showing exact timestamp |
| Editable config from within dashboard | Config editing in TUI requires form widget library, input validation, file write logic — significant scope expansion for a stats viewer | Direct users to edit ~/.wakatime.cfg and restart; fail fast on launch with a clear config error message |
| Interactive column sorting | Adds model state complexity for marginal value; sort order preference is rarely changed mid-session | Sort by time descending always; most-used first is the universal expectation and matches WakaTime web UI |
| Notifications or desktop alerts | Out of scope for a stats viewer; belongs in a separate monitoring/goal-tracking tool | Not planned |
| Authentication flow inside the dashboard | Secure input (masked API key entry) in a TUI requires careful handling; nontrivial edge cases with terminal echo | Require config to be set before launching; exit with a clear error message pointing to ~/.wakatime.cfg |
| Custom data export from within dashboard | CSV/JSON export is a different workflow (analytics, not dashboard); better handled as a separate CLI subcommand | Not planned for dashboard mode; could be a future wakafetch --export flag |

---

## WakaTime API Data Available for Visualization

Confirmed from official WakaTime API documentation. All of these are safe to build against.

**Segmented breakdowns from stats endpoint (each breakdown includes time + percentage):**
- Languages
- Projects
- Editors
- Operating systems
- Categories (Coding, Debugging, Building, Writing Tests, Writing Docs, etc.)
- Machines (useful for multi-device developers)
- Dependencies

**Summary fields from stats endpoint:**
- Total seconds for range
- Daily average seconds
- Best day (date + seconds)
- Human-readable time strings (pre-formatted)

**Requires separate durations API endpoint:**
- Hourly activity breakdown within a single day (needed for sparkline showing coding rhythm)

**Wakapi compatibility note:** Wakapi implements the WakaTime-compatible API. Not all endpoints exist in all Wakapi versions. The goals endpoint in particular may not be available. Code defensively for optional endpoints.

---

## Feature Dependencies

```
Auto-refresh loop
  --> Summary header panel (updates on each tick)
  --> Top N languages panel (updates on each tick)
  --> Top N projects panel (updates on each tick)
  --> Refresh indicator (shows time since last tick completed)

Full-screen TUI layout
  --> All panels (panels need a layout container to anchor to)
  --> Multi-panel side-by-side layout (requires layout foundation first)
  --> Panel toggle (requires panel identity tracked in model)

Sparkline for today's hourly activity
  --> Durations API endpoint integration (separate from stats; new API call)
  --> Full-screen TUI layout (needs dedicated panel area)
  --> ntcharts Sparkline/Streamline component

Time range switcher (runtime r key)
  --> All data panels (all re-fetch when range changes)
  --> Auto-refresh loop (range selection persists across ticks)

Panel toggle
  --> Full-screen TUI layout (panels must have identifiable regions)
  --> Multi-panel layout (toggle only meaningful when multiple panels exist)

Color themes
  --> Full-screen TUI layout (themes must apply globally via lipgloss styles)
  --> All panels (every rendered component references theme palette)

Editors panel
  --> Same API call as languages/projects (no additional fetch needed)
  --> Panel toggle (editors panel is a toggleable panel, not always visible)
```

---

## MVP Recommendation

### Phase 1: Core Dashboard (ship first)

1. Full-screen Bubbletea TUI layout with lipgloss
2. Auto-refresh loop (60s default; +/- to adjust at runtime)
3. Summary header panel (today total, week total, daily average, best day callout)
4. Top N languages bar chart panel
5. Top N projects bar chart panel
6. Keyboard quit (q), help overlay (?)
7. Visible last-updated timestamp in footer
8. Graceful resize handling
9. Error state display (friendly, non-crashing)

### Phase 2: Enhanced Visualization (second milestone)

10. Sparkline for today's hourly activity (requires durations endpoint)
11. Panel toggle keybindings (e for editors, o for OS, m for machines)
12. Time range switcher (r key cycles today/week/month)
13. Editors panel (toggleable)
14. Multi-column layout on wide terminals (degrade to stacked on narrow)
15. Color theme support (--theme flag, 2-3 presets)

### Defer to Phase 3 or Later

- Goals and streak display (Wakapi may not support; goals API varies)
- Operating systems panel (low user interest relative to complexity of adding another panel)
- Machines panel (only useful for multi-device developers; niche)
- Dependencies breakdown (rarely the first thing developers check)

---

## Comparison to Existing Tools

| Feature | wakafetch (existing) | wakatime-cli (jaebradley) | wakadash (target) |
|---------|---------------------|--------------------------|-------------------|
| Output mode | Inline stdout | Inline stdout | Full-screen TUI |
| Auto-refresh | No | No | Yes (configurable) |
| Bar charts | No (text only) | No | Yes (ntcharts) |
| Sparklines | No | No | Yes (phase 2) |
| Multi-panel | No | No | Yes |
| Time range switch | CLI flag, restart required | CLI flag, restart required | Runtime r key |
| Color themes | No | No | Yes (2-3 presets) |
| Heatmap | Yes (--heatmap flag) | No | Inherit via wakafetch; not duplicated |
| Help overlay | No | No | Yes (? key) |
| Editors visible | --full flag only | Limited | Dedicated panel (toggleable) |

---

## Sources

- [WakaTime API Documentation](https://wakatime.com/developers) — confirmed all data fields, endpoint scopes, and data shapes (HIGH confidence)
- [NimbleMarkets/ntcharts](https://github.com/NimbleMarkets/ntcharts) — confirmed chart types: Sparkline, BarChart, HeatMap, StreamlineChart, TimeSeriesChart (HIGH confidence)
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) — confirmed as Go TUI framework standard; v2 released October 2025 (HIGH confidence)
- [btop++ feature analysis](https://linuxblog.io/btop-the-htop-alternative/) — +/- refresh adjustment, region toggle, keyboard-first UX patterns (MEDIUM confidence, community sources)
- [Wakapi feature set](https://wakapi.dev/) — confirmed WakaTime API compatibility and scope of supported endpoints (MEDIUM confidence)
- [sahaj-b/wakafetch](https://github.com/sahaj-b/wakafetch) — baseline existing feature set to avoid rebuilding (HIGH confidence, this is the existing project)
- [4 Ways to Visualize WakaTime Programming Data](https://wakatime.com/blog/15-4-ways-to-visualize-your-programming-data) — dashboard patterns and common visualizations (MEDIUM confidence)
- [Grafana WakaTime Dashboard](https://grafana.com/grafana/dashboards/12790-wakatime-coding-stats/) — reference for which metrics are most valued by developers building dashboards (MEDIUM confidence)
