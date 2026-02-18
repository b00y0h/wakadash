# Domain Pitfalls

**Domain:** Live-updating terminal dashboard in Go (TUI milestone for existing CLI)
**Researched:** 2026-02-17
**Confidence:** HIGH for concurrency/API pitfalls (official docs + multiple corroborating sources), MEDIUM for terminal compatibility specifics

---

## Critical Pitfalls

Mistakes that cause rewrites or leave the terminal in unusable state.

---

### Pitfall 1: Mutating TUI Model State Outside the Event Loop

**What goes wrong:**
Race conditions corrupt UI state. Data renders partially updated or crashes. The `-race` detector fires, but the crashes are non-deterministic and hard to reproduce. In Bubble Tea, this typically appears as flicker, corrupted display, or panics in `View()`.

**Why it happens:**
When an API fetch goroutine writes directly to a shared model struct while the Bubble Tea event loop is calling `View()`, two goroutines access the same memory concurrently. The pattern "start goroutine, write to model when done" is natural to Go developers but violates Bubble Tea's single-owner model.

**Consequences:**
- Non-deterministic crashes in production but not in tests
- Corrupted display output that looks like a rendering bug (but isn't)
- Full rewrite of state management once the root cause is found

**Prevention:**
Never mutate model state from outside `Update()`. All API responses must arrive as `tea.Msg` through the event loop:

```go
// WRONG: goroutine writes to model directly
go func() {
    data, _ := fetchStats()
    model.data = data  // race condition
}()

// CORRECT: goroutine sends a message; Update() handles it
type statsMsg struct{ data *types.StatsResponse }

func fetchStatsCmd() tea.Cmd {
    return func() tea.Msg {
        data, err := fetchStats()
        if err != nil {
            return errMsg{err}
        }
        return statsMsg{data}
    }
}
```

**Detection:**
- Run with `go run -race .` during development
- Any non-deterministic display corruption
- `Update()` not being the only place model fields are assigned

**Phase to address:**
Phase 1 (TUI Foundation) — The architecture decision. Getting this wrong means rewriting state flow later.

---

### Pitfall 2: Blocking the Event Loop with Synchronous API Calls

**What goes wrong:**
The dashboard freezes completely during API fetches. Keyboard input stops responding. The terminal appears hung. Users Ctrl+C to escape and the terminal may be left in raw mode.

**Why it happens:**
The existing `fetchStats()` and `fetchSummary()` functions in the codebase make synchronous HTTP calls with a 10-second timeout. If these are called directly in `Update()` or `Init()` rather than wrapped in a `tea.Cmd`, they block the single-threaded event loop.

**Consequences:**
- Complete UI freeze for the duration of each API call (potentially 10 seconds)
- No spinner or loading indicator is possible because rendering is also blocked
- If the network is slow or WakaTime returns a 302 instead of 429, the freeze extends

**Prevention:**
Every API call must be a `tea.Cmd` (runs in a goroutine, returns a `tea.Msg`):

```go
// WRONG: called synchronously in Init() or Update()
func (m Model) Init() tea.Cmd {
    data, _ := fetchStats(m.apiKey, m.apiURL, m.rangeStr)  // blocks
    m.data = data
    return nil
}

// CORRECT: deferred to goroutine via Cmd
func (m Model) Init() tea.Cmd {
    return fetchStatsCmd(m.apiKey, m.apiURL, m.rangeStr)
}
```

**Detection:**
- UI freezes when refresh happens
- No ability to quit with `q` or `Ctrl+C` during data load

**Phase to address:**
Phase 1 (TUI Foundation) — The wrapping of existing API calls must be part of initial TUI wiring.

---

### Pitfall 3: Panic Leaves Terminal in Raw Mode

**What goes wrong:**
Any unrecovered panic in a command goroutine (API call, ticker callback, etc.) leaves the terminal in raw mode. The cursor disappears, typed characters don't echo, the user cannot use the terminal at all. They must run `reset` to restore it.

**Why it happens:**
Bubble Tea puts the terminal into raw mode at startup. If a panic occurs in a goroutine that Bubble Tea spawned for a `tea.Cmd`, that goroutine's panic handler is separate from the main program panic handler. Panics in commands are caught individually — but only if Bubble Tea's `WithoutCatchPanics` option has NOT been set.

The existing codebase uses `os.Exit(0)` in `showCustomHelp()`. If similar patterns are used inside TUI commands, the terminal cleanup is bypassed.

**Consequences:**
- Unusable terminal session requiring `reset`
- User has to know to run `reset` — many don't
- Can corrupt ongoing terminal multiplexer sessions (tmux, screen)

**Prevention:**
1. Never call `os.Exit()` from within a running Bubble Tea program
2. Keep the default `CatchPanics` behavior enabled (do NOT use `tea.WithoutCatchPanics()`)
3. Ensure cleanup on SIGINT/SIGTERM: use `tea.WithAltScreen()` so the terminal restores on exit
4. Test panic recovery explicitly during development

**Detection:**
- Terminal cursor disappears after program exits
- Typed characters don't appear in the terminal after closing the dashboard
- `stty -echo` visible in output

**Phase to address:**
Phase 1 (TUI Foundation) — Set up correctly from the start; retrofitting panic safety is error-prone.

---

### Pitfall 4: Ticker-Driven Auto-Refresh Creates Goroutine Leaks

**What goes wrong:**
Each auto-refresh cycle spawns a goroutine that may not be cleaned up. Over time (especially when the user changes range or pauses/unpauses), leaked goroutines accumulate. Memory grows unboundedly. In extreme cases, multiple goroutines race to update the same model state.

**Why it happens:**
The typical implementation uses `time.NewTicker` inside a goroutine and sends messages on each tick. If the Bubble Tea program quits, exits, or the ticker is recreated (e.g., user changes refresh interval), the old ticker goroutine keeps running — Go's GC does not collect running goroutines.

**Consequences:**
- Memory leak in long-running dashboard sessions
- Multiple simultaneous API requests when ticker restarts accumulate
- Hitting WakaTime's rate limit (10 req/s avg over 5 minutes) from leaked refresh goroutines

**Prevention:**
Use Bubble Tea's built-in ticker pattern with context cancellation:

```go
// Bubble Tea's built-in approach: return a Cmd that fires once,
// then requeue after data arrives
func tickCmd(interval time.Duration) tea.Cmd {
    return tea.Tick(interval, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

// In Update(), after handling tickMsg, return the next tick:
case tickMsg:
    return m, tea.Batch(fetchStatsCmd(...), tickCmd(m.refreshInterval))
```

Always cancel contexts and stop tickers in the `quit` message handler.

**Detection:**
- Increasing memory usage over time (visible with `htop`)
- More API calls than expected (exceeding 1 per refresh interval)
- `-race` detector firing after range changes

**Phase to address:**
Phase 2 (Live Refresh) — This pitfall is specific to implementing the auto-refresh loop.

---

### Pitfall 5: WakaTime 202 Response Treated as Error or Success

**What goes wrong:**
The `/stats` endpoint returns HTTP 202 (Accepted) when stats are still being calculated. The existing code only checks for `StatusOK`. A 202 response causes a JSON decode failure or is treated as an unknown error, crashing the dashboard or displaying "server error" to the user.

**Why it happens:**
WakaTime processes stats asynchronously. For free plan users and for time ranges >= 1 year, stats may not be immediately available. The API explicitly returns 202 with a `percent_calculated` field indicating background processing progress. This is documented behavior, not an error.

**Consequences:**
- Dashboard shows error to user when stats are legitimately still computing
- User assumes the dashboard is broken when WakaTime is working correctly
- Aggressive retry logic (to "fix" the apparent error) hammers the API unnecessarily

**Prevention:**
Handle 202 explicitly as a "loading" state with retry:

```go
case http.StatusAccepted:  // 202: stats still computing
    return nil, ErrStatsNotReady  // special sentinel error
// Caller: show "Calculating..." and retry after delay
```

Check `is_up_to_date` and `percent_calculated` fields in the response body when available.

**Detection:**
- "Stats not available" or JSON decode errors on first launch
- Works fine after waiting and refreshing manually

**Phase to address:**
Phase 2 (Live Refresh) — Handle during API client enhancement for dashboard use.

---

### Pitfall 6: Terminal Width Detection Breaks Layout on Resize

**What goes wrong:**
The existing `getTerminalCols()` in `render.go` calls `stty` via `exec.Command` and has no Windows support (returns `fallback = 9999`). In TUI mode, the terminal size must be known at every render, and terminal resize events (SIGWINCH) must update the layout. Without this, the dashboard overflows or mis-aligns on resize.

**Why it happens:**
The current stty-based approach is a one-shot measurement during static rendering. For a live dashboard, the terminal can be resized at any time. Bubble Tea sends a `tea.WindowSizeMsg` whenever the terminal is resized, but the layout logic must be wired to consume it. If the old `getTerminalCols()` approach is reused, it won't receive resize events.

**Consequences:**
- Cards overflow terminal width after resize
- Layout doesn't adapt to narrow vs wide terminals
- `stty` subprocess calls on every render are expensive (measured in milliseconds)

**Prevention:**
Replace `getTerminalCols()` with Bubble Tea's `WindowSizeMsg` approach:

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    return m, nil
```

Use `m.width` in all layout calculations. The stty-based approach must not be used inside the TUI render loop.

**Detection:**
- Resize terminal while dashboard is running — layout breaks
- `stty` subprocess calls appearing in `strace`/`dtruss` output during rendering

**Phase to address:**
Phase 1 (TUI Foundation) — Layout must use `WindowSizeMsg` from the start; retrofitting is a rewrite of the layout layer.

---

## Moderate Pitfalls

---

### Pitfall 7: WakaTime Rate Limit (429) Not Handled Gracefully in Dashboard Mode

**What goes wrong:**
In dashboard mode with frequent auto-refresh, a 429 response causes the existing error handler to surface an error to the user on every tick. The dashboard shows "Rate limit exceeded" repeatedly. The auto-refresh continues hammering the API despite getting 429 responses (no backoff).

**Why it happens:**
The current `fetchApi()` converts 429 to an error string and returns it. In a one-shot CLI, this is correct. In a live dashboard, the refresh ticker keeps firing regardless of previous errors. Without exponential backoff, the dashboard makes the rate limit worse.

**Prevention:**
1. Implement exponential backoff with jitter on 429 responses
2. Check `Retry-After` header before next retry
3. Show "Rate limited — retrying in Xs" in the dashboard status bar instead of an error
4. Minimum refresh interval of 60 seconds (well under 10 req/s averaged over 5 minutes)

**Detection:**
- Console logs showing repeated 429 errors at fixed intervals
- No "retry after" delay between requests

**Phase to address:**
Phase 2 (Live Refresh) — Implement alongside the refresh ticker.

---

### Pitfall 8: ANSI Escape Codes in Non-TTY Output After Adding TUI

**What goes wrong:**
When output is piped (e.g., `wakadash | grep ...`) or redirected, raw ANSI escape codes appear in the output. The existing code already handles this with `colorsShouldBeEnabled()` TTY check. But if the TUI framework is used without the alt-screen buffer, escape sequences appear in the shell history and pipe output.

**Why it happens:**
Adding a TUI framework changes output mode. The existing color guard (`NO_COLOR`, TTY check) only applies to the static render path. If the dashboard accidentally starts in a non-TTY context (e.g., piped, inside scripts), Bubble Tea still emits escape codes.

**Prevention:**
1. Detect non-TTY at startup and fall back to static output mode (existing behavior)
2. Use `tea.WithAltScreen()` so TUI output doesn't pollute the scrollback buffer
3. Preserve the existing `--no-colors` and `NO_COLOR` honor path for the static mode

**Detection:**
- `wakadash --watch | cat` shows escape codes

**Phase to address:**
Phase 1 (TUI Foundation) — Mode detection must be in the entry point before Bubble Tea starts.

---

### Pitfall 9: ANSI Color Code Compatibility Across Terminal Emulators

**What goes wrong:**
The existing code uses `\x1b[38;2;128;128;128m` (24-bit RGB true color for `MidGray`). This works in modern terminals (iTerm2, Windows Terminal, Kitty) but not in older terminals (some SSH clients, older macOS Terminal.app versions, tmux without `terminal-overrides`). In unsupported terminals, the color code renders as literal text or breaks surrounding formatting.

**Why it happens:**
24-bit true color (`\x1b[38;2;R;G;Bm`) is not universally supported. The existing colors.go uses it for `MidGray` but falls back to standard 256-color or 8-color codes for all other colors. This inconsistency can cause visible artifacts in degraded terminals.

**Prevention:**
1. Use 256-color or 8-color codes as the primary palette (already done for most colors)
2. For `MidGray`, fall back to `\x1b[90m` (bright black / dark gray) when true color is unsupported
3. Check `COLORTERM=truecolor` env var before using 24-bit codes
4. Bubble Tea's Lip Gloss handles color degradation automatically — use it instead of manual ANSI strings

**Detection:**
- Test in `tmux` without `set -g terminal-overrides`
- Test with `TERM=xterm` explicitly set
- Literal characters appearing where colors should be

**Phase to address:**
Phase 3 (Polish) — Acceptable to address after core dashboard works. Use Lip Gloss early to avoid retroactive replacement.

---

### Pitfall 10: Adding Cobra Breaks Existing Flag Parsing

**What goes wrong:**
The existing codebase uses the standard `flag` package directly with short (`-r`) and long (`--range`) flag variants registered manually. If Cobra is introduced for the new `--watch` or `--interval` flags, the two flag systems conflict. Cobra's `pflag` library does not coexist cleanly with stdlib `flag` without explicit bridging.

**Why it happens:**
Cobra uses `pflag` (POSIX-compliant flags) while the existing code uses stdlib `flag`. When both are registered, the same flag name can be parsed by either, leading to silently ignored flags or panics on duplicate registration.

**Prevention:**
Two safe paths:
1. Add new flags to the existing stdlib `flag` system (no Cobra introduction)
2. Migrate entirely to `pflag`/Cobra in a single refactor (not incrementally)

Do not mix `flag` and `pflag` in the same binary without a bridge layer.

**Detection:**
- Existing `-r` short flag stops working after Cobra added
- `flag redefined: range` panic at startup

**Phase to address:**
Phase 1 (TUI Foundation) — Decision on flag system must precede adding any new flags.

---

### Pitfall 11: State Management Complexity Explosion

**What goes wrong:**
The dashboard state grows to include: current data, loading state, error state, refresh timer, selected range, selected tab, terminal size, color mode, refresh interval — all in one model. `Update()` becomes an unmaintainable switch statement with hundreds of cases. Bugs in one state transition corrupt another.

**Why it happens:**
Bubble Tea's single-model pattern is clean for simple apps but requires discipline for multi-screen dashboards. Without deliberate component boundaries, the model struct and Update() function accrete state without structure.

**Prevention:**
1. Define explicit state machine states (`Loading`, `Ready`, `Error`, `Paused`)
2. Separate viewport/tab models as sub-components with their own `Update()` methods
3. Keep the top-level model as an orchestrator, not a data store
4. Use typed messages, not booleans: `type loadingMsg struct{}` not `m.isLoading = true`

**Detection:**
- Model struct has more than ~8 fields
- `Update()` function longer than ~60 lines
- Bug fixes in one feature break another unrelated feature

**Phase to address:**
Phase 1 (TUI Foundation) — Define the state machine before writing display logic.

---

## Minor Pitfalls

---

### Pitfall 12: Hardcoded Refresh Intervals Hit API Unexpectedly

**What goes wrong:**
A default refresh interval of 30 seconds seems reasonable, but with 10 req/s averaged over 5 minutes (300 seconds), the budget is 3,000 requests per 5-minute window. However, if the user opens multiple terminal windows running the dashboard, or if the dashboard makes multiple API calls per refresh (stats + summary), the combined load approaches the limit.

**Prevention:**
- Default refresh interval: minimum 60 seconds
- Make interval configurable via `--interval` flag
- Count API calls per refresh: if more than 2 endpoints are called, increase the default interval proportionally
- Display the next-refresh countdown in the status bar so users understand the cadence

**Phase to address:**
Phase 2 (Live Refresh) — Set the default during implementation.

---

### Pitfall 13: WakaTime 302 Redirect Treated as Rate Limit

**What goes wrong:**
WakaTime sometimes returns HTTP 302 instead of 429 when rate limiting. The existing error handler catches only 429 explicitly. A 302 causes the HTTP client to follow the redirect, potentially hitting a different endpoint or timing out.

**Prevention:**
Configure the HTTP client to not follow redirects automatically:

```go
client := &http.Client{
    Timeout: timeout,
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse  // don't follow redirects
    },
}
// Then handle 302 the same as 429 with backoff
```

**Phase to address:**
Phase 2 (Live Refresh) — Handle during API client enhancement.

---

### Pitfall 14: WakaTime `cached_at` Staleness in Status Bar Endpoints

**What goes wrong:**
The `/status_bar/today` endpoint is backed only by cached data. During dashboard display, this endpoint returns an empty `{"data":{"chart_data":[]}}` response while the cache updates. If the dashboard shows this data without checking `cached_at`, it displays empty charts during the cache refresh window.

**Prevention:**
Always check `cached_at` timestamp. If older than the expected update interval, show "Updating..." state instead of empty charts.

**Phase to address:**
Phase 2 (Live Refresh) — Handle during data display implementation.

---

### Pitfall 15: Emoji/Unicode Characters Break Layout in Some Terminals

**What goes wrong:**
The existing `render.go` uses `🬋` (BLOCK SEXTANT character, U+1FB0B) as the bar character. This is a Unicode 13.0 character not supported in all terminal fonts. When the character is unsupported, it renders as a replacement character (□) with different width, breaking bar chart alignment.

**Why it happens:**
The code already acknowledges this with the comment `// ❙ 🬋 ▆ ❘ ❚ █ ━ ▭ ╼ ━ 🬋`. When the character isn't available in the user's font, the fallback is invisible but the byte width still takes space.

**Prevention:**
- Detect terminal Unicode support via `LANG` env var or `TERM` capabilities
- Provide an ASCII fallback (`=`, `-`, `|`) when Unicode is not available
- Use `--no-unicode` flag or auto-detect

**Phase to address:**
Phase 3 (Polish) — Low priority; existing users already see this in static mode.

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| TUI Foundation | Model state mutation from goroutines | Use tea.Cmd for all async work; no direct model writes |
| TUI Foundation | Blocking event loop with sync API calls | Wrap all existing fetch functions as tea.Cmd immediately |
| TUI Foundation | Panic leaving terminal in raw mode | Keep CatchPanics enabled; use WithAltScreen |
| TUI Foundation | Terminal width using stty subprocess | Replace getTerminalCols() with WindowSizeMsg handler |
| TUI Foundation | Flag system conflict (flag vs pflag) | Decide: stay on stdlib flag or migrate to Cobra fully |
| Live Refresh | Goroutine leak from ticker | Use tea.Tick pattern; cancel on quit; one ticker at a time |
| Live Refresh | 429 rate limit from aggressive polling | 60s minimum interval; exponential backoff; show retry status |
| Live Refresh | 202 treated as error | Handle StatusAccepted as "computing" sentinel, retry with delay |
| Live Refresh | 302 redirect from WakaTime rate limiting | Disable redirect following; treat 302 same as 429 |
| Live Refresh | cached_at staleness showing empty charts | Check cached_at; show "Updating..." not empty state |
| Polish | True color codes in degraded terminals | Use Lip Gloss adaptive colors or check COLORTERM |
| Polish | Unicode bar chars in limited fonts | Provide ASCII fallback; auto-detect or --no-unicode flag |

---

## Integration Pitfalls: Adding TUI to This Specific Codebase

The existing wakafetch codebase has specific patterns that create integration risk:

| Existing Pattern | Risk When Adding TUI | Correct Migration |
|-----------------|---------------------|-------------------|
| `ui.Errorln(err.Error())` calls `os.Exit(1)` | Will exit inside running TUI, no cleanup | Return errors; display via TUI error state |
| `getTerminalCols()` spawns `stty` subprocess | Slow, breaks on resize, Windows returns 9999 | Replace with Bubble Tea `WindowSizeMsg` |
| Global `var Clr Colors` package state | Safe in static mode; unsafe if TUI modifies it | Keep static mode path; TUI uses Lip Gloss directly |
| `fetchApi[T]()` uses sync HTTP with 10s timeout | Will block event loop if called directly | Wrap in `tea.Cmd` — function body stays the same |
| `flag` package for CLI flags | Conflicts with Cobra/pflag if introduced | Add `--watch` and `--interval` to existing `flag` setup |
| `main()` directly calls display functions | No separation between data and render layers | Introduce model layer before TUI integration |

---

## "Looks Done But Isn't" Checklist

- [ ] **Race condition check**: Run `go run -race .` in dashboard mode with fast refresh — no races reported
- [ ] **Resize handling**: Resize terminal while running — layout adapts correctly
- [ ] **Panic recovery**: Force a panic in a Cmd goroutine — terminal restores to normal state
- [ ] **Rate limit simulation**: Force 429 response — backoff activates, no retry storm
- [ ] **202 handling**: Use WakaTime's `all_time` range on first load — shows "calculating" not error
- [ ] **Goroutine leak**: Run for 30 minutes with auto-refresh — goroutine count stays stable
- [ ] **Non-TTY mode**: `wakadash --watch | cat` — falls back to static output, no escape codes
- [ ] **Quit during fetch**: Press `q` while data is loading — exits cleanly, terminal restored
- [ ] **ticker cleanup**: Change range while refreshing — old ticker cancelled, no duplicate requests

---

## Sources

### Official Documentation
- [WakaTime API Documentation](https://wakatime.com/developers) — Rate limits, 202 handling, cached_at, pagination
- [Bubble Tea pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbletea) — CatchPanics, WithAltScreen, WindowSizeMsg, tea.Tick
- [Go Race Detector](https://go.dev/doc/articles/race_detector) — Detection methodology

### Verified Community Sources (MEDIUM confidence)
- [Tips for Building Bubble Tea Programs](https://leg100.github.io/en/posts/building-bubbletea-programs/) — Event loop pitfalls, layout arithmetic, panic recovery
- [Bubble Tea on pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbletea) — Signal handling, ErrInterrupted, WithoutCatchPanics
- [How to Implement Retry Logic in Go with Exponential Backoff](https://oneuptime.com/blog/post/2026-01-07-go-retry-exponential-backoff/view) — Backoff patterns for 429

### Project Code Analysis
- `/workspace/wakafetch/api.go` — Existing sync HTTP client, 10s timeout, 429 handling gap
- `/workspace/wakafetch/ui/render.go` — stty-based terminal size detection, sync rendering
- `/workspace/wakafetch/ui/colors.go` — 24-bit true color usage for MidGray
- `/workspace/wakafetch/main.go` — os.Exit() in help handler, flag package usage

### Background
- [ANSI Escape Code Standards (2025)](https://jvns.ca/blog/2025/03/07/escape-code-standards/) — Terminal compatibility landscape
- [Building Bubbletea Programs (Hacker News)](https://news.ycombinator.com/item?id=41369065) — Community experience with pitfalls
- [Understanding Goroutine Leaks in Go](https://leapcell.io/blog/understanding-and-debugging-goroutine-leaks-in-go-web-servers) — Ticker and goroutine lifecycle

---

*Research confidence: HIGH for concurrency/architecture pitfalls (official Bubble Tea docs + race detector docs + project code analysis). MEDIUM for terminal compatibility (community sources, no single authoritative spec). LOW confidence items are flagged inline.*
