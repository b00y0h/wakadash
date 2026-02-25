# Domain Pitfalls: Version Update Checking

**Domain:** CLI/TUI version update notification
**Researched:** 2026-02-25
**Context:** Adding gh-style version update checking to existing Go TUI (wakadash)

## Critical Pitfalls

These mistakes cause user complaints, degraded UX, or system instability.

### Pitfall 1: Blocking Startup with Synchronous Version Check
**What goes wrong:** Version check HTTP request runs synchronously, adding 100-500ms+ latency to every app launch.

**Why it happens:**
- Default Go `http.Client` has no timeout (waits forever)
- Developers forget network calls are blocking unless wrapped in goroutine
- "Just one quick check" mindset — doesn't feel expensive until users complain

**Consequences:**
- Dashboard startup delayed by network latency (100-500ms best case)
- Complete hang if GitHub API is slow or unreachable
- Degraded UX for users with poor connectivity
- Users perceive the app as "slow" even though core functionality is fast

**Prevention:**
```go
// ❌ BAD: Blocks startup
func Init() tea.Cmd {
    newVersion := checkVersion() // Blocks 100-500ms minimum
    return fetchStatsCmd(client)
}

// ✅ GOOD: Async with timeout
func Init() tea.Cmd {
    return tea.Batch(
        fetchStatsCmd(client),
        checkVersionCmd(), // Returns immediately, runs in goroutine
    )
}
```

**Detection:**
- Startup feels sluggish (>200ms delay before first render)
- Network profiling shows HTTP request before app initialization completes
- Users report freezing when offline

**Remediation:** Always wrap version checks in `tea.Cmd` that executes in background goroutine, just like WakaTime API fetches.

---

### Pitfall 2: No Timeout on Version Check Request
**What goes wrong:** Version check hangs indefinitely when GitHub API is unreachable, eventually causing TUI to appear frozen.

**Why it happens:**
- Go's default `http.Client` has **no timeout** (`Timeout: 0` means infinite)
- Developers copy example code without timeout configuration
- Version check feels "optional" so timeout seems unnecessary

**Consequences:**
- App appears frozen for 2+ minutes when GitHub API is down
- Users force-quit the app, thinking it crashed
- No graceful degradation — check either succeeds or hangs forever

**Prevention:**
```go
// ❌ BAD: No timeout
client := &http.Client{} // Timeout: 0 (infinite)

// ✅ GOOD: Short timeout for version check
client := &http.Client{
    Timeout: 3 * time.Second, // Fast failure for non-critical feature
}

// ✅ BEST: Context deadline for fine-grained control
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
```

**Detection:**
- App hangs when internet connection drops mid-check
- Network profiling shows requests with no timeout
- Users report "app won't start" when GitHub is experiencing outages

**GitHub API Best Practice:** Use 3-5 second timeout for version checks. gh CLI checks once per 24 hours, so fast failure is acceptable.

**Related:** gh CLI environment variable `GH_NO_UPDATE_NOTIFIER` disables checks — consider similar opt-out mechanism.

---

### Pitfall 3: GitHub API Rate Limiting Without Authentication
**What goes wrong:** Unauthenticated GitHub API requests limited to 60/hour/IP. Version check consumes quota, potentially blocking CI/CD or other tools.

**Why it happens:**
- Developers don't realize unauthenticated API has strict rate limits
- "It's just reading public release data" — seems harmless
- Testing with personal usage doesn't reveal shared IP rate limit issues

**Consequences:**
- Users behind corporate NAT hit rate limit after 60 checks across all tools
- CI/CD pipelines fail when multiple jobs check versions
- Error messages are cryptic: `403 rate limit exceeded`
- Users blame the app, not GitHub API limits

**Prevention:**
```go
// ✅ GOOD: Use conditional requests with ETag caching
// First request
resp, _ := http.Get(url)
etag := resp.Header.Get("ETag")
saveToCache(etag)

// Subsequent requests (doesn't count against rate limit if 304 returned)
req.Header.Set("If-None-Match", cachedETag)
resp, _ := client.Do(req)
if resp.StatusCode == 304 {
    // No new version, doesn't consume rate limit quota
}
```

**GitHub API Rate Limit Facts:**
- **Unauthenticated:** 60 requests/hour per IP
- **Authenticated:** 5,000 requests/hour per token
- **Conditional requests (304):** Don't count against limit
- **Best Practice:** Use `ETag` + `If-None-Match` for caching

**Detection:**
- Version check returns `403 Forbidden` errors
- Response headers show `X-RateLimit-Remaining: 0`
- Users in corporate networks report consistent failures

**Homebrew Context:** wakadash distributed via Homebrew — users likely have `gh` CLI installed, which also checks versions. Shared IP quota exhaustion is real risk.

---

### Pitfall 4: Semantic Version Comparison Done Wrong
**What goes wrong:** String comparison (`"v1.10.0" < "v1.9.0"` → true because `"1" < "9"`) or missing handling for pre-releases, build metadata, version prefixes.

**Why it happens:**
- Developers use lexicographic string comparison instead of semver library
- Forgetting version strings have leading `v` (Go convention: `v1.2.3`)
- Pre-release versions (`v1.2.3-beta.1`) sort incorrectly with string comparison

**Consequences:**
- False update notifications: `v1.9.0` shown as "newer" than `v1.10.0`
- Pre-release versions incorrectly advertised as stable updates
- Users lose trust in update notifications after false positives

**Prevention:**
```go
// ❌ BAD: String comparison
if latestVersion > currentVersion { // "v1.9.0" > "v1.10.0" → true!
    showUpdateNotice()
}

// ✅ GOOD: Use golang.org/x/mod/semver (requires "v" prefix)
import "golang.org/x/mod/semver"

if semver.Compare(latestVersion, currentVersion) > 0 {
    showUpdateNotice()
}

// ✅ ALTERNATIVE: github.com/Masterminds/semver (more lenient parsing)
import "github.com/Masterminds/semver/v3"

latest, _ := semver.NewVersion(latestVersion)  // Handles "v1.2.3" or "1.2.3"
current, _ := semver.NewVersion(currentVersion)
if latest.GreaterThan(current) {
    showUpdateNotice()
}
```

**Go Semver Quirks:**
- `golang.org/x/mod/semver` **requires** leading `v` (matches Go module convention)
- `github.com/Masterminds/semver` accepts with/without `v` prefix
- Both handle pre-releases correctly: `v1.2.3-beta.1 < v1.2.3`

**Detection:**
- Update notices for versions you already have installed
- Pre-release versions incorrectly advertised to stable users
- User reports "says update available but I'm already on latest"

---

### Pitfall 5: Showing Update Notice on Every Startup
**What goes wrong:** Update notification displayed every time the app starts, becoming nagging spam users ignore or disable.

**Why it happens:**
- No cache/persistence for "already checked" state
- Developers focus on detection logic, forget UX implications
- "Users should know about updates!" — excessive notification frequency backfires

**Consequences:**
- Users disable update checks completely (lose security update visibility)
- Update notices become "banner blindness" — users ignore them
- Negative reviews mentioning "annoying update spam"

**Prevention:**
```go
// ✅ GOOD: Cache check result for 24 hours (gh CLI pattern)
type VersionCache struct {
    LastChecked time.Time
    LatestVersion string
}

func checkVersion() (string, error) {
    cache := loadCache()

    // Only check if cache expired (24 hours)
    if time.Since(cache.LastChecked) < 24*time.Hour {
        return cache.LatestVersion, nil
    }

    // Perform actual check
    latest := fetchLatestVersion()
    saveCache(VersionCache{
        LastChecked: time.Now(),
        LatestVersion: latest,
    })
    return latest, nil
}
```

**gh CLI Pattern:**
- Checks **once per 24 hours** when any command runs
- Cache stored in `$TMPDIR/gh-cli-cache` or similar
- `GH_NO_UPDATE_NOTIFIER=1` disables checks entirely

**Detection:**
- Users complain about repetitive notices
- No timestamp/cache file in app directory
- Every app launch triggers network request

---

### Pitfall 6: Panic in Version Check Goroutine Crashes TUI
**What goes wrong:** Unhandled panic inside version check goroutine crashes entire TUI, terminal left in broken state.

**Why it happens:**
- Bubble Tea commands run in goroutines without implicit panic recovery
- JSON parsing, network errors, nil pointer dereferences not defensively handled
- Testing focuses on happy path, doesn't simulate malformed GitHub responses

**Consequences:**
- Terminal left in raw mode (no echo, broken cursor)
- User must run `reset` command to fix terminal
- Crash reports difficult to debug (no stack trace visible in TUI)

**Prevention:**
```go
// ✅ GOOD: Explicit panic recovery in tea.Cmd
func checkVersionCmd() tea.Cmd {
    return func() (msg tea.Msg) {
        defer func() {
            if r := recover(); r != nil {
                var err error
                switch v := r.(type) {
                case error:
                    err = v
                default:
                    err = fmt.Errorf("panic in checkVersionCmd: %v", r)
                }
                msg = versionCheckErrMsg{err: err}
            }
        }()

        version, err := fetchLatestVersion()
        if err != nil {
            return versionCheckErrMsg{err: err}
        }
        return versionCheckedMsg{version: version}
    }
}
```

**Existing Code Pattern:** wakadash already uses this pattern in `fetchStatsCmd`, `fetchDurationsCmd`, etc. (see `commands.go` lines 60-71). Apply same pattern to version check.

**Detection:**
- TUI crashes with no error message
- Terminal requires `reset` after crash
- Happens inconsistently (only when specific error conditions trigger panic)

---

## Moderate Pitfalls

These cause degraded functionality but not complete failures.

### Pitfall 7: Displaying Raw GitHub API Errors to Users
**What goes wrong:** User sees cryptic error: `"API rate limit exceeded for 203.0.113.42"` instead of actionable guidance.

**Why it happens:**
- Developers pass through error messages from GitHub API directly
- Focus on functionality, not error message UX
- Error handling as afterthought

**Consequences:**
- Users don't understand what went wrong or how to fix it
- Support burden increases (users asking "what does this mean?")
- Negative perception of app quality

**Prevention:**
```go
// ❌ BAD: Raw API error
if err != nil {
    return fmt.Errorf("version check failed: %w", err)
}

// ✅ GOOD: User-friendly, actionable messages
if resp.StatusCode == 429 {
    return fmt.Errorf("GitHub rate limit reached - version check skipped (will retry later)")
}
if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
    return fmt.Errorf("version check timed out - skipping (will retry next startup)")
}
```

**Best Practice:** Version check is **non-critical** — failures should be silent or minimally intrusive, not error dialogs.

---

### Pitfall 8: Assuming Homebrew Installation Method
**What goes wrong:** Update notice says `brew upgrade wakadash` but user installed via `go install`, binary download, or other method.

**Why it happens:**
- Developers use Homebrew for their own workflow
- "We only officially support Homebrew" doesn't match reality of distribution
- Detection logic not implemented

**Consequences:**
- Users see incorrect upgrade instructions
- Users try `brew upgrade`, get "not installed via Homebrew" error
- Confusion and frustration

**Prevention:**
```go
// ✅ GOOD: Detect installation method
func detectInstallMethod() string {
    // Check if installed via Homebrew
    if execPath, err := os.Executable(); err == nil {
        if strings.Contains(execPath, "/Cellar/") || strings.Contains(execPath, "/opt/homebrew/") {
            return "homebrew"
        }
    }

    // Check for go install
    if _, err := exec.LookPath("go"); err == nil {
        return "go-install" // Suggest: go install github.com/user/repo@latest
    }

    return "unknown" // Generic: "Visit github.com/user/repo/releases"
}

func upgradeCommand() string {
    switch detectInstallMethod() {
    case "homebrew":
        return "brew upgrade wakadash"
    case "go-install":
        return "go install github.com/b00y0h/wakadash@latest"
    default:
        return "Download from github.com/b00y0h/wakadash/releases"
    }
}
```

**Detection:** User feedback "brew upgrade doesn't work" or "I don't use Homebrew."

---

### Pitfall 9: Version Check Competing with WakaTime API Requests
**What goes wrong:** Version check HTTP request uses same `http.Client` as WakaTime API, competing for connection pool slots, increasing latency.

**Why it happens:**
- Reusing existing `api.Client` for convenience
- Not realizing HTTP client connection pool is shared resource
- "One less thing to configure" — reuse seems simpler

**Consequences:**
- Increased latency for WakaTime API requests during version check
- Connection pool exhaustion if version check hangs or retries aggressively
- Difficult to debug (intermittent performance degradation)

**Prevention:**
```go
// ✅ GOOD: Separate HTTP client for version checks
var versionCheckClient = &http.Client{
    Timeout: 3 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:       2,  // Low limit (only used for version checks)
        IdleConnTimeout:    10 * time.Second,
        DisableCompression: true, // GitHub API returns small JSON
    },
}

// Existing WakaTime API client remains unchanged
var apiClient = &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:       10,
        IdleConnTimeout:    90 * time.Second,
    },
}
```

**wakadash Context:** Current `api.Client` has 10s timeout (see `client.go:17`). Version check should use separate, faster-failing client.

---

### Pitfall 10: Not Handling GitHub API Schema Changes
**What goes wrong:** GitHub API changes `/repos/:owner/:repo/releases/latest` response format, version check breaks with JSON parse error.

**Why it happens:**
- Assuming GitHub API is stable (it mostly is, but not guaranteed)
- No fallback when expected fields are missing
- Brittle JSON unmarshaling with no validation

**Consequences:**
- Version check fails silently after GitHub API update
- Users never see update notifications (undetected failure)
- No telemetry to alert developers of the issue

**Prevention:**
```go
// ✅ GOOD: Defensive JSON parsing with fallback
type GitHubRelease struct {
    TagName string `json:"tag_name"` // Required
    Name    string `json:"name"`     // Optional fallback
}

func parseRelease(data []byte) (string, error) {
    var release GitHubRelease
    if err := json.Unmarshal(data, &release); err != nil {
        return "", fmt.Errorf("failed to parse GitHub release: %w", err)
    }

    // Prefer tag_name, fallback to name
    if release.TagName != "" {
        return release.TagName, nil
    }
    if release.Name != "" {
        return release.Name, nil
    }

    return "", fmt.Errorf("GitHub release missing version information")
}
```

**Detection:**
- Version check suddenly stops working for all users simultaneously
- Logs show JSON unmarshal errors
- GitHub API changelog mentions schema updates

---

## Minor Pitfalls

These cause minor UX degradation or edge case issues.

### Pitfall 11: Multi-line Update Notice Breaks Status Bar Layout
**What goes wrong:** Update notice text wraps awkwardly or overflows status bar, breaking TUI layout.

**Why it happens:**
- Not accounting for terminal width when formatting message
- Multi-line text not properly measured with `lipgloss.Height()`
- Testing only on wide terminals (didn't catch narrow terminal breakage)

**Consequences:**
- Status bar becomes unreadable on narrow terminals
- Text overlaps with other UI elements
- Users resize terminal to "fix" the layout

**Prevention:**
```go
// ✅ GOOD: Respect terminal width constraints
func formatUpdateNotice(version, command string, width int) string {
    // Reserve space for borders and padding
    maxWidth := width - 4

    line1 := fmt.Sprintf("Update available: %s", version)
    line2 := fmt.Sprintf("Run: %s", command)

    // Truncate if exceeds width
    if lipgloss.Width(line1) > maxWidth {
        line1 = truncate(line1, maxWidth)
    }
    if lipgloss.Width(line2) > maxWidth {
        line2 = truncate(line2, maxWidth)
    }

    return lipgloss.JoinVertical(lipgloss.Left, line1, line2)
}
```

**wakadash Context:** Current `renderStatusBar()` uses single-line format (see `model.go:466-488`). Multi-line notice needs height calculation.

---

### Pitfall 12: No Visual Distinction Between New and Seen Notices
**What goes wrong:** Update notice shown every startup until user upgrades, even after they've seen it. No way to dismiss.

**Why it happens:**
- Notification persistence not considered beyond initial "check once per day"
- No UX for "I saw this, stop showing it"
- Assumption: "users will upgrade right away"

**Consequences:**
- Users annoyed by persistent notification they can't dismiss
- "I'll upgrade later" users see notice every day for weeks
- No way to signal "I know, stop reminding me"

**Prevention:**
```go
// ✅ GOOD: Track last seen version
type VersionCache struct {
    LastChecked   time.Time
    LatestVersion string
    LastSeen      string // User has seen notice for this version
}

func shouldShowNotice(current, latest string, cache VersionCache) bool {
    // Don't show if already running latest
    if semver.Compare(latest, current) <= 0 {
        return false
    }

    // Don't show if user already saw this version's notice
    if latest == cache.LastSeen {
        return false
    }

    return true
}

// When user dismisses notice (maybe 'd' key?), record it
cache.LastSeen = latestVersion
saveCache(cache)
```

**UX Enhancement:** Add keybinding to dismiss notice (e.g., `d` = dismiss, stops showing until next version).

---

## Phase-Specific Warnings

Guidance for each implementation phase.

### Phase 1: GitHub Releases API Integration
**Likely Pitfalls:**
- **Pitfall 2:** No timeout (infinite hang)
- **Pitfall 3:** Rate limiting without ETag caching
- **Pitfall 10:** Brittle JSON parsing

**Mitigation:**
- Start with 3-second timeout on dedicated `http.Client`
- Implement ETag caching from day 1 (not "later optimization")
- Add comprehensive error handling tests (malformed JSON, missing fields)

---

### Phase 2: Version Comparison Logic
**Likely Pitfalls:**
- **Pitfall 4:** String comparison instead of semver
- **Pitfall 10:** Assuming version format never changes

**Mitigation:**
- Use `golang.org/x/mod/semver` (already used in Go ecosystem)
- Defensively parse version strings (handle missing `v` prefix gracefully)
- Unit test edge cases: pre-releases, build metadata, leading `v`

---

### Phase 3: TUI Integration (Status Bar Notice)
**Likely Pitfalls:**
- **Pitfall 1:** Blocking startup with synchronous check
- **Pitfall 6:** Panic crashes TUI
- **Pitfall 11:** Multi-line notice breaks layout

**Mitigation:**
- Follow existing `tea.Cmd` pattern (see `commands.go`)
- Add panic recovery wrapper like other commands
- Calculate status bar height with `lipgloss.Height()` before rendering
- Test on narrow terminals (80 columns, 24 rows)

---

### Phase 4: Cache & Frequency Management
**Likely Pitfalls:**
- **Pitfall 5:** Showing notice every startup
- **Pitfall 12:** No way to dismiss notices

**Mitigation:**
- Implement 24-hour cache (match gh CLI pattern)
- Store cache in `$HOME/.config/wakadash/version-cache.json` (XDG compliance)
- Add environment variable `WAKADASH_NO_UPDATE_NOTIFIER` for opt-out

---

### Phase 5: Installation Method Detection
**Likely Pitfalls:**
- **Pitfall 8:** Assuming Homebrew for all users

**Mitigation:**
- Detect installation path (Homebrew vs go install vs manual)
- Show appropriate upgrade command per method
- Fall back to generic "visit releases page" for unknown methods

---

## Testing Checklist

Scenarios to validate before release:

### Network Failure Scenarios
- [ ] GitHub API unreachable (DNS failure)
- [ ] Request timeout (slow network)
- [ ] Rate limit exceeded (403 response)
- [ ] Malformed JSON response
- [ ] Invalid semver in tag_name field

### TUI Integration
- [ ] Check runs async (doesn't block startup)
- [ ] Panic in check doesn't crash TUI
- [ ] Notice renders correctly in narrow terminal (80x24)
- [ ] Multi-line notice doesn't overflow status bar
- [ ] Loading spinner doesn't interfere with check

### Cache Behavior
- [ ] Check runs once per 24 hours, not every startup
- [ ] Cache persists across app restarts
- [ ] Cache invalidates after 24 hours
- [ ] Corrupt cache file doesn't crash app

### Version Comparison
- [ ] Correctly compares `v1.9.0` vs `v1.10.0`
- [ ] Handles pre-releases: `v1.2.3-beta` < `v1.2.3`
- [ ] Handles missing `v` prefix gracefully
- [ ] Detects when current version is latest

### Installation Method Detection
- [ ] Homebrew installation detected correctly
- [ ] `go install` method detected
- [ ] Manual binary download shows generic instructions
- [ ] Upgrade command matches detected method

---

## Sources

**CLI/TUI Best Practices:**
- [Want to Build a TUI or CLI App? Read This Before You Start](https://yorukot.me/en/blog/before-you-build-a-tui-or-cli-app/) — Version numbering pitfalls (semver), auto-update UX considerations
- [Improve the UX of CLI tools with version update warnings](https://medium.com/trabe/improve-the-ux-of-cli-tools-with-version-update-warnings-23eb8fcb474a) — CLI version check UX, synchronous vs async patterns
- [Command Line Interface Guidelines](https://clig.dev/) — General CLI best practices
- [Node.js CLI Apps Best Practices](https://github.com/lirantal/nodejs-cli-apps-best-practices) — Configuration persistence, avoiding repetitive prompts

**GitHub API Rate Limits & Caching:**
- [Rate limits for the REST API - GitHub Docs](https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api) — Unauthenticated: 60/hour, authenticated: 5,000/hour, conditional requests (304) don't count
- [Best Practices for Using the REST API - GitHub Docs](https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api) — Use ETag + If-None-Match for caching, handle 429 with exponential backoff
- [A Developer's Guide: Managing Rate Limits for the GitHub API](https://www.lunar.dev/post/a-developers-guide-managing-rate-limits-for-the-github-api) — Authentication strategies, conditional requests
- [GitHub API Rate Limiting Solutions](https://software-testing-tips.medium.com/top-10-solutions-for-github-api-rate-limiting-fb0caf2b5557) — Token rotation, caching strategies

**Async Startup Patterns:**
- [BackgroundService Gotcha: Startup](https://blog.stephencleary.com/2020/05/backgroundservice-gotcha-startup.html) — Synchronous code blocks startup until first await
- [Improve the UX of CLI tools with version update warnings](https://medium.com/trabe/improve-the-ux-of-cli-tools-with-version-update-warnings-23eb8fcb474a) — CLI tools should avoid blocking on async operations, use DeAsync.js pattern
- [Making BackgroundService startup not block Host startup](https://github.com/dotnet/runtime/issues/36063) — Workaround: wrap in Task.Run or use Task.Yield()

**Caching Strategies:**
- [Prettier 2.7: new --cache CLI option](https://prettier.io/blog/2022/06/14/2.7.0) — Cache keys include tool version, options, file metadata/content, enables dramatic performance improvement
- [npm-cache | npm Docs](https://docs.npmjs.com/cli/v8/commands/npm-cache/) — Cache is self-healing, corruption triggers auto-refetch, rarely needs manual clearing

**Bubble Tea Concurrency Patterns:**
- [Building a Terminal IRC Client with Bubble Tea](https://sngeth.com/go/terminal/ui/bubble-tea/2025/08/17/building-terminal-ui-with-bubble-tea/) — Central message channel for goroutine communication, each command in isolated goroutine with panic recovery
- [HTTP and Async Operations | Bubble Tea Tutorials](https://deepwiki.com/charmbracelet/bubbletea/6.4-step-by-step-tutorials) — Bubbletea command pattern for async HTTP, blocking I/O in separate goroutines doesn't freeze UI
- [Injecting messages from outside the program loop](https://github.com/charmbracelet/bubbletea/issues/25) — Program.Send() for thread-safe message injection
- [The Bubbletea State Machine Pattern](https://zackproser.com/blog/bubbletea-state-machine) — Async work via tea.Cmd, event loop handles messages sequentially (no race conditions)

**Semantic Versioning:**
- [Masterminds/semver - Go Package](https://github.com/Masterminds/semver) — StrictNewVersion vs NewVersion (coercing), pre-release ordering, build metadata handling
- [golang.org/x/mod/semver](https://pkg.go.dev/golang.org/x/mod/semver) — Requires "v" prefix, shorthands (v1 → v1.0.0), follows Semver 2.0.0 with exceptions
- [Semantic Versioning Pitfalls](https://commandbox.ortusbooks.com/package-management/semantic-versioning) — Comparison methods differ from range rules, npm/js vs Cargo vs PHP patterns diverge

**HTTP Timeouts & Context:**
- [Timeouts in Go: A Comprehensive Guide](https://betterstack.com/community/guides/scaling-go/golang-timeouts/) — Default http.Client has no timeout (footgun), always set timeout for production
- [The Complete Guide to Go net/http Timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) — Deadlines vs timeouts, transport-level granular timeouts (DialContext, TLSHandshakeTimeout, ResponseHeaderTimeout)
- [How to Handle HTTP Client Timeouts Properly in Go](https://oneuptime.com/blog/post/2026-02-01-go-http-client-timeouts/view) — Use context for per-request control, client timeout for global default
- [Context Deadline Exceeded in Go](https://gosamples.dev/context-deadline-exceeded/) — Context propagation triggers inner layers to give up, deadlines trigger cancellation

**Graceful Degradation:**
- [Graceful Degradation | TechTarget](https://www.techtarget.com/searchnetworking/definition/graceful-degradation) — System maintains limited functionality when components fail, prevents catastrophic failure
- [How to Build Fault-Tolerant Services with Graceful Degradation in Go](https://oneuptime.com/blog/post/2026-01-25-fault-tolerant-graceful-degradation-go/view) — Every external call needs timeout (use context.WithTimeout), try/catch for network failures, serve fallback instead of error
- [Graceful Degradation in Microservices](https://medium.com/@mani.saksham12/graceful-degradation-in-a-microservice-architecture-using-kubernetes-d47aa80b7d20) — Circuit breaker pattern, retry strategies for transient failures, fallbacks must be independent
- [AWS Well-Architected: Graceful Degradation](https://docs.aws.amazon.com/wellarchitected/latest/reliability-pillar/rel_mitigate_interaction_failure_graceful_degradation.html) — Transform hard dependencies into soft dependencies, minimize impact on callers

**gh CLI Version Check Implementation:**
- [gh help environment - GitHub CLI](https://cli.github.com/manual/gh_help_environment) — `GH_NO_UPDATE_NOTIFIER` disables update checks, checks run once per 24 hours when any command executes, upgrade notice on stderr
- [gh config clear-cache](https://cli.github.com/manual/gh_config_clear-cache) — Cache stored in `$TMPDIR/gh-cli-cache`, occasionally needs clearing for troubleshooting

**Homebrew CLI:**
- [brew(1) – Homebrew Documentation](https://docs.brew.sh/Manpage) — `brew outdated` lists available updates, `brew upgrade <formula>` upgrades specific package
- [Homebrew FAQ](https://docs.brew.sh/FAQ) — Auto-cleanup runs every 30 days, uninstalls old formula versions automatically
