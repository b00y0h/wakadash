# Phase 5: TUI Foundation - Research

**Researched:** 2026-02-19
**Domain:** Bubbletea TUI framework, Lipgloss styling, async architecture in Go
**Confidence:** HIGH

## Summary

Phase 5 builds an async full-screen dashboard using the charmbracelet/bubbletea framework. Bubbletea implements The Elm Architecture (Model-Update-View) where all I/O runs in goroutines via `tea.Cmd`, messages flow back into `Update()`, and `View()` is a pure render function. This architecture is the correct approach for DASH-05 and makes async fetching natural.

The standard approach is: `tea.WithAltScreen()` for full-screen mode (not a command — a program option), `tea.Tick()` for periodic refresh, a `tea.Cmd` closure for async WakaTime API calls, and lipgloss for layout/styling. The bubbles package provides ready-made spinner, help, and key-binding components.

No CONTEXT.md exists. Research drives all recommendations.

**Primary recommendation:** Use bubbletea v1.3.10 + bubbles v1.0.0 + lipgloss v1.1.0 as a unit. Avoid v2 (still RC, breaking API). Wire altscreen via `WithAltScreen()` program option, fetch via `tea.Cmd`, refresh via `tea.Tick()` self-loop, help overlay via boolean flag in model.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/charmbracelet/bubbletea | v1.3.10 | TUI event loop, Elm Architecture | De facto standard Go TUI; battle-tested, stable v1 |
| github.com/charmbracelet/lipgloss | v1.1.0 | Terminal styling, layout, borders | Companion to bubbletea; CSS-like API for terminal |
| github.com/charmbracelet/bubbles | v1.0.0 | Ready-made components (spinner, help, timer, key) | Prevents hand-rolling common widgets |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| bubbles/spinner | (from bubbles v1.0.0) | Loading indicator during async fetch | Show while API call is in-flight |
| bubbles/help | (from bubbles v1.0.0) | Auto-generated keybinding help view | `?` key to show/hide help overlay |
| bubbles/key | (from bubbles v1.0.0) | Typed key bindings with help text | Declare all keybindings as structured data |
| bubbles/timer | (from bubbles v1.0.0) | Countdown to next refresh | Shows time-until-next-refresh in status bar |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| bubbletea v1.3.10 | bubbletea v2 RC | v2 is rc.2 only (Nov 2024), breaking API, new module path charm.land/bubbletea/v2 — avoid until stable |
| bubbles/timer | tea.Tick self-loop | tea.Tick is simpler for basic countdown; bubbles/timer is better for user-controllable countdown |
| lipgloss v1.1.0 | lipgloss v2 alpha | v2 alpha — not stable; v1.1.0 is production-ready |

### Installation
```bash
cd /workspace/wakadash
go get github.com/charmbracelet/bubbletea@v1.3.10
go get github.com/charmbracelet/bubbles@v1.0.0
go get github.com/charmbracelet/lipgloss@v1.1.0
```

**Go version note:** bubbletea v1.3.10 requires `go 1.24.0` in go.mod (confirmed by running `go get`). The current go.mod has `go 1.21.0` — this needs updating to `go 1.24.2` (minimum for bubbles v1.0.0). The CI workflow uses `go-version: stable` so it will work fine in CI. Local environment has go 1.21.0 but Go toolchain management handles this automatically.

## Architecture Patterns

### Recommended Project Structure
```
wakadash/
├── cmd/wakadash/
│   └── main.go              # Entrypoint: config load, tea.NewProgram, WithAltScreen
├── internal/
│   ├── api/
│   │   └── client.go        # Already exists - WakaTime API client
│   ├── config/
│   │   └── config.go        # Already exists - ~/.wakatime.cfg loader
│   ├── tui/
│   │   ├── model.go         # tea.Model: state, Init, Update, View
│   │   ├── messages.go      # Custom tea.Msg types
│   │   ├── commands.go      # tea.Cmd factories (fetchStats, tick)
│   │   ├── keymap.go        # key.Binding declarations for help integration
│   │   └── styles.go        # lipgloss style definitions
│   └── types/
│       └── types.go         # Already exists - API response types
```

### Pattern 1: The Elm Architecture Model
**What:** All application state lives in one struct. View is a pure function of state. Mutations only happen in Update.
**When to use:** Always — this is bubbletea's core constraint.

```go
// Source: https://pkg.go.dev/github.com/charmbracelet/bubbletea
type model struct {
    // Layout
    width  int
    height int

    // Data
    stats     *types.StatsResponse
    loading   bool
    err       error
    lastFetch time.Time

    // Refresh countdown
    refreshInterval time.Duration
    nextRefresh     time.Time

    // UI components
    spinner  spinner.Model
    help     help.Model
    keys     keymap

    // State flags
    showHelp bool
    quitting bool
}
```

### Pattern 2: Async Fetch via tea.Cmd
**What:** Network I/O runs in a goroutine. Result flows back as a message to Update().
**When to use:** Any blocking I/O — API calls, file reads. Never in Update() or View().

```go
// Source: https://github.com/charmbracelet/bubbletea/blob/main/examples/http/main.go
// and https://pkg.go.dev/github.com/charmbracelet/bubbletea

// Custom message types (messages.go)
type statsFetchedMsg struct {
    stats *types.StatsResponse
}
type fetchErrMsg struct {
    err error
}

// Command factory (commands.go)
func fetchStatsCmd(client *api.Client, rangeStr string) tea.Cmd {
    return func() tea.Msg {
        stats, err := client.FetchStats(rangeStr)
        if err != nil {
            return fetchErrMsg{err: err}
        }
        return statsFetchedMsg{stats: stats}
    }
}

// Handle in Update (model.go)
case statsFetchedMsg:
    m.loading = false
    m.stats = msg.stats
    m.lastFetch = time.Now()
    m.nextRefresh = time.Now().Add(m.refreshInterval)
    return m, scheduleRefresh(m.refreshInterval)

case fetchErrMsg:
    m.loading = false
    m.err = msg.err
    return m, scheduleRefresh(m.refreshInterval)
```

### Pattern 3: Auto-Refresh with tea.Tick Self-Loop
**What:** tea.Tick fires once. To repeat, return another Tick from Update().
**When to use:** Periodic data refresh, countdown display.

```go
// Source: https://pkg.go.dev/github.com/charmbracelet/bubbletea (tea.Tick docs)

type refreshMsg time.Time
type countdownTickMsg time.Time

// Fires once after refreshInterval, triggers re-fetch
func scheduleRefresh(interval time.Duration) tea.Cmd {
    return tea.Tick(interval, func(t time.Time) tea.Msg {
        return refreshMsg(t)
    })
}

// Fires every second for countdown display
func tickEverySecond() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return countdownTickMsg(t)
    })
}

// In Init: kick off both
func (m model) Init() tea.Cmd {
    return tea.Batch(
        fetchStatsCmd(m.client, m.rangeStr),  // immediate first fetch
        m.spinner.Tick,                        // spinner animation
        tickEverySecond(),                     // countdown tick
    )
}

// In Update:
case refreshMsg:
    m.loading = true
    return m, tea.Batch(
        fetchStatsCmd(m.client, m.rangeStr),
        m.spinner.Tick,
    )

case countdownTickMsg:
    return m, tickEverySecond()  // self-loop to keep countdown going
```

### Pattern 4: AltScreen via ProgramOption
**What:** Use `tea.WithAltScreen()` program option — NOT `tea.EnterAltScreen` command.
**When to use:** Always for full-screen apps. The command version has race conditions.

```go
// Source: https://pkg.go.dev/github.com/charmbracelet/bubbletea (WithAltScreen docs)
// "Because commands run asynchronously, EnterAltScreen should not be used in Init.
//  To initialize your program with the altscreen enabled use the WithAltScreen
//  ProgramOption instead."

func main() {
    cfg, err := config.Load()
    if err != nil { ... }

    client := api.New(cfg.APIKey, cfg.APIURL)
    m := newModel(client)

    p := tea.NewProgram(m, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Pattern 5: WindowSizeMsg for Responsive Layout
**What:** bubbletea sends `tea.WindowSizeMsg` on startup and every terminal resize. Store dimensions and use in View().
**When to use:** Always in full-screen dashboards. Required for correct sizing.

```go
// Source: https://pkg.go.dev/github.com/charmbracelet/bubbletea (WindowSizeMsg docs)
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    m.help.Width = msg.Width
    return m, nil
```

**Critical:** View() is called before the first WindowSizeMsg arrives. Initialize `width` and `height` to safe defaults (e.g., 80x24) to avoid divide-by-zero or empty renders.

### Pattern 6: Help Overlay via showHelp Flag
**What:** Toggle a `showHelp bool` in the model. View() switches between dashboard and help rendering.
**When to use:** `?` key press. Simple and clean — no overlay library needed.

```go
// Source: https://pkg.go.dev/github.com/charmbracelet/bubbles/help
// Source: https://pkg.go.dev/github.com/charmbracelet/bubbles/key

// keymap.go
type keymap struct {
    Quit key.Binding
    Help key.Binding
    Refresh key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
    return []key.Binding{k.Help, k.Quit}
}

func (k keymap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Help, k.Quit},
        {k.Refresh},
    }
}

var defaultKeymap = keymap{
    Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
    Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
    Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh now")),
}

// In Update:
case tea.KeyMsg:
    switch {
    case key.Matches(msg, m.keys.Quit):
        m.quitting = true
        return m, tea.Quit
    case key.Matches(msg, m.keys.Help):
        m.showHelp = !m.showHelp
        return m, nil
    case key.Matches(msg, m.keys.Refresh):
        m.loading = true
        return m, tea.Batch(fetchStatsCmd(m.client, m.rangeStr), m.spinner.Tick)
    }

// In View:
func (m model) View() string {
    if m.showHelp {
        return m.renderHelp()
    }
    return m.renderDashboard()
}
```

### Pattern 7: Lipgloss Layout
**What:** Use JoinHorizontal/JoinVertical for layout. Use Width/Height functions to measure rendered content before calculating panel sizes.
**When to use:** Building multi-panel layouts. Avoids hard-coded math that breaks on border changes.

```go
// Source: https://pkg.go.dev/github.com/charmbracelet/lipgloss

var (
    borderStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("62"))

    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("205"))

    dimStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("241"))
)

func (m model) renderDashboard() string {
    // Always account for borders subtracting 2 from dimensions
    statsPanel := borderStyle.
        Width(m.width - 2).
        Height(m.height - lipgloss.Height(m.renderStatusBar()) - 4).
        Render(m.renderStats())

    statusBar := m.renderStatusBar()

    return lipgloss.JoinVertical(lipgloss.Left, statsPanel, statusBar)
}
```

### Anti-Patterns to Avoid
- **Using goroutines directly:** bubbletea's concurrency model IS tea.Cmd. Raw goroutines bypass the event loop and cause race conditions. Use `tea.Cmd` instead.
- **Modifying model outside Update():** The model is value-typed. Only Update() returns a new model. Never mutate from Init() or View().
- **Hard-coded dimensions:** `m.height - 5` breaks when borders/padding change. Measure with `lipgloss.Height(renderedWidget)` instead.
- **tea.EnterAltScreen in Init():** Race condition. Use `tea.WithAltScreen()` program option instead.
- **fmt.Println for debugging:** Interferes with TUI rendering. Use `tea.LogToFile("debug.log", "debug")` instead.
- **Forgetting WindowSizeMsg before first render:** Initialize width/height to 80/24; update on WindowSizeMsg.
- **Panicking in tea.Cmd:** bubbletea recovers panics in the event loop but NOT in Cmd goroutines. Terminal left in broken state. Use explicit error returns via errMsg types.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Loading animation | Custom spinner loop | bubbles/spinner | Handles tick timing, frame advancement, ID disambiguation |
| Keybinding help | Custom help renderer | bubbles/help + bubbles/key | Auto-truncates to terminal width, disabled keys auto-hidden |
| Countdown timer | Manual time.Since math | bubbles/timer or tea.Tick | Race-free, handles pause/resume, ID-disambiguates multiple timers |
| Terminal cleanup | Manual ioctl calls | bubbletea's WithAltScreen + Run() | Run() handles cleanup on exit AND panic recovery |
| Layout sizing | Hand-calculated pixel math | lipgloss.Height/Width() + JoinVertical | Accurate with borders/padding, survives style changes |
| Colored output | ANSI escape codes | lipgloss.Color + styles | Cross-platform, handles TTY detection, 256/truecolor/ANSI fallback |

**Key insight:** Every item that "seems simple" has terminal-specific edge cases (Windows, SSH, different terminal emulators, ANSI support levels). The charm libraries handle all of these.

## Common Pitfalls

### Pitfall 1: View() Called Before WindowSizeMsg
**What goes wrong:** The dashboard renders with zero dimensions. Panels collapse, layout math panics with divide-by-zero.
**Why it happens:** bubbletea calls View() immediately after Init() before the terminal reports its size. WindowSizeMsg arrives slightly later.
**How to avoid:** Initialize `width = 80` and `height = 24` in the model constructor. These defaults get overridden immediately but prevent blank renders.
**Warning signs:** Empty/blank first frame; layout panics on startup.

### Pitfall 2: Panic Inside tea.Cmd Breaks Terminal
**What goes wrong:** If API client panics inside a `tea.Cmd` goroutine, the terminal is left in raw mode (no echo, broken input).
**Why it happens:** bubbletea's panic recovery only covers the event loop, not Cmd goroutines. (Open issue: github.com/charmbracelet/bubbletea/issues/1459)
**How to avoid:** Always return error messages (`fetchErrMsg`) instead of panicking. Use `recover()` in Cmd functions if calling third-party code that might panic.
**Warning signs:** After crash, typed text not echoed. Fix with `reset` in terminal.

### Pitfall 3: Ticker Drift / Double Tickers
**What goes wrong:** Multiple refresh tickers firing simultaneously, causing duplicate API calls.
**Why it happens:** Returning a new tick from both `refreshMsg` and `statsFetchedMsg` creates two active tickers.
**How to avoid:** Only one code path should schedule the next tick. The `statsFetchedMsg` handler schedules refresh; `refreshMsg` handler never returns a new refresh command.
**Warning signs:** API called twice per interval; error messages doubled.

### Pitfall 4: Go Version Mismatch
**What goes wrong:** `go build` fails locally because bubbletea v1.3.10 requires go 1.24.0, but go.mod says `go 1.21.0`.
**Why it happens:** bubbletea v1.3.10's go.mod declares `go 1.24.0`. Go will reject building this with an older toolchain.
**How to avoid:** Update go.mod to `go 1.24.2` (minimum required by bubbles v1.0.0). CI uses `go-version: stable` so it's fine there.
**Warning signs:** `go: github.com/charmbracelet/bubbletea@v1.3.10 requires go >= 1.24.0` error.

### Pitfall 5: AltScreen Content Printed Before Run()
**What goes wrong:** Startup messages appear in the alternate screen, then disappear when bubbletea takes over. Or: content printed after program starts is invisible.
**Why it happens:** Alternate screen is a separate buffer. fmt.Printf in altscreen mode is silently dropped by `p.Printf()`. `tea.LogToFile` is the only output mechanism.
**How to avoid:** Use `tea.LogToFile("debug.log", "debug")` for any debugging. Print user-facing startup errors before calling `p.Run()`.
**Warning signs:** Debug output never appears; startup messages show briefly then vanish.

### Pitfall 6: Value Receivers Break Model Mutations
**What goes wrong:** Updating a field on `m` in Update() has no effect — the model reverts.
**Why it happens:** bubbletea models use value receivers (not pointer receivers) per the Elm Architecture pattern. Mutations must be returned as the first return value of Update().
**How to avoid:** Always `return m, cmd` after mutating `m`. Never store state in `m.somePointer` and mutate through the pointer.
**Warning signs:** State changes disappear on next render.

## Code Examples

Verified patterns from official sources:

### Complete model Init with Batch
```go
// Source: https://pkg.go.dev/github.com/charmbracelet/bubbletea (Batch docs)
func (m model) Init() tea.Cmd {
    return tea.Batch(
        fetchStatsCmd(m.client, "last_7_days"),  // immediate fetch
        m.spinner.Tick,                           // start spinner
        tickEverySecond(),                        // start countdown
    )
}
```

### Complete Update skeleton
```go
// Source: https://pkg.go.dev/github.com/charmbracelet/bubbletea (Model interface docs)
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.help.Width = msg.Width
        return m, nil

    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Quit):
            return m, tea.Quit
        case key.Matches(msg, m.keys.Help):
            m.showHelp = !m.showHelp
        case key.Matches(msg, m.keys.Refresh):
            m.loading = true
            return m, tea.Batch(fetchStatsCmd(m.client, m.rangeStr), m.spinner.Tick)
        }
        return m, nil

    case statsFetchedMsg:
        m.loading = false
        m.stats = msg.stats
        m.err = nil
        m.lastFetch = time.Now()
        return m, scheduleRefresh(m.refreshInterval)

    case fetchErrMsg:
        m.loading = false
        m.err = msg.err
        return m, scheduleRefresh(m.refreshInterval)

    case refreshMsg:
        m.loading = true
        return m, tea.Batch(fetchStatsCmd(m.client, m.rangeStr), m.spinner.Tick)

    case countdownTickMsg:
        return m, tickEverySecond()

    case spinner.TickMsg:
        var cmd tea.Cmd
        m.spinner, cmd = m.spinner.Update(msg)
        return m, cmd
    }

    return m, nil
}
```

### Spinner setup
```go
// Source: https://pkg.go.dev/github.com/charmbracelet/bubbles/spinner
s := spinner.New()
s.Spinner = spinner.Dot
s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
```

### Help model setup
```go
// Source: https://pkg.go.dev/github.com/charmbracelet/bubbles/help
h := help.New()
h.Width = 80  // updated on WindowSizeMsg
```

### Debug logging (use instead of fmt.Println)
```go
// Source: https://pkg.go.dev/github.com/charmbracelet/bubbletea (LogToFile docs)
if os.Getenv("DEBUG") != "" {
    f, err := tea.LogToFile("debug.log", "debug")
    if err != nil {
        fmt.Fprintln(os.Stderr, "fatal:", err)
        os.Exit(1)
    }
    defer f.Close()
}
```

### Lipgloss status bar with countdown
```go
// Source: https://pkg.go.dev/github.com/charmbracelet/lipgloss
func (m model) renderStatusBar() string {
    var status string
    if m.loading {
        status = m.spinner.View() + " fetching..."
    } else if m.err != nil {
        status = "error: " + m.err.Error()
    } else {
        remaining := time.Until(m.nextRefresh).Round(time.Second)
        status = fmt.Sprintf("last updated: %s  next: %s",
            m.lastFetch.Format("15:04:05"),
            remaining,
        )
    }
    helpHint := dimStyle.Render("? help  q quit")
    gap := strings.Repeat(" ", max(0, m.width-lipgloss.Width(status)-lipgloss.Width(helpHint)))
    return dimStyle.Render(status) + gap + helpHint
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| tea.EnterAltScreen as Cmd in Init | tea.WithAltScreen() ProgramOption | v0.x to v1.x | Eliminates race condition |
| Manual goroutines for async | tea.Cmd returning tea.Msg | Always the right way | Event loop safe, no races |
| termbox-go | bubbletea | 2020+ | bubbletea is now de facto standard |
| bubbletea v2 (charm.land/bubbletea/v2) | bubbletea v1.3.x | v2 still RC as of Nov 2024 | Use v1; v2 has breaking API, new module path |

**Deprecated/outdated:**
- `tea.EnterAltScreen` as a command: Still exists but causes race conditions; use `WithAltScreen()` instead.
- Direct goroutine spawning: Always use `tea.Cmd`. Raw goroutines are not visible to the event loop.
- `termbox-go`: Predecessor; bubbletea supersedes it for interactive TUIs.

## Open Questions

1. **Configurable refresh interval**
   - What we know: Requirements say "configurable interval (default 60s)"
   - What's unclear: Via flag? Via config file? Interactive `+/-` keys?
   - Recommendation: Add `--refresh` flag to `wakadash` command in main.go; pass to model. Start with flag only, config file in a later phase.

2. **Which WakaTime API endpoint to use for dashboard**
   - What we know: Both `/stats/{range}` and `/summaries` exist. `/stats` is simpler for totals; `/summaries` gives daily breakdown.
   - What's unclear: Phase 5 says "basic stats display" — which stats?
   - Recommendation: Use `/stats/last_7_days` as default for phase 5 (simplest, matches what users expect). Match `wakafetch` defaults.

3. **go.mod version update**
   - What we know: bubbletea v1.3.10 requires go 1.24.0; bubbles v1.0.0 requires go 1.24.2.
   - What's unclear: Does the team want to pin a specific toolchain or float to latest?
   - Recommendation: Update go.mod to `go 1.24.2`. No `toolchain` directive needed (CI uses stable). Document the change.

## Sources

### Primary (HIGH confidence)
- `pkg.go.dev/github.com/charmbracelet/bubbletea` — Model interface, Batch, Sequence, Tick, Every, WithAltScreen, WindowSizeMsg, LogToFile (verified by reading official docs)
- `pkg.go.dev/github.com/charmbracelet/lipgloss` — NewStyle, JoinHorizontal/Vertical, Width/Height, Border, Color (verified by reading official docs)
- `pkg.go.dev/github.com/charmbracelet/bubbles` — spinner, help, key, timer sub-packages (verified by reading official docs)
- `github.com/charmbracelet/bubbletea/blob/main/examples/` — altscreen-toggle, timer, http, spinner, fullscreen examples (verified by fetching raw source)
- `go get github.com/charmbracelet/bubbletea@latest` — confirmed v1.3.10, go 1.24.0 requirement (verified by running in environment)
- `go get github.com/charmbracelet/bubbles@latest` — confirmed v1.0.0, go 1.24.2 requirement (verified by running in environment)

### Secondary (MEDIUM confidence)
- `github.com/charmbracelet/bubbletea/releases` — version timeline, v2 RC status (fetched from releases page)
- `leg100.github.io/en/posts/building-bubbletea-programs/` — layout pitfalls, lipgloss measurement patterns (verified against official docs)

### Tertiary (LOW confidence)
- WebSearch results on pitfalls (panic in Cmd, mouse tracking cleanup) — cross-referenced with GitHub issues #1459 and v1.3.10 notes

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — versions confirmed by `go get` in environment
- Architecture: HIGH — verified against official pkg.go.dev docs and example source
- Pitfalls: HIGH (panic/terminal) / MEDIUM (ticker drift) — terminal panic verified via GitHub issues; ticker pitfall from community sources

**Research date:** 2026-02-19
**Valid until:** 2026-05-19 (90 days — v1 stable, low churn expected; watch for bubbles v1.x patch releases)
