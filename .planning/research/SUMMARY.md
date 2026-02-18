# Project Research Summary

**Project:** wakadash — Live Dashboard Milestone (added to wakafetch)
**Domain:** Live-updating terminal dashboard (TUI) for WakaTime/Wakapi coding statistics
**Researched:** 2026-02-17
**Confidence:** HIGH (stack + architecture + core pitfalls) / MEDIUM (ntcharts maturity, terminal compatibility)

## Executive Summary

wakadash is a live-refresh terminal dashboard mode added to the existing wakafetch Go CLI. The pattern is well-established — htop, btop, and k9s define user expectations. The charmbracelet ecosystem (bubbletea + lipgloss + bubbles) is the clear standard for Go TUI development in 2026, with bubbletea v1.3.10 being production-stable (v2 is alpha and must be avoided). The existing wakafetch codebase is well-structured for this extension: api.go is pure and reusable, ui/ functions already return composable []string slices, and config loading is already decoupled. The recommended approach is to add a new `dashboard/` package that wraps existing code rather than rewriting it, activated by a `--dashboard` / `-w` flag.

The primary architectural decision is to adopt the Elm Architecture (Model-Update-View) from bubbletea, which solves the live-refresh problem cleanly via tea.Tick and async tea.Cmd for HTTP calls. The most dangerous mistakes are concurrency-related: mutating model state outside Update(), blocking the event loop with synchronous API calls, and goroutine leaks from mismanaged tickers. All three are prevented by consistently using tea.Cmd for async work — this is the single most important pattern to enforce from day one. ntcharts is the only bubbletea-native charting library but has no tagged releases; the existing handwritten graph/heatmap code in ui/ is a viable fallback if ntcharts proves problematic.

For features, the research identifies a clear MVP boundary: a full-screen TUI with auto-refresh, summary header, languages and projects bar charts, visible refresh status, keyboard quit, help overlay, graceful resize, and non-crashing error display. This set differentiates wakadash from a fancier wakafetch without overreaching. The sparkline for hourly activity, panel toggles, runtime range switching, and color themes are confirmed differentiators that belong in a second phase after the dashboard foundation is solid. Mouse interaction, real-time streaming, in-app config editing, and interactive sorting are explicitly out of scope and should not be built.

## Key Findings

### Recommended Stack

The existing wakafetch codebase has zero external dependencies — a deliberate design choice. This milestone is the first point where external libraries are introduced, making the choice consequential. The charmbracelet ecosystem is the right call: bubbletea v1.3.10 for the event loop, lipgloss v1.1.0 for layout composition, and bubbles v1.0.0 for loading spinners and help components. These three form a coherent, actively maintained suite (last releases: Sep 2025, Mar 2025, Feb 2025 respectively) with 18,200+ dependent projects on bubbletea alone. They are the first external dependencies added via go.mod.

ntcharts (NimbleMarkets/ntcharts) provides the only purpose-built charting for bubbletea: sparklines, bar charts, and heatmaps. The risk is its lack of tagged releases — pin by commit hash. If ntcharts becomes unmaintained, the existing graphStr() and heatmap() functions in ui/ are a direct fallback since they already produce []string slices that compose into bubbletea View() methods. Do not use bubbletea v2 (alpha, breaking API, import path changed); do not introduce Cobra/pflag (conflicts with existing stdlib flag setup).

**Core technologies:**
- charmbracelet/bubbletea v1.3.10: TUI event loop, Elm Architecture, auto-refresh via tea.Tick — the only Go TUI framework with production stability and ecosystem depth at this scale
- charmbracelet/lipgloss v1.1.0: Layout composition, borders, responsive column sizing — replaces handwritten ANSI layout code; handles color degradation automatically
- charmbracelet/bubbles v1.0.0: Spinner, help component — pre-built components that match bubbletea's model/update/view cycle
- NimbleMarkets/ntcharts (pin by commit): Sparklines, bar charts, heatmaps native to bubbletea — the only library supporting all three required chart types in a bubbletea context

### Expected Features

The research identifies what users expect from a "live dashboard" versus what differentiates wakadash from both wakafetch and browser-based WakaTime dashboards. See FEATURES.md for the full comparison table against existing tools.

**Must have (table stakes):**
- Auto-refresh loop (configurable interval, default 60s; +/- keys to adjust) — the defining characteristic of a live dashboard
- Full-screen TUI layout (AltScreen mode, not inline scrolling) — users expect dedicated terminal occupation
- Keyboard quit (q / Ctrl-C) with clean terminal restore — absence is jarring; bubbletea handles this
- Visible refresh indicator (last-updated timestamp in footer) — users must know if data is live or stale
- Graceful terminal resize without crashing — mandatory for any tool used in resizable terminals
- Color-coded bar chart for top languages — core visual requirement; every WakaTime dashboard shows this
- Summary header panel (today total, week total, daily average, best day) — always-visible glanceable stats
- Top N projects bar chart panel — second most-watched metric after languages
- Error state display (friendly, non-crashing) — API failures must not surface Go stack traces
- Help overlay (? key) — all serious TUI tools provide in-app keybinding reference

**Should have (competitive):**
- Sparkline for today's hourly activity — answers "when am I most productive?" — requires separate durations API endpoint
- Multi-panel layout (side-by-side on wide terminals, stacked on narrow) — information density matching btop
- Configurable refresh interval at runtime (+/- keys) — btop pattern; users control data staleness
- Panel toggle (show/hide editors, OSs, machines panels) — btop region-toggle UX
- Time range switcher at runtime (r key cycles today/week/month) — eliminates restart-to-change-range workflow
- Color themes (--theme flag, 2-3 presets: dark/light/high-contrast) — differentiates from all existing WakaTime CLI tools
- Editors panel (toggleable bar chart) — valuable for polyglot developers

**Defer (v2+):**
- Goals and streak display (Wakapi may not support goals API; endpoint availability is uncertain)
- Operating systems panel (low user interest relative to adding another panel)
- Machines panel (niche — only useful for multi-device developers)
- Dependencies breakdown (rarely the first developer dashboard check)
- Mouse interaction, real-time heartbeat streaming, in-app config editing, interactive sorting (all explicitly anti-features for this milestone)

### Architecture Approach

The recommended architecture introduces a new `dashboard/` package (model.go, messages.go, fetch.go, layout.go) that wraps existing code. The existing api.go, config.go, types/, and ui/ packages require zero logic changes — only ui/ function export visibility needs updating (lowercase graphStr → exported GraphStr, etc.). The entry point (main.go, flags.go) adds the --dashboard/-w flag branch and launches tea.NewProgram with AltScreen. This is the minimum-change strategy: the static CLI mode remains entirely unchanged.

The data flow is: Init() returns tea.Batch(fetchStatsCmd(), tickCmd()) to kick off initial fetch and schedule the first tick concurrently. The tick fires after the configured interval; Update() checks elapsed time and issues a new fetch command. All HTTP calls run in goroutines managed by bubbletea via tea.Cmd — never called synchronously in Update(). View() calls layout.RenderDashboard(m), which uses lipgloss.JoinVertical/Horizontal to compose ui/ function outputs into panels.

**Major components:**
1. `dashboard/model.go` — Root bubbletea model: holds all dashboard state, routes all messages to panels
2. `dashboard/messages.go` — Typed Msg definitions: TickMsg, StatsDataMsg, SummaryDataMsg, ErrMsg
3. `dashboard/fetch.go` — tea.Cmd wrappers around existing api.go calls; no HTTP logic duplicated
4. `dashboard/layout.go` — Lipgloss composition of ui/ []string returns into joined panel view string + status bar
5. `ui/*.go` (existing, minimally modified) — Export visibility changes only; logic unchanged
6. `api.go` / `config.go` / `types/` (existing, unchanged) — Reused directly as-is

### Critical Pitfalls

The pitfall research is unusually strong because the codebase was analyzed directly. All pitfalls are traced to specific existing code patterns. Six critical pitfalls were identified; five must be addressed in Phase 1 (TUI Foundation) before any feature work begins.

1. **Mutating model state outside Update()** — Never write model.field = value from a goroutine. All API responses must arrive as tea.Msg through the event loop. Enforce with `go run -race .` throughout development. Getting this wrong causes non-deterministic crashes that masquerade as rendering bugs.

2. **Blocking the event loop with synchronous API calls** — The existing fetchStats() and fetchSummary() use synchronous HTTP with a 10-second timeout. Calling these directly in Init() or Update() freezes the entire UI. Wrap every API call in tea.Cmd from the start — this cannot be retrofitted cleanly.

3. **Panic leaves terminal in raw mode** — Any unrecovered panic in a bubbletea Cmd goroutine can leave the terminal unusable. Do not use tea.WithoutCatchPanics(). Never call os.Exit() from within a running bubbletea program (the existing showCustomHelp() does this — it must not be reused inside TUI paths).

4. **Terminal width via stty subprocess** — The existing getTerminalCols() spawns a stty subprocess. In AltScreen TUI mode this is slow, breaks on resize, and returns 9999 on Windows. Replace with bubbletea's WindowSizeMsg handler from the start. This cannot be retrofitted without rewriting the layout layer.

5. **Flag system conflict** — The existing codebase uses stdlib flag package. Introducing Cobra/pflag creates silent conflicts or panics on duplicate registration. Add --dashboard and --interval to the existing stdlib flag setup. Do not introduce Cobra in this milestone.

6. **Goroutine leaks from ticker** — Using time.NewTicker directly creates goroutines that never clean up when the ticker is recreated (e.g., user changes refresh interval). Use bubbletea's tea.Tick pattern exclusively — one-shot per tick, requeued after each data arrival, cancelled on quit.

## Implications for Roadmap

All four research sources converge on a 3-phase structure with a clear dependency ordering. The architecture research provides the build sequence directly, validated by the pitfall research showing which decisions must come first.

### Phase 1: TUI Foundation

**Rationale:** Five of six critical pitfalls must be addressed before any feature work begins. Architectural decisions made here (state model, message types, async patterns, flag system, terminal width handling) cannot be changed later without rewriting the dashboard layer. The dependency graph in ARCHITECTURE.md confirms that messages.go and fetch.go must exist before model.go, which must exist before layout.go, which must exist before any feature panels.

**Delivers:** A running but minimal bubbletea dashboard — full-screen AltScreen mode, initial data fetch (async), loading state, basic stats display, keyboard quit, terminal resize handling. No auto-refresh yet; no fancy layout. Proves the integration works end-to-end before adding complexity.

**Addresses (from FEATURES.md):** Full-screen TUI layout, keyboard quit (q/Ctrl-C), graceful terminal resize, error state display, summary header panel, visible refresh indicator

**Avoids (from PITFALLS.md):** Model state mutation outside Update() (#1), blocking event loop (#2), panic leaving raw mode (#3), stty subprocess (#6), flag system conflict (#10), state complexity explosion (#11)

**Stack elements:** bubbletea v1.3.10, lipgloss v1.1.0; bubbles optional at this phase

**Architecture components:** dashboard/messages.go, dashboard/fetch.go, dashboard/model.go, dashboard/layout.go (minimal), ui/ export visibility changes, main.go + flags.go entry point

**Research flag:** Standard patterns — bubbletea Elm Architecture is thoroughly documented in official docs and tutorials. ARCHITECTURE.md provides the complete build sequence. No additional research-phase needed; follow ARCHITECTURE.md build sequence directly.

### Phase 2: Live Refresh and Data Panels

**Rationale:** With the TUI foundation proven, all remaining table-stakes features can be implemented. The auto-refresh loop is the defining characteristic of the dashboard; the bar chart panels are the primary visual differentiator from wakafetch. Three API-level pitfalls (goroutine leaks, 429 rate limiting, 202 computing state) are specific to the refresh loop and must be addressed here, not deferred.

**Delivers:** The complete MVP dashboard — auto-refresh ticker, languages bar chart panel, projects bar chart panel, configurable refresh interval (--interval flag and +/- keys), help overlay (? key), last-updated footer with countdown, rate-limit handling with exponential backoff, 202 "calculating" state display. This is the shippable milestone.

**Addresses (from FEATURES.md):** Auto-refresh loop, color-coded bar chart for top languages, top N projects panel, help overlay, configurable refresh interval at runtime, best day callout in header

**Avoids (from PITFALLS.md):** Goroutine leaks from ticker (#4), WakaTime 202 response treated as error (#5), 429 rate limit without backoff (#7), ANSI codes in non-TTY output (#8), hardcoded intervals hitting API (#12), 302 redirect treated as rate limit (#13), cached_at staleness (#14)

**Stack elements:** ntcharts (bar chart component) or existing graphStr() as fallback; bubbles spinner for loading state

**Architecture components:** Auto-refresh tick loop in model.go, ntcharts BarChart in layout.go, status bar with countdown, enhanced fetch.go with backoff and 202/302 handling

**Research flag:** ntcharts integration needs validation before this phase begins — no tagged releases, heatmap quality vs existing custom renderer needs comparison. Validate ntcharts bar chart output quality before committing to it; the existing graphStr() is a full fallback.

### Phase 3: Enhanced Visualization and Polish

**Rationale:** With a solid, correct dashboard in place, the differentiating features (sparkline, panel toggles, runtime range switching, color themes) can be added incrementally. These are all model-state additions — new keys update model fields, new renders respond to them. The architecture is already set up for this; they are additive. Terminal compatibility polish (true color degradation, Unicode fallbacks) belongs here as well.

**Delivers:** The full differentiated dashboard — sparkline for hourly activity (hourly rhythm view), panel toggle keybindings (e/o/m), runtime time range switcher (r key), multi-column layout on wide terminals, editors panel, color theme support (--theme flag), terminal compatibility hardening.

**Addresses (from FEATURES.md):** Sparkline for today's hourly activity, panel toggle, time range switcher, color themes, multi-panel side-by-side layout, editors panel

**Avoids (from PITFALLS.md):** True color in degraded terminals (#9), Unicode bar chars in limited fonts (#15)

**Stack elements:** ntcharts Sparkline/StreamlineChart for hourly activity; lipgloss adaptive colors for theme system

**Architecture components:** Durations API endpoint integration in fetch.go, sparkline panel in layout.go, panel visibility state in model, theme configuration via lipgloss style config

**Research flag:** Wakapi durations endpoint availability needs validation — not all Wakapi versions are confirmed to support the hourly activity endpoint needed for the sparkline. Design Phase 3 sparkline with a capability check — if endpoint returns 404, hide the sparkline panel silently rather than showing an error.

### Phase Ordering Rationale

- Pitfall research is unambiguous that concurrency/event-loop correctness must come first — five critical pitfalls are Phase 1 concerns that cannot be retrofitted after feature work begins
- Feature dependency graph from FEATURES.md confirms ordering: full-screen TUI layout must exist before panels, panels before toggles, all before themes
- Architecture build sequence from ARCHITECTURE.md maps directly to Phase 1: messages.go → fetch.go → layout.go → model.go → entry point
- Sparkline (Phase 3) requires the durations API endpoint, a separate HTTP call with uncertain Wakapi support — correctness gate before adding that complexity
- Color themes (Phase 3) require lipgloss to be established across all panels — must come after all panels exist in Phase 2

### Research Flags

Phases needing deeper research during planning:
- **Phase 2:** Validate ntcharts bar chart output quality vs existing graphStr() before choosing ntcharts vs the existing code. The existing heatmap renderer uses a custom 5-level green RGB gradient that ntcharts may not replicate exactly — determine early whether to use ntcharts or keep the existing chart renderers.
- **Phase 3:** Confirm Wakapi durations endpoint availability across common Wakapi versions before designing the sparkline panel. Design for graceful absence (hide panel, not error).

Phases with standard patterns (skip research-phase):
- **Phase 1:** bubbletea Elm Architecture is thoroughly documented. ARCHITECTURE.md provides the exact build sequence and all code patterns needed.
- **Phase 2 (refresh loop):** tea.Tick pattern is official API with clear documentation. PITFALLS.md provides the exact code patterns for backoff and 202 handling.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | bubbletea/lipgloss/bubbles all have official docs and stable releases. ntcharts is MEDIUM — no tagged releases, small team (4 contributors); fallback to existing ui/ code is viable and confirmed composable. |
| Features | HIGH | WakaTime API confirmed for all data fields. Existing codebase confirmed as baseline. Feature set validated against btop/k9s/lazygit UX patterns. Wakapi compatibility is MEDIUM for optional endpoints. |
| Architecture | HIGH | Official bubbletea pkg.go.dev docs verified all patterns (WindowSizeMsg, tea.Cmd, tea.Tick, AltScreen). Existing codebase read directly to confirm reuse path and identify specific files/functions. |
| Pitfalls | HIGH (concurrency), MEDIUM (terminal compat) | Concurrency pitfalls verified against official Go race detector docs and bubbletea API docs. Terminal compatibility from community sources only — no single authoritative spec for degradation behavior. |

**Overall confidence:** HIGH

### Gaps to Address

- **ntcharts heatmap fidelity:** The existing heatmap uses a custom 5-level green RGB gradient with specific color values. ntcharts heatmap may not match visually. Validate before Phase 2 — if the custom gradient is important to the project, keep the existing heatmap renderer and only use ntcharts for bar charts and sparklines.
- **Wakapi durations endpoint:** MEDIUM confidence that all Wakapi versions support the hourly activity durations endpoint needed for the sparkline. Design Phase 3 sparkline with a capability check — if endpoint returns 404, hide the panel silently rather than showing an error.
- **Refresh interval default:** Research suggests 60s minimum to stay well under WakaTime's rate limit (10 req/s averaged over 5 minutes) when making 2 API calls per refresh. Validate the actual call count per refresh cycle and adjust the default accordingly during Phase 2 implementation.
- **WakaTime free plan stats lag:** Free plan users may regularly see 202 responses on first load. The "Calculating..." state needs clear UX — not an error, but a distinguishable waiting state. Confirm this is common enough to warrant a polished UI treatment before Phase 2 ships.

## Sources

### Primary (HIGH confidence)
- [charmbracelet/bubbletea v1.3.10](https://github.com/charmbracelet/bubbletea) — Elm Architecture, tea.Tick, tea.Every, tea.Batch, tea.Cmd, WithAltScreen, WindowSizeMsg
- [pkg.go.dev bubbletea](https://pkg.go.dev/github.com/charmbracelet/bubbletea) — CatchPanics, ErrInterrupted, WithoutCatchPanics behavior
- [charmbracelet/lipgloss v1.1.0](https://github.com/charmbracelet/lipgloss) — JoinHorizontal, JoinVertical, Border styles, adaptive colors
- [charmbracelet/bubbles v1.0.0](https://github.com/charmbracelet/bubbles/releases) — Spinner, help components
- [WakaTime API Documentation](https://wakatime.com/developers) — Rate limits, 202 handling, cached_at, all data fields confirmed
- [Existing codebase: /workspace/wakafetch](https://github.com/sahaj-b/wakafetch) — api.go, ui/*.go, types/types.go, main.go — read directly for integration analysis
- [Go Race Detector](https://go.dev/doc/articles/race_detector) — Race condition detection methodology
- [Bubbletea Commands Tutorial](https://github.com/charmbracelet/bubbletea/blob/main/tutorials/commands/README.md) — Official async command patterns

### Secondary (MEDIUM confidence)
- [NimbleMarkets/ntcharts](https://github.com/NimbleMarkets/ntcharts) — Chart types confirmed: Sparkline, BarChart, HeatMap, StreamlineChart. No tagged releases — pin by commit.
- [Tips for Building Bubble Tea Programs](https://leg100.github.io/en/posts/building-bubbletea-programs/) — Event loop pitfalls, state management patterns, layout arithmetic
- [Build a System Monitor TUI in Go](https://penchev.com/posts/create-tui-with-go/) — htop-style dashboard pattern with bubbletea, tea.Every refresh
- [Handling Polling in BubbleTea for Go](https://m3talsmith.medium.com/handling-polling-in-bubbletea-for-go-b17185835549) — tea.Batch + tick pattern validation
- [Wakapi](https://wakapi.dev/) — WakaTime API compatibility scope, supported endpoints
- [btop++ feature analysis](https://linuxblog.io/btop-the-htop-alternative/) — +/- refresh adjustment, region toggle, keyboard-first UX reference

### Tertiary (LOW confidence, needs validation)
- ntcharts maintenance trajectory — 4 contributors, no tagged releases. Validate heatmap output quality and long-term maintenance before committing.
- Wakapi durations endpoint availability across versions — not all Wakapi deployments confirmed. Design for graceful absence.
- ANSI escape code terminal compatibility specifics — community sources only; no single authoritative spec for degradation behavior.

---
*Research completed: 2026-02-17*
*Ready for roadmap: yes*
