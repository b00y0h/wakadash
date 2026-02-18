# Technology Stack

**Project:** wakadash — Live Dashboard Milestone
**Researched:** 2026-02-17
**Confidence:** HIGH (bubbletea ecosystem) / MEDIUM (ntcharts maturity)

---

## Context: What Already Exists

The existing `wakafetch` codebase is zero-dependency pure Go. It already implements:
- Raw ANSI escape codes for color (inline `\x1b[...]` sequences in `ui/colors.go`)
- Custom heatmap renderer in `ui/heatmap.go` (custom green RGB gradient)
- Horizontal bar chart in `ui/graph.go`
- Card/border layout system in `ui/card.go` and `ui/render.go`
- Terminal width detection via `stty` syscall (`ui/render.go:getTerminalCols`)

This is important: the new dashboard milestone adds a **live-refresh mode** (like htop) on top of existing static rendering. The question is which library to adopt, not whether to rewrite from scratch.

---

## Recommended Stack Additions

### Primary TUI Framework

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| charmbracelet/bubbletea | v1.3.10 (stable) | Live-refresh dashboard loop, event handling, auto-resize | Elm Architecture maps cleanly to a polling dashboard. `tea.Tick` provides the htop-style refresh loop. v1.x is production-stable with no breaking changes. 18,200+ dependent projects. Active: last release Sep 2025. |
| charmbracelet/lipgloss | v1.1.0 | Layout composition, borders, responsive column sizing | Replaces the handwritten `cardify()` + `printLeftRight()` layout system. CSS-like API. Handles ANSI color degradation automatically. Last release Mar 2025. |
| charmbracelet/bubbles | v1.0.0 | Spinner for loading state, help text component | Pre-built components that match bubbletea's model/update/view cycle. Avoids reinventing loading indicators and keyboard help views. Released Feb 2025. |

### Chart Rendering

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| NimbleMarkets/ntcharts | latest commit (no formal releases) | Sparklines, bar charts, heatmaps inside bubbletea views | Only terminal charting library purpose-built for bubbletea. Supports all three required chart types (sparkline, bar chart, heatmap). 632 stars, 4 contributors, 121 commits, actively maintained. Uses lipgloss + BubbleZone. |

**ntcharts caveat:** No tagged releases. Pin by commit hash in go.mod for reproducibility. This is the primary risk in this stack — if ntcharts becomes unmaintained, the existing handwritten chart code is a viable fallback since it already works.

**Alternative if ntcharts is abandoned:** The existing `graphStr()` (bar chart) and `heatmap()` functions in `ui/` can be ported directly into bubbletea `View()` methods. No external charting library is strictly required.

---

## Full Dependency List to Add

```bash
go get github.com/charmbracelet/bubbletea@v1.3.10
go get github.com/charmbracelet/lipgloss@v1.1.0
go get github.com/charmbracelet/bubbles@v1.0.0
go get github.com/NimbleMarkets/ntcharts@latest
```

Current `go.mod` has zero dependencies (`go 1.24.3`, no `require` block). All four libraries above would be the first external dependencies added.

---

## What NOT to Add

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| bubbletea v2 | Alpha/RC state as of Feb 2026. Import path changed to `charm.land/bubbletea/v2`. Breaking changes to `Update()` signature and `View()` API. | v1.3.10 — stable, no breaking changes policy |
| tview (rivo/tview) | Widget-based OOP model conflicts with existing functional rendering style. Full-screen only (no inline mode). Would require complete rewrite of existing `ui/` package. | bubbletea — Elm Architecture, inline + full-screen support |
| termui / gotui | termui is unmaintained. gotui is a fork with 0 stars and unclear maintenance. Neither integrates with charmbracelet ecosystem. | ntcharts (bubbletea-native) |
| pterm | Output-only library, not designed for interactive live-updating dashboards. No event loop. | bubbletea for interactivity |
| go-echarts | Generates HTML, not terminal output. | ntcharts for terminal charts |
| lipgloss v2 | Still in alpha (`v2.0.0-alpha.2`). New compositing API is unstable. | lipgloss v1.1.0 |
| gocui | Minimalist, no charts, requires manual widget management. Less ecosystem support than bubbletea. | bubbletea |
| WebSocket/SSE streaming | Over-engineered for a CLI tool. WakaTime API only updates once per day/hour. | Polling with `tea.Tick` |

---

## Live Update Architecture Pattern

Bubbletea provides two tick primitives:

```go
// tea.Tick — fixed interval relative to program start
// Use this for dashboard refresh (e.g., every 30 seconds)
func doRefresh() tea.Cmd {
    return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
        return refreshMsg(t)
    })
}

// tea.Every — synchronized to system clock
// Use this for "refresh at the top of each minute"
func doClockSync() tea.Cmd {
    return tea.Every(time.Minute, func(t time.Time) tea.Msg {
        return refreshMsg(t)
    })
}
```

The update loop pattern for live data:
```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case refreshMsg:
        return m, tea.Batch(
            fetchDataCmd(),   // async HTTP call
            doRefresh(),      // schedule next tick
        )
    case dataFetchedMsg:
        m.data = msg.data
        return m, nil
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }
    return m, nil
}
```

**Key pattern:** `tea.Batch(fetchDataCmd(), doRefresh())` runs HTTP fetch concurrently with scheduling the next tick. The existing `fetchStats()` and `fetchSummary()` functions in `api.go` wrap cleanly into bubbletea commands.

---

## Terminal Compatibility

Bubbletea and lipgloss both use `muesli/termenv` internally for color detection. Automatic degradation:
- TrueColor (24-bit) → 256 colors → 16 colors → no color
- `NO_COLOR` environment variable honored automatically
- The existing `colorsShouldBeEnabled()` check in `main.go` remains valid for static mode

Cross-platform: bubbletea v1 supports macOS and Linux (target platforms). Windows support exists but is secondary given project context.

---

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| TUI framework | bubbletea v1.3.10 | tview | tview is OOP/widget-based, full-screen only, clashes with existing code style |
| TUI framework | bubbletea v1.3.10 | bubbletea v2 | v2 still RC/alpha, breaking API changes, import path churn |
| Charts | ntcharts | hand-rolled (existing code) | ntcharts gives sparklines for free; existing heatmap/bar code is fallback if needed |
| Charts | ntcharts | asciigraph | asciigraph is line-graph only, no heatmap or bar chart |
| Styling | lipgloss v1.1.0 | raw ANSI (existing) | lipgloss handles terminal width, color degradation, responsive layout automatically |
| Polling | tea.Tick | time.Ticker in goroutine | Direct goroutine use bypasses bubbletea's message loop, causes race conditions |

---

## Integration Path with Existing Code

The existing `ui/` package can coexist with bubbletea during migration:

1. `ui/colors.go` — Replace with lipgloss style definitions (lipgloss auto-detects color support, removing the `colorsShouldBeEnabled()` concern)
2. `ui/graph.go` + `ui/heatmap.go` — Port to ntcharts OR keep as-is and call from bubbletea `View()` methods (existing functions return `[]string`, which composes trivially into views)
3. `ui/card.go` + `ui/render.go` — Replace `cardify()` + `printLeftRight()` with lipgloss `lipgloss.JoinHorizontal` / `lipgloss.JoinVertical` and border styles
4. `getTerminalCols()` — Replace with `tea.WindowSizeMsg` (bubbletea delivers window dimensions automatically)

The static CLI mode (no dashboard flag) can remain unchanged. Bubbletea only activates when `--dashboard` or `-d` flag is used.

---

## Sources

**HIGH CONFIDENCE (Official Documentation):**
- [Bubbletea GitHub](https://github.com/charmbracelet/bubbletea) — v1.3.10, Sep 17 2025
- [Bubbletea pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbletea) — `tea.Tick`, `tea.Every`, `tea.Batch` documented
- [Bubbles v1.0.0 Release](https://github.com/charmbracelet/bubbles/releases) — Feb 10, 2025
- [Lipgloss v1.1.0](https://github.com/charmbracelet/lipgloss/releases) — Mar 13, 2025
- [Bubbletea v2 Discussion](https://github.com/charmbracelet/bubbletea/discussions/1374) — Breaking changes confirmed, Mar 26 2025

**MEDIUM CONFIDENCE (Verified Multi-Source):**
- [ntcharts GitHub](https://github.com/NimbleMarkets/ntcharts) — 632 stars, 121 commits, heatmap/sparkline/bar confirmed. No formal releases — pin by commit.
- [Polling pattern article](https://m3talsmith.medium.com/handling-polling-in-bubbletea-for-go-b17185835549) — Dec 2025, confirms tea.Batch + tick pattern
- [System monitor TUI tutorial](https://penchev.com/posts/create-tui-with-go/) — htop-style pattern with bubbletea

**LOW CONFIDENCE (Needs Validation):**
- ntcharts maintenance trajectory: small team (4 contributors), no tagged releases. Validate the heatmap output quality matches the existing custom heatmap before committing.
