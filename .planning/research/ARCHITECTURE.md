# Architecture Patterns: wakadash Live Terminal Dashboard

**Domain:** Live-updating terminal dashboard (TUI) added to existing Go CLI
**Researched:** 2026-02-17
**Confidence:** HIGH (Bubbletea API verified via pkg.go.dev official docs; existing code read directly)

## Context: Adding Dashboard to Existing CLI

wakafetch already has a working CLI with the following architecture:

```
main.go           -- flag parsing, API routing
config.go         -- reads ~/.wakatime.cfg (api_url, api_key)
api.go            -- fetchStats(), fetchSummary(), fetchApi[T]() generic HTTP
flags.go          -- flag definitions (flag stdlib, no cobra)
ui/
  display.go      -- DisplayStats(), DisplaySummary(), DisplayBreakdown(), DisplayHeatmap()
  render.go       -- render(), renderCardSection(), printLeftRight()
  card.go         -- cardify() -- border box drawing with ANSI strings
  graph.go        -- graphStr() -- bar chart rendering
  heatmap.go      -- heatmap() -- activity heatmap
  breakdown.go    -- dailyBreakdownStr()
  colors.go       -- ANSI color helpers
  utils.go        -- timeFmt(), topItemName()
types/
  types.go        -- StatsResponse, SummaryResponse, StatItem, DayData
```

Key observations that drive architecture decisions:

1. `api.go` is already pure and reusable — `fetchStats()` and `fetchSummary()` return typed structs, no I/O side effects
2. `ui/` functions produce `[]string` slices, not rendered terminal output directly — they are composable
3. `card.go` and `graph.go` already implement the visual components needed for a dashboard panel
4. `getTerminalCols()` uses `stty` subprocess — must be replaced with Bubbletea's `WindowSizeMsg` for the dashboard
5. Config loading (`parseConfig()`) is already decoupled — reuse directly

---

## Recommended Architecture

### Framework: Bubbletea + Lipgloss

Use **charmbracelet/bubbletea v1.x** (current: v1.3.10, released Sep 2025) as the TUI runtime. It implements the Elm Architecture (Model-Update-View), which is the standard for Go TUI dashboards. Use **charmbracelet/lipgloss** for layout composition (replacing the manual `printLeftRight()` approach).

Do NOT use tview, termui, or tcell directly. Bubbletea is the dominant Go TUI framework with 9,300+ dependents, active maintenance, and the best documented composition model. (HIGH confidence — official docs + ecosystem search)

---

## System Structure

```
wakafetch/
├── main.go                    -- MODIFIED: add --dashboard / -w flag routing
├── dashboard/                 -- NEW package
│   ├── model.go               -- Root Bubbletea model (state + Update + View)
│   ├── messages.go            -- Message type definitions
│   ├── fetch.go               -- tea.Cmd wrappers around existing api.go calls
│   └── layout.go              -- Panel composition using lipgloss
├── ui/                        -- UNCHANGED (reused as rendering functions)
│   └── ... (existing files)
├── api.go                     -- UNCHANGED
├── config.go                  -- UNCHANGED
├── types/                     -- UNCHANGED
└── flags.go                   -- MODIFIED: add dashboard flag
```

The `dashboard/` package is entirely new. The existing `ui/`, `api.go`, `config.go`, and `types/` packages are reused as-is. This is the minimum-change integration strategy.

---

## Component Boundaries

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| `dashboard/model.go` | Root model: holds all dashboard state, routes messages to panels | `dashboard/fetch.go` (via Cmd), `dashboard/layout.go` (via View call) |
| `dashboard/messages.go` | Defines all Msg types: `TickMsg`, `StatsDataMsg`, `SummaryDataMsg`, `ErrMsg`, `RefreshMsg` | Consumed by `model.go` Update() |
| `dashboard/fetch.go` | Wraps `api.fetchStats()` and `api.fetchSummary()` as `tea.Cmd` functions | Calls existing `api.go`; returns typed Msg structs |
| `dashboard/layout.go` | Composes `ui/` rendering functions into lipgloss-joined panel strings | Calls existing `ui/graph.go`, `ui/card.go`, `ui/heatmap.go` |
| `ui/` (existing) | Produces `[]string` representations of data — unchanged | Called by `dashboard/layout.go` |
| `api.go` (existing) | HTTP fetching — unchanged | Called by `dashboard/fetch.go` |
| `config.go` (existing) | Reads `~/.wakatime.cfg` — unchanged | Called from `main.go` dashboard entrypoint |

---

## Data Flow for Live Updates

```
App Start
    │
    ▼
main.go: loadAPIConfig() → apiURL, apiKey
    │
    ▼
tea.NewProgram(model, tea.WithAltScreen())
    │
    ▼
model.Init()
    │── returns: tea.Batch(fetchStatsCmd(), tickCmd())
    │
    ▼
goroutine: fetchStatsCmd() runs api.fetchStats()  [does not block UI]
goroutine: tickCmd() fires after interval
    │
    ▼
Update(StatsDataMsg{data}) → model.stats = data → returns tickCmd()
    │
    ▼
View() → layout.Render(model) → lipgloss.JoinVertical(panels...)
    │
    ▼
[on next tick]
Update(TickMsg) → if time.Since(lastFetch) >= interval → returns fetchStatsCmd()
    │
    ▼
[loop continues]
```

### Refresh Strategy

Use `tea.Tick()` with a configurable interval (default: 5 minutes, matching WakaTime's heartbeat aggregation cadence). When the tick fires, issue a new fetch command. Do NOT use `tea.Every()` for API calls — `tea.Tick()` gives you a message to check elapsed time, avoiding stacked fetches if a request takes longer than the interval.

```go
// messages.go
type TickMsg time.Time
type StatsDataMsg struct{ Data *types.StatsResponse }
type SummaryDataMsg struct{ Data *types.SummaryResponse }
type ErrMsg struct{ Err error }

// fetch.go
func fetchStatsCmd(apiKey, apiURL, rangeStr string) tea.Cmd {
    return func() tea.Msg {
        data, err := fetchStats(apiKey, apiURL, rangeStr)
        if err != nil {
            return ErrMsg{Err: err}
        }
        return StatsDataMsg{Data: data}
    }
}

func tickCmd(interval time.Duration) tea.Cmd {
    return tea.Tick(interval, func(t time.Time) tea.Msg {
        return TickMsg(t)
    })
}
```

```go
// model.go Update()
case TickMsg:
    if time.Since(m.lastFetch) >= m.refreshInterval {
        return m, tea.Batch(fetchStatsCmd(...), tickCmd(m.refreshInterval))
    }
    return m, tickCmd(m.refreshInterval)

case StatsDataMsg:
    m.statsData = msg.Data
    m.lastFetch = time.Now()
    m.loading = false
    return m, nil
```

---

## Layout Pattern

The existing `printLeftRight()` + ANSI string approach works for one-shot CLI output but cannot be used in Bubbletea's `View()` because:
- It calls `fmt.Println()` directly (side effect, not allowed in View())
- Terminal width is fetched via `stty` subprocess (not available in TUI context)

Replace with lipgloss composition in `dashboard/layout.go`:

```go
// layout.go
func RenderDashboard(m Model) string {
    termWidth := m.width   // from WindowSizeMsg
    termHeight := m.height

    // Reuse existing rendering functions, they return []string
    langLines, langW := ui.GraphStr(m.statsData.Data.Languages, 8)
    statsLines, statsW := ui.FieldsStr(m.heading, m.statsFields)

    // Convert []string to single string for lipgloss
    langPanel := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("240")).
        Render(strings.Join(langLines, "\n"))

    statsPanel := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("240")).
        Render(strings.Join(statsLines, "\n"))

    body := lipgloss.JoinHorizontal(lipgloss.Top, statsPanel, langPanel)
    statusBar := renderStatusBar(termWidth, m.lastFetch, m.loading)

    return lipgloss.JoinVertical(lipgloss.Left, body, statusBar)
}
```

This requires making `ui/` functions exported (capitalizing `graphStr` → `GraphStr`, `fieldsStr` → `FieldsStr`, etc.). That is the only change needed in the existing `ui/` package.

---

## New vs Modified Components

### New (must build from scratch)

| Component | File | What it does |
|-----------|------|--------------|
| Dashboard model | `dashboard/model.go` | Bubbletea Model interface: holds all state, routes all messages |
| Message types | `dashboard/messages.go` | Typed Msg values for all async events |
| Fetch commands | `dashboard/fetch.go` | Wraps api.go calls as tea.Cmd |
| Layout composer | `dashboard/layout.go` | Turns `[]string` panels into lipgloss-joined view string |
| Status bar | `dashboard/layout.go` | Shows last refresh time, refresh interval, loading state |

### Modified (minimal changes)

| Component | File | Change |
|-----------|------|--------|
| Entry point | `main.go` | Add `--dashboard` / `-w` flag branch, call `tea.NewProgram(...)` |
| Flag definitions | `flags.go` | Add `dashboardFlag`, `intervalFlag` |
| ui/ exports | `ui/*.go` | Capitalize function names: `graphStr` → `GraphStr`, `fieldsStr` → `FieldsStr`, `cardify` → `Cardify` |
| Terminal width | `ui/render.go` | Remove `getTerminalCols()` usage in dashboard path; pass width via model |

### Unchanged (zero modification needed)

- `api.go` — already clean, just called from fetch.go
- `config.go` — already clean, called from main.go
- `types/types.go` — already clean
- `ui/heatmap.go`, `ui/graph.go`, `ui/card.go`, `ui/colors.go`, `ui/utils.go` — logic unchanged, only export visibility changes

---

## Patterns to Follow

### Pattern 1: AltScreen for Dashboard Mode

Always enter AltScreen for full-window dashboard. This keeps the original terminal session clean and allows `q` to quit and restore original terminal state.

```go
p := tea.NewProgram(
    initialModel(apiURL, apiKey, config),
    tea.WithAltScreen(),
)
```

### Pattern 2: WindowSizeMsg for Responsive Layout

Bubbletea sends `tea.WindowSizeMsg` on startup and on every terminal resize. Store dimensions in the model and use them in View() for responsive layout. Never call `stty` or `exec.Command` inside a Bubbletea program.

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    return m, nil
```

### Pattern 3: Loading State During Fetch

Show a spinner or "Loading..." indicator while the first fetch is in flight. The model starts with `loading: true`; set to `false` when `StatsDataMsg` arrives.

```go
func (m Model) View() string {
    if m.loading {
        return "Loading..."  // or a spinner component from charmbracelet/bubbles
    }
    return RenderDashboard(m)
}
```

### Pattern 4: Error Display Without Crashing

API errors should display in the status bar and schedule a retry, not crash the program. Return `ErrMsg` from fetch commands; store in model; display in status bar; retry on next tick.

```go
case ErrMsg:
    m.lastErr = msg.Err
    m.loading = false
    return m, tickCmd(m.refreshInterval)
```

### Pattern 5: Keyboard Navigation

Minimal keybindings for MVP: `q`/`ctrl+c` to quit, `r` to manual refresh, `?` to toggle help. Do not build a full interactive UI in the first milestone.

```go
case tea.KeyMsg:
    switch msg.String() {
    case "q", "ctrl+c":
        return m, tea.Quit
    case "r":
        m.loading = true
        return m, fetchStatsCmd(m.apiKey, m.apiURL, m.rangeStr)
    }
```

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Goroutines Inside Update()

**What:** Spawning `go func()` directly inside Update() to do async work
**Why bad:** Race conditions on model state, messages can arrive out of order, Bubbletea's event loop is not goroutine-safe for model mutation
**Instead:** Always return a `tea.Cmd` from Update(). Bubbletea manages the goroutine for you.

### Anti-Pattern 2: fmt.Println Inside View()

**What:** Calling `fmt.Println()` or writing to stdout inside View()
**Why bad:** View() must return a `string`. Side effects in View() corrupt the terminal rendering. The existing `ui/` functions that call `printStrs()` cannot be called from View() — wrap them to return strings instead.
**Instead:** Make `ui/` functions return `string` or `[]string` and call `strings.Join()` in View().

### Anti-Pattern 3: Blocking HTTP in Update()

**What:** Making an HTTP request directly inside Update()
**Why bad:** Update() blocks the event loop. The entire TUI freezes until the request completes.
**Instead:** Return a `tea.Cmd` from Update() that wraps the HTTP call. It runs in its own goroutine.

### Anti-Pattern 4: Calling stty or exec.Command for Terminal Size

**What:** Using the existing `getTerminalCols()` function (which runs `stty -F /dev/tty size`) inside the dashboard
**Why bad:** In AltScreen mode, /dev/tty handling is managed by Bubbletea. The subprocess approach is fragile and cross-platform unreliable. Also, it does not react to terminal resize events.
**Instead:** Handle `tea.WindowSizeMsg` and store `m.width` / `m.height` in the model.

### Anti-Pattern 5: Sharing Mutable State Between Cmd and Model

**What:** Using pointers to model fields inside a `tea.Cmd` closure
**Why bad:** The Cmd runs in a goroutine. If the model is replaced (as Bubbletea does on every Update call), the pointer may reference stale data.
**Instead:** Capture only plain values (strings, ints) in Cmd closures, not model references.

---

## Suggested Build Order

Dependencies flow strictly from bottom to top. Build in this sequence:

```
Phase 1: Foundation (no external dependencies)
    dashboard/messages.go      -- pure type definitions
    dashboard/fetch.go         -- wrap api.go as tea.Cmd (depends only on existing api.go + types/)

Phase 2: Layout Plumbing
    ui/ export visibility      -- capitalize graphStr, fieldsStr, cardify (one-line changes)
    dashboard/layout.go        -- lipgloss composition (depends on ui/ exports + lipgloss)

Phase 3: Root Model
    dashboard/model.go         -- Bubbletea Model (depends on messages.go, fetch.go, layout.go)

Phase 4: Entrypoint
    flags.go                   -- add --dashboard flag
    main.go                    -- dashboard mode routing, tea.NewProgram(...)

Phase 5: Polish
    Status bar with last-refresh time, loading indicator, error display
    Manual refresh keybinding (r)
    Help overlay (?)
    Configurable refresh interval (--interval flag)
```

Each phase produces a runnable/testable artifact before the next phase begins.

---

## Scalability Considerations

| Concern | MVP (single panel) | Multi-panel | Future |
|---------|-------------------|-------------|--------|
| Terminal width | Hardcode 2-column layout | Use `m.width` for responsive breakpoints | Drag-to-resize panels |
| Refresh rate | Fixed 5m interval | User-configurable `--interval` | Per-panel refresh rates |
| Data staleness | Single stats fetch | Batch: stats + summary in parallel with `tea.Batch()` | Background cache with delta updates |
| Panel count | 2 panels (stats + languages) | 4 panels (+ heatmap + projects) | Plugin panel architecture |
| Error recovery | Show error in status bar | Exponential backoff retry | Circuit breaker per endpoint |

---

## Dependencies to Add

```bash
go get github.com/charmbracelet/bubbletea@v1.3.10
go get github.com/charmbracelet/lipgloss@latest
```

Optional (defer until needed):
```bash
go get github.com/charmbracelet/bubbles@latest  # for spinner, viewport components
```

The existing `ui/` package uses raw ANSI escape codes for colors. Lipgloss will handle layout colors. These coexist safely — lipgloss does not conflict with existing ANSI output in the `[]string` returns from `ui/` functions.

---

## Sources

### Official Documentation (HIGH confidence)
- [Bubbletea API — pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbletea) — v1.3.10, current API surface, WindowSizeMsg, tea.Tick, tea.Every, tea.Batch, tea.Sequence, WithAltScreen
- [Lipgloss — charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) — JoinHorizontal, JoinVertical, Style, border patterns
- [Bubbletea Commands Tutorial](https://github.com/charmbracelet/bubbletea/blob/main/tutorials/commands/README.md) — official async command pattern

### Architecture Articles (MEDIUM confidence, corroborated by official docs)
- [Tips for Building Bubble Tea Programs — leg100.github.io](https://leg100.github.io/en/posts/building-bubbletea-programs/) — tree-of-models composition, Update/View performance, message ordering
- [Build a System Monitor TUI in Go — penchev.com](https://penchev.com/posts/create-tui-with-go/) — tea.Every refresh pattern, gopsutil data fetching, lipgloss layout for dashboard
- [Building a Terminal IRC Client with Bubble Tea — sngeth.com](https://sngeth.com/go/terminal/ui/bubble-tea/2025/08/17/building-terminal-ui-with-bubble-tea/) — custom message types, real-time update patterns

### Existing Codebase (HIGH confidence — read directly)
- `/workspace/wakafetch/api.go` — `fetchStats()`, `fetchSummary()`, `fetchApi[T]()` — confirmed reusable
- `/workspace/wakafetch/ui/render.go` — `render()`, `getTerminalCols()`, `printLeftRight()` — confirmed must change for TUI
- `/workspace/wakafetch/ui/graph.go` — `graphStr()` returns `[]string` — confirmed composable
- `/workspace/wakafetch/ui/card.go` — `cardify()` returns `[]string` — confirmed composable
- `/workspace/wakafetch/types/types.go` — `StatsResponse`, `SummaryResponse` — confirmed data contracts

---
*Architecture research for: wakadash live terminal dashboard milestone*
*Researched: 2026-02-17*
