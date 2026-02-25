# Project Research Summary

**Project:** wakadash v2.2 milestone
**Domain:** CLI/TUI version update checking for Homebrew-distributed Go applications
**Researched:** 2026-02-25
**Confidence:** HIGH

## Executive Summary

Version update checking in Go TUI applications follows established patterns from the GitHub CLI (gh): async checks at startup, semantic version comparison against GitHub Releases API, and non-intrusive status bar notifications. The recommended implementation uses minimal dependencies (only `golang.org/x/mod/semver` for version parsing), leverages existing Bubble Tea async command patterns already present in wakadash, and integrates with the current status bar rendering without blocking startup.

The critical architectural insight is that version checking is structurally identical to the existing WakaTime API fetches already implemented in wakadash. The same patterns in `internal/tui/commands.go` (goroutine-based tea.Cmd with panic recovery) and message-passing architecture apply directly. The main complexity lies in UX concerns: caching check results to avoid API rate limits (60 requests/hour unauthenticated), respecting terminal width constraints in multi-line notifications, and gracefully handling network failures without disrupting the dashboard experience.

Key risks center on blocking startup with synchronous checks (adds 100-500ms latency), hitting GitHub API rate limits without ETag caching, incorrect semantic version comparison using string operations instead of proper semver parsing, and panics in the goroutine crashing the TUI. All of these are preventable by following existing code patterns and implementing timeouts, proper error handling, and cache-based throttling from day one.

## Key Findings

### Recommended Stack

The technology stack is intentionally minimal, reusing wakadash's existing HTTP client patterns and adding only one new dependency for semantic version comparison.

**Core technologies:**
- **golang.org/x/mod/semver** (v0.33.0): Semantic version parsing and comparison — zero external dependencies beyond stdlib, maintained by Go team, requires "v" prefix which matches GitHub tags exactly
- **net/http + encoding/json** (stdlib): GitHub Releases API client — already used in `internal/api/client.go`, sufficient for single `/repos/{owner}/{repo}/releases/latest` endpoint
- **goroutines + channels** (stdlib): Non-blocking async execution — integrates with Bubble Tea's tea.Cmd pattern, follows existing `fetchStatsCmd()` pattern in codebase

**Implementation approach:** Build a lightweight `internal/version/checker.go` component that wraps GitHub API calls with proper timeouts (3-5 seconds), uses semver for comparison, and returns results via Bubble Tea messages. No need for full GitHub API client libraries like `google/go-github` for a single endpoint.

**Anti-patterns to avoid:** Don't use `github.com/Masterminds/semver` (overkill for simple A > B comparison), don't add retry logic (single attempt is sufficient for informational feature), don't reuse WakaTime API client (separate client with shorter timeout prevents connection pool contention).

### Expected Features

**Must have (table stakes):**
- **Background/async check** — users expect instant startup, not blocked by network calls (gh CLI standard)
- **GitHub Releases API integration** — standard version source via `/repos/{owner}/{repo}/releases/latest`
- **Semantic version comparison** — properly detect v1.10.0 > v1.9.0 (not string comparison)
- **Check frequency throttling** — 24-hour cache interval (industry standard from update-notifier, Homebrew)
- **Non-blocking notification** — status bar message, never modal popup (TUI philosophy)
- **Graceful network failure** — timeout after 2-5s, silent failure, log error but continue
- **Homebrew upgrade command** — actionable instruction: `brew upgrade wakadash`
- **Cache last check time** — store in `~/.config/wakadash/update_cache.json` to enable throttling

**Should have (competitive):**
- Release notes preview — show what's new in latest version (fetch `body` from API)
- Changelog URL link — include `https://github.com/b00y0h/wakadash/releases/tag/vX.Y.Z`
- Major vs minor callout — highlight breaking changes (major bumps) differently
- "Never check" opt-out — environment variable like `WAKADASH_NO_UPDATE_CHECK=1` (gh CLI pattern)

**Defer (v2+):**
- Skip version reminder — per-version dismissal state (requires additional persistence)
- Configurable check frequency — wait for user feedback on 24h default
- Pre-release detection — beta tester opt-in (parse semver.Prerelease())
- First-run grace period — wait one interval before first notification (update-notifier pattern)

**Anti-features to avoid:**
- Auto-update/auto-install — security risk, breaks Homebrew management
- Modal/blocking notification — interrupts workflow, anti-TUI philosophy
- Check on every startup without cache — API rate limiting, network overhead
- Showing errors to user — network failures are common, user can't fix them
- Retry on failure — wastes bandwidth, single attempt sufficient

### Architecture Approach

Version checking integrates into wakadash's existing Bubble Tea architecture using the established async command pattern. The implementation follows the same structure as WakaTime API fetches: launch a goroutine via `tea.Cmd` in `Init()`, wrap in panic recovery, return results via custom message type (`versionCheckCompleteMsg`), handle in `Update()` to update model state (`updateAvailable` field), and conditionally render in `View()` when notification is present.

**Major components:**
1. **internal/version/checker.go** — HTTP client for GitHub Releases API with 3-5 second timeout, semantic version comparison using golang.org/x/mod/semver, ETag-based caching for rate limit mitigation
2. **internal/tui/commands.go** — `checkVersionCmd()` wrapper following existing panic recovery pattern from `fetchStatsCmd()` (lines 60-71)
3. **internal/tui/messages.go** — `versionCheckCompleteMsg` and `versionCheckFailedMsg` types for async results
4. **internal/tui/model.go** — `updateAvailable *UpdateInfo` field, message handlers in `Update()`, multi-line status bar rendering in `View()`

**Data flow timing:** On startup, `Init()` launches version check in parallel with stats fetching. Check runs in background goroutine (200-800ms typical for GitHub API), doesn't block initial render. When result arrives via message, `Update()` sets `m.updateAvailable` field. Next render cycle, `View()` checks field and appends notification to status bar if present. Network failures return error message, handler logs silently but doesn't show to user.

**Patterns to follow:**
- Async Command with Panic Recovery — already used in `fetchStatsCmd()`, `fetchDurationsCmd()`, `fetchSummaryCmd()`
- Silent Failure for Non-Critical Operations — log errors but don't display (version check is informational)
- Conditional Rendering Based on Model State — like help overlay (`if m.showHelp`) and theme picker (`if m.showPicker`)
- Batch Multiple Init Commands — use `tea.Batch()` to launch all async operations in parallel

### Critical Pitfalls

1. **Blocking startup with synchronous check** — Adding 100-500ms+ latency to every launch, complete hang if GitHub API unreachable. Prevention: Always use `tea.Cmd` async pattern, never call version check directly in `Init()`. Follow existing `fetchStatsCmd()` pattern exactly. Launch in `tea.Batch()` with other init commands.

2. **No timeout on HTTP requests** — Default `http.Client` has infinite timeout, causing app to hang when GitHub API unreachable. Prevention: Create dedicated client with `Timeout: 3*time.Second` and use `context.WithTimeout` for request-level control. Dual timeouts (context + client) per 2026 best practices.

3. **GitHub API rate limiting** — Unauthenticated requests limited to 60/hour per IP, easily exhausted in corporate networks or by users running multiple tools. Prevention: Implement ETag-based conditional requests (`If-None-Match` header) that don't count against quota when returning 304. Store ETag in 24-hour cache file alongside version data.

4. **Incorrect semantic version comparison** — String comparison makes "v1.10.0" < "v1.9.0" (lexicographic), pre-releases sort incorrectly. Prevention: Use `golang.org/x/mod/semver.Compare()` for proper SemVer parsing with pre-release ordering (v1.2.3-beta < v1.2.3). Validate both versions with `semver.IsValid()` before comparison.

5. **Panic in goroutine crashes TUI** — Unhandled panics leave terminal in broken state (raw mode, no echo). Prevention: Wrap all `tea.Cmd` functions in defer/recover pattern, already established in `commands.go` lines 60-71. Return error message type on panic for graceful handling.

6. **Showing update notice on every startup** — Becomes nagging spam users ignore or disable. Prevention: Cache check result for 24 hours in `~/.config/wakadash/update_cache.json` with timestamp, only re-check when cache expires. Match gh CLI and Homebrew patterns.

7. **Multi-line notice breaks layout** — Update text wraps awkwardly or overflows on narrow terminals (80 columns). Prevention: Calculate available width, truncate lines if needed using `lipgloss.Width()`, measure total height with `lipgloss.Height()` before rendering.

8. **Assuming Homebrew installation** — User installed via `go install` or binary download, sees incorrect upgrade command. Prevention: Detect installation path (check for `/Cellar/` or `/opt/homebrew/` in executable path), show appropriate command per method, fallback to generic instructions.

## Implications for Roadmap

Based on research, suggested phase structure with clear dependency chain:

### Phase 1: Core Version Checker (Foundation)
**Rationale:** Build and validate GitHub API integration in isolation before TUI complexity. Test HTTP client, JSON parsing, version comparison independently. Establishes reliability foundation for async integration.

**Delivers:**
- `internal/version/checker.go` with GitHub Releases API client
- Semantic version comparison using `golang.org/x/mod/semver`
- Unit tests for API response parsing, timeout handling, version comparison edge cases (v1.9.0 vs v1.10.0, pre-releases, missing "v" prefix)
- ETag-based caching to minimize rate limit exposure

**Addresses:**
- GitHub Releases API integration (table stakes)
- Semantic version comparison (table stakes)

**Avoids:**
- Pitfall 2 (no timeout) — implement 3s timeout from start
- Pitfall 3 (rate limiting) — ETag caching built in
- Pitfall 4 (incorrect semver comparison) — use proper library with test coverage
- Pitfall 10 (brittle JSON parsing) — defensive parsing with fallback, missing field handling

**Research needs:** Standard patterns well-documented, skip `/gsd:research-phase`

### Phase 2: Bubble Tea Message Integration (Async Flow)
**Rationale:** Wire version checker into Bubble Tea lifecycle before UI rendering. Establishes data flow: Init() → tea.Cmd → goroutine → message → Update() → model state. Dependencies verified before visual complexity.

**Delivers:**
- `versionCheckCompleteMsg` and `versionCheckFailedMsg` message types
- `checkVersionCmd()` wrapper with panic recovery (copy pattern from `fetchStatsCmd`)
- `updateAvailable *UpdateInfo` field in model
- Message handlers in `Update()` (silent failure on error, populate field on success)
- Launch from `Init()` via `tea.Batch()` in parallel with stats fetches

**Addresses:**
- Background/async check (table stakes)
- Graceful network failure (table stakes)

**Avoids:**
- Pitfall 1 (blocking startup) — follows tea.Cmd async pattern, returns immediately
- Pitfall 6 (panic crashes TUI) — panic recovery wrapper matching existing code
- Pitfall 9 (connection pool contention) — separate HTTP client with shorter timeout

**Research needs:** Established Bubble Tea patterns already in codebase, skip research

### Phase 3: Status Bar UI Rendering (Visual Display)
**Rationale:** UI layer depends on reliable data flow from Phase 2. Multi-line rendering requires careful layout calculation to avoid breaking terminal display at various widths (test 80 columns minimum).

**Delivers:**
- Refactored `renderStatusBar()` supporting multi-line output
- `buildUpdateNotification()` helper with version diff and upgrade command
- Theme-consistent styling (accent color for info, not red for alarm)
- Terminal width constraint handling using `lipgloss.Width()` measurement
- Format: "Update available: v1.2.3 → v1.2.4 | Run: brew upgrade wakadash"

**Addresses:**
- Non-blocking notification (table stakes)
- Current vs latest version display (table stakes)
- Homebrew upgrade command (table stakes)

**Avoids:**
- Pitfall 11 (layout breakage) — width constraints, height calculation with lipgloss
- Pitfall 7 (raw error display) — silent failure handling from Phase 2

**Research needs:** Standard lipgloss patterns established in codebase, skip research

### Phase 4: Cache & Throttling (Production Reliability)
**Rationale:** Prevents API rate limit exhaustion and reduces network overhead. Required before production use to avoid hitting GitHub's 60 requests/hour limit in corporate environments where multiple users share IP.

**Delivers:**
- `~/.config/wakadash/update_cache.json` persistence (XDG Base Directory spec)
- 24-hour check interval (gh CLI and Homebrew pattern)
- Cache validation on startup (check timestamp before API call)
- Timestamp tracking for last check, stored version data
- Corrupt cache file handling (delete and recreate on parse error)

**Addresses:**
- Check frequency throttling (table stakes)
- Cache last check time (table stakes)

**Avoids:**
- Pitfall 3 (rate limiting) — reduces API call frequency by 24x minimum
- Pitfall 5 (notice every startup) — respects 24h interval
- Pitfall 12 (no dismissal) — foundation for future per-version dismissal state

**Research needs:** Standard cache file I/O patterns, skip research

### Phase 5: Integration Testing & Edge Cases (Polish)
**Rationale:** End-to-end validation with real GitHub API and simulated failure scenarios. Ensures production readiness before deployment. Tests scenarios that unit tests miss.

**Delivers:**
- `cmd/wakadash/main.go` integration with version checker creation
- Real GitHub API testing against wakadash repo
- Network failure simulation (disconnect network, test graceful degradation)
- Rate limit handling validation (simulate 403 response)
- Terminal width testing (80 columns, 24 rows minimum)
- Narrow terminal layout verification

**Addresses:**
- All table stakes features complete and validated end-to-end

**Avoids:**
- Pitfall 8 (installation method detection) — validate Homebrew path detection on Intel/ARM Macs
- All 12 documented pitfalls validated in realistic scenarios

**Research needs:** Skip research, pure testing phase

### Phase Ordering Rationale

- **Foundation → Integration → UI → Production** follows dependency chain: can't integrate what isn't built, can't render what isn't wired, can't deploy without reliability features
- **Phases 1-2 are isolated** — can be developed/tested without visual TUI running, faster iteration on core logic without terminal UI overhead
- **Phase 3 has clear interface** — expects `m.updateAvailable` populated by Phase 2, pure rendering concern with no business logic
- **Phase 4 is orthogonal** — cache logic independent of rendering, can be added/tested separately, optional for initial testing
- **Phase 5 validates assumptions** — real-world testing catches edge cases missed in unit tests (slow networks, narrow terminals, shared IPs)

**Critical path:** Phases 1-2-3 must complete sequentially (dependency chain). Phase 4 can technically happen anytime but should complete before v2.2 release (production requirement). Phase 5 is final validation gate before merge to main.

### Research Flags

**Phases with standard patterns (skip research-phase):**
- **Phase 1:** GitHub Releases API extensively documented, semver libraries mature and stable
- **Phase 2:** Bubble Tea async patterns already demonstrated 3x in existing `commands.go`
- **Phase 3:** Lipgloss rendering patterns established in current status bar implementation
- **Phase 4:** Cache file I/O is standard Go stdlib operations (`os`, `json`, `time` packages)
- **Phase 5:** Testing phase, no new research needed (validation only)

**Future phases needing research (out of scope for v2.2):**
- Installation method detection beyond Homebrew (if multi-platform support added) — would need research on `go install` paths, binary installation detection
- Self-update mechanisms (security implications require deep research) — code signing, verification, rollback
- Release notes parsing and display (GitHub API markdown rendering complexity) — markdown-to-terminal rendering

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | golang.org/x/mod/semver official Go team package with 1,821+ dependents, GitHub API well-documented with 100+ Go library examples, stdlib HTTP patterns proven in wakadash |
| Features | HIGH | Table stakes validated across gh CLI implementation, Homebrew patterns, update-notifier library (15k+ GitHub stars), multiple production CLI tool examples, UX research on notification frequency |
| Architecture | HIGH | Bubble Tea async pattern already implemented 3x in wakadash codebase (fetchStatsCmd, fetchDurationsCmd, fetchSummaryCmd at lines 58-151), exact pattern to replicate with version checking |
| Pitfalls | HIGH | Pitfalls sourced from production CLI tools (gh, npm, docker), HTTP timeout best practices from official Go blogs (Feb 2026), GitHub API rate limit documentation, Bubble Tea panic recovery patterns in existing code |

**Overall confidence:** HIGH

Research based on official documentation (GitHub API docs, golang.org packages, Bubble Tea framework), production tool implementations (gh CLI source code at cli/cli repo), established libraries with thousands of users (update-notifier, go-github), and patterns already proven in wakadash codebase. The problem domain (CLI version checking) is well-understood with 10+ years of precedents across npm, Homebrew, rustup, gh, and similar tools.

### Gaps to Address

- **Installation method detection:** Research assumes Homebrew-only distribution per PROJECT.md constraints. If support expands to `go install` or binary downloads, will need detection logic to show appropriate upgrade commands. Detection complexity is low (check executable path for `/Cellar/` or `/opt/homebrew/` strings) but untested in multi-platform scenarios. *Mitigation: Defer to v2.3+, start with Homebrew-only as planned. Document assumption in code comments.*

- **Cache corruption handling:** Research identified need for defensive cache file parsing but didn't specify detailed recovery strategy beyond "delete and recreate". Edge case: what if cache directory doesn't exist or isn't writable? *Mitigation: During Phase 4 implementation, handle JSON parse errors by deleting corrupt cache and performing fresh check. If directory creation fails, skip cache entirely (treat as first run). Log all cache failures for debugging but never show to user.*

- **Pre-release version handling:** Unclear if wakadash will use pre-release tags (v2.2.0-beta.1) in GitHub releases and whether to notify users about them. golang.org/x/mod/semver handles pre-releases correctly (beta < stable) but should betas trigger "update available" notices for stable users? *Mitigation: Decide during Phase 1 testing — if pre-releases should trigger notifications, no code change needed. If not, add `semver.Prerelease(latest) == ""` check to filter out betas. Document decision in version checker godoc.*

- **Multi-architecture Homebrew bottles:** Homebrew path detection (`/opt/homebrew/` vs `/usr/local/`) varies by architecture (ARM vs Intel Macs). Path patterns well-known but testing needed on both architectures. *Mitigation: Check for both paths in installation detection (Phase 5). Test on Intel and ARM Macs during integration testing. Consider adding detection for Linux Homebrew at `/home/linuxbrew/.linuxbrew/` if wakadash supports Linux distribution via Homebrew.*

- **Build-time version injection:** Wakadash needs current version at runtime for comparison. Typical approach is `go build -ldflags "-X main.Version=v1.2.3"` in Homebrew formula. Need to verify homebrew-wakadash formula supports this. *Mitigation: Check existing formula during Phase 1. If missing, add ldflags to formula before implementing version checking. Version will be "dev" or "unknown" without proper injection.*

## Sources

### Primary (HIGH confidence)
- [golang.org/x/mod/semver documentation](https://pkg.go.dev/golang.org/x/mod/semver) — Official Go semver package, API reference, version format requirements
- [GitHub Releases API documentation](https://docs.github.com/en/rest/releases/releases) — REST API reference, `/repos/{owner}/{repo}/releases/latest` endpoint
- [GitHub API rate limits](https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api) — 60 unauthenticated, 5000 authenticated, conditional requests don't count
- [GitHub API best practices](https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api) — ETag caching, 429 handling
- [Bubble Tea framework](https://github.com/charmbracelet/bubbletea) — TUI framework architecture and async patterns
- [gh CLI environment variables](https://cli.github.com/manual/gh_help_environment) — GH_NO_UPDATE_NOTIFIER, 24h check interval
- [Go HTTP client timeout best practices (Feb 2026)](https://oneuptime.com/blog/post/2026-02-01-go-http-client-timeouts/view) — Current recommendations, context vs client timeouts
- [Go context timeout patterns (Feb 2026)](https://oneuptime.com/blog/post/2026-02-06-fix-context-canceled-errors-otel-go-background-work/view) — Background work patterns, context.Background() usage
- [How to Build Command Line Tools with Bubbletea in Go (Jan 2026)](https://oneuptime.com/blog/post/2026-01-30-how-to-build-command-line-tools-with-bubbletea-in-go/view) — Async patterns and event loop

### Secondary (MEDIUM confidence)
- [sindresorhus/update-notifier](https://github.com/sindresorhus/update-notifier) — Node.js library demonstrating 24h default interval pattern, 15k+ stars
- [Christian1984/go-update-checker](https://github.com/Christian1984/go-update-checker) — Go caching implementation reference
- [rhysd/go-github-selfupdate](https://github.com/rhysd/go-github-selfupdate) — Self-update mechanism patterns
- [Homebrew autoupdate](https://github.com/DomT4/homebrew-autoupdate) — 24h interval validation
- [CLI UX best practices (Evil Martians)](https://evilmartians.com/chronicles/cli-ux-best-practices-3-patterns-for-improving-progress-displays) — Non-intrusive notification guidance
- [Notification design research (Webflow)](https://webflow.com/blog/notification-ux) — 10% disable rate per week finding from UX research
- [Building Terminal UI with Bubble Tea](https://sngeth.com/go/terminal/ui/bubble-tea/2025/08/17/building-terminal-ui-with-bubble-tea/) — Panic recovery patterns, message channels
- [The Bubbletea State Machine pattern](https://zackproser.com/blog/bubbletea-state-machine) — State machine for robust async operations
- [Cloudflare's Complete Guide to Go net/http Timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) — Transport-level granular timeouts

### Tertiary (LOW confidence)
- [tcnksm/go-latest](https://github.com/tcnksm/go-latest) — Unmaintained since 2015 but demonstrates patterns
- Push notification frequency research — 10% uninstall rate extrapolated to CLI context (source: notification design articles)
- [Want to Build a TUI or CLI App? Read This Before You Start](https://yorukot.me/en/blog/before-you-build-a-tui-or-cli-app/) — Version numbering pitfalls (semver)

---
*Research completed: 2026-02-25*
*Ready for roadmap: yes*
