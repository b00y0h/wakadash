# Feature Landscape: CLI Version Update Notifications

**Domain:** Command-line tool version checking
**Researched:** 2026-02-25
**Confidence:** HIGH

## Table Stakes

Features users expect. Missing = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Background/async check** | Startup must not be delayed | Medium | Update check in separate goroutine, don't block main program; gh CLI standard |
| **GitHub Releases API query** | Standard source for version info | Low | Query `/repos/{owner}/{repo}/releases/latest` endpoint |
| **Semantic version comparison** | Determine if update exists | Low | Use golang.org/x/mod/semver.Compare() for proper SemVer (v1.10.0 > v1.9.0) |
| **Cache last check time** | Avoid API spam, rate limiting | Low | Store timestamp in `~/.config/wakadash/update_cache.json` or similar |
| **Check frequency throttling** | Daily or weekly, not every run | Low | **Default 24 hours** (industry standard: update-notifier, Homebrew, Salesforce CLI) |
| **Non-blocking notification** | Show message without interrupting workflow | Medium | Display in status bar, not modal popup; Bubble Tea integration |
| **Current vs latest version display** | Show what user has and what's available | Low | Format: "Update available: v1.2.0 → v1.3.0" |
| **Graceful network failure** | Don't crash/hang if GitHub unreachable | Medium | Timeout 2-5s, silent failure, log error but continue |
| **Installation-specific upgrade command** | Tell user HOW to upgrade | Low | For wakadash: `brew upgrade wakadash` (Homebrew-only v1) |

**Rationale for table stakes:**
- **gh CLI precedent**: GitHub CLI demonstrates this pattern—users familiar with one expect it in others
- **Homebrew convention**: Tools distributed via Homebrew often include update checks with upgrade commands
- **Developer audience**: wakadash users are developers who expect professional CLI UX (non-blocking, informative, actionable)
- **Research finding**: "A push notification a week leads to 10% of users disabling notifications" — must respect user time

## Differentiators

Features that set product apart. Not expected, but valued if present.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Release notes preview** | Show what's new in latest version | Medium | Fetch release `body` from GitHub API, display first few lines |
| **Changelog/release URL link** | Users can review full changes before upgrading | Low | Include `https://github.com/b00y0h/wakadash/releases/tag/vX.Y.Z` |
| **Skip version reminder** | Users can dismiss notice until next version | Medium | Requires local state persistence in `~/.config/wakadash/` |
| **Major vs minor callout** | Highlight breaking changes (major bumps) differently | Low | Use semver.Major() to detect version jump, add warning styling |
| **Configurable check frequency** | User control over notification cadence | Low | Config option for interval (daily/weekly/never); defer to user feedback |
| **"Never check" opt-out** | Respect user preference completely | Low | Config flag or env var (e.g., `WAKADASH_NO_UPDATE_CHECK=1`) |
| **First-run grace period** | Don't show update on first run (even if available) | Low | Wait one interval before first notification (update-notifier pattern) |
| **Offline mode awareness** | Detect offline, skip check entirely | Low | Pre-check network connectivity before API call |
| **Pre-release detection** | Notify beta testers of pre-release versions | Low | Parse semver.Prerelease(), show opt-in notice |

**Recommendation:** Focus on table stakes for v2.2 milestone. Differentiators can be added later if users request them.

## Anti-Features

Features to explicitly NOT build.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **Auto-update/auto-install** | Security risk, breaks Homebrew management, unwanted for CLI tools | Manual upgrade only; show command, let user decide |
| **Modal/blocking notification** | Interrupts workflow, annoying UX, anti-TUI philosophy | Non-intrusive status bar message or end-of-session notice |
| **Check on every startup (no cache)** | API rate limiting, network overhead, slow startup, user annoyance | Cache check results, throttle to 24h default minimum |
| **Notification on first run** | Annoying to new users who just installed latest version | Wait until first interval expires (update-notifier pattern) |
| **Intrusive color/formatting** | Red warnings feel like errors, not helpful info | Subtle info styling (accent colors), not alarm colors |
| **Update nag on every command** | User fatigue, leads to disabled checks, 10% uninstall rate | Show once per session or less frequently |
| **Phone-home analytics/telemetry** | Privacy concerns, unnecessary for update check, trust erosion | Pure version comparison, no tracking |
| **Email/external notifications** | Outside CLI scope, adds complexity, not terminal-native | Terminal-only notifications |
| **Version checking for non-Homebrew installs** | Complex (multiple package managers), unclear upgrade path | Homebrew-only for v1 (PROJECT.md constraint) |
| **Forced checks/retry on failure** | Wastes bandwidth, treats users like children | Allow silent failure, never retry on timeout |

**Rationale:**
- **Homebrew philosophy**: Package manager handles updates, app just notifies
- **TUI principles**: Non-intrusive, user-controlled, no surprises
- **Developer trust**: wakadash users are technical—they'll update when ready
- **Research finding**: "Notifications can cut both ways. If handled well, they boost UX, but when executed poorly, risk becoming an annoyance"

## Feature Dependencies

```
Semantic version comparison
    ↓
GitHub latest release fetch
    ↓
Version diff calculation
    ↓
Status bar display ← Non-blocking async check
                   ← Silent error handling
                   ← Check frequency throttling (cache)
```

**Critical path:** Must have semver comparison working before UI display makes sense. Version check must be async before app is usable (table stakes).

## MVP Recommendation

**Prioritize for v2.2:**

1. ✅ **Background/async check** - Startup must feel instant (launch in goroutine)
2. ✅ **GitHub releases/latest API** - Fetch latest version tag
3. ✅ **Semantic version comparison** - Detect when update available
4. ✅ **Status bar notice** - Display version diff + upgrade command
5. ✅ **Silent failure** - Network errors don't break dashboard
6. ✅ **Homebrew upgrade command** - Actionable next step
7. ✅ **Cache last check time** - Enable 24h throttling
8. ✅ **Check frequency throttling** - Default 24h interval

**Defer to future (if requested):**

- ⏸️ Skip version reminder (state persistence beyond last check time)
- ⏸️ Changelog/release notes preview (nice-to-have)
- ⏸️ Configurable check frequency (wait for user feedback on 24h default)
- ⏸️ Major vs minor styling (polish)
- ⏸️ Pre-release opt-in (advanced)
- ⏸️ First-run grace period (low priority for dev tool)
- ⏸️ "Never check" opt-out (add if users request it)

**Rationale:** MVP delivers complete basic experience matching gh CLI. Deferred features are polish, not core value. Ship fast, iterate based on user feedback.

## Implementation Details

### Check Frequency Research

**Industry standard: 24 hours (1 day)**

From research:
- npm's update-notifier: 1 day default (`updateCheckInterval`)
- Homebrew autoupdate: 24 hours default interval
- Salesforce CLI: Shows update warnings on command runs
- Research finding: "A push notification a week leads to 10% of users disabling notifications, and 6% to uninstalling the apps"

**Recommendation:** Start with 24h, make configurable later only if users request different intervals.

### Non-Intrusive Display Patterns

From update-notifier and UX research:

1. **First-run grace period:** Wait one interval before showing notification (even if update exists)
2. **Background check in unref'd process:** Don't block startup or exit
3. **Cache results between runs:** Load cached state into model
4. **TTY detection:** Only show in interactive terminals
5. **Passive notification:** Informational, not urgent; low attention level
6. **No error displays:** Silent failure on network issues

### Message Format Examples

Based on CLI patterns from research:

```
Simple format (Salesforce CLI style):
› Warning: wakadash update available from 1.2.0 to 1.3.0.

Detailed format (feature request style):
Update available: wakadash v1.2.0 → v1.3.0
Run: brew upgrade wakadash

Multi-line status bar format (target for wakadash):
┌─────────────────────────────────────────────────────┐
│ ⬆ Update available: v1.2.0 → v1.3.0                 │
│ Run: brew upgrade wakadash                          │
│ Release: https://github.com/b00y0h/wakadash/rel...  │
└─────────────────────────────────────────────────────┘
```

### Bubble Tea Integration

For wakadash TUI implementation:

**Components available:**
- Bubble Tea core: message passing for async updates
- bubbles/list: Has `NewStatusMessage` pattern for timed messages
- Custom statusbar component: `StatusMessageLifetime time.Duration` field

**Integration pattern:**
```go
// In Init()
func (m model) Init() tea.Cmd {
    return tea.Batch(
        // ... other init commands
        checkForUpdate(), // Launch async check
    )
}

// Async check function
func checkForUpdate() tea.Cmd {
    return func() tea.Msg {
        // Check version in background goroutine
        // Return updateAvailableMsg or nil
    }
}

// In Update()
case updateAvailableMsg:
    m.updateNotice = msg.formatNotice()
    return m, nil

// In View()
// Display m.updateNotice in status bar area
```

**Key points:**
- Launch check in `Init()` as tea.Cmd
- Don't block on network call
- Store update notice in model
- Display in status bar (non-modal)
- Silent if check fails

### GitHub Releases API

**Endpoint:**
```
GET https://api.github.com/repos/{owner}/{repo}/releases/latest
```

**Rate limits:**
- Unauthenticated: 60 requests/hour
- Authenticated (with token): 5000 requests/hour

**Response fields:**
- `tag_name`: Version string (e.g., "v1.2.3")
- `name`: Release title
- `body`: Release notes (markdown)
- `html_url`: Link to release page
- `prerelease`: Boolean flag

**Implementation notes:**
- Use authenticated requests via `GITHUB_TOKEN` env var for higher limit
- Cache results to minimize API calls
- Set timeout: 2-5 seconds recommended
- Use context with timeout for cancellation
- Handle 404 (no releases), 403 (rate limited), network errors

### Caching Strategy

**File location:** `~/.config/wakadash/update_cache.json` or similar

**Cache fields:**
```json
{
  "last_check": "2026-02-25T10:30:00Z",
  "latest_version": "v1.2.4",
  "current_version": "v1.2.3",
  "checked_at_version": "v1.2.3"
}
```

**Logic:**
1. On startup, read cache file
2. Check `last_check` timestamp
3. If < 24h ago, skip API call (use cached data)
4. If >= 24h ago, make API call, update cache
5. If `current_version != checked_at_version`, clear cache (user upgraded)

**Benefits:**
- Respects API rate limits
- Faster startup (no network call most runs)
- Reduces user bandwidth usage
- Prevents "update nag" on every run

### Timeout Handling

**Recommendations from research:**
- HTTP client timeout: 2-5 seconds
- Context with deadline for cancellation
- Log error to debug output (if available)
- Don't display error to user
- Continue normal operation if check fails

**Example pattern:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
resp, err := client.Do(req)
if err != nil {
    // Log error, return nil (silent failure)
    return nil
}
```

## User Stories

### Primary Story (MVP)

```
AS A wakadash user who installed via Homebrew
WHEN I start the dashboard after a new version is released
THEN I see a notice in the status bar showing the version diff and upgrade command
SO THAT I can update to the latest version when convenient
```

**Acceptance criteria:**
- Notice appears within 5 seconds of startup (or not at all if network slow)
- Dashboard loads immediately, check doesn't block
- Version diff is clear: "Update available: v1.2.3 → v1.2.4"
- Upgrade command is actionable: "Run: brew upgrade wakadash"
- If network fails, dashboard works normally (no error shown)
- Check only happens once per 24 hours (cached)

### Secondary Stories (Future)

**Dismissable notice:**
```
AS A user who knows an update exists but isn't ready to upgrade
WHEN I press a key to dismiss the update notice
THEN it doesn't reappear until the next version is released
SO THAT I'm not nagged repeatedly about the same version
```

**Bandwidth respect:**
```
AS A user on metered/slow connection
WHEN I start wakadash multiple times in one day
THEN version check only happens once per 24 hours
SO THAT I don't waste bandwidth on repeated checks
```

## Competitive Analysis

How other CLI tools handle version updates:

| Tool | Approach | Lessons |
|------|----------|---------|
| **gh (GitHub CLI)** | Async check, status bar notice, "gh upgrade" command | Gold standard - non-intrusive, actionable |
| **npm** | Synchronous check, visible delay, blocks startup | What NOT to do - users complain about slowness |
| **brew** | Controlled updates, app notifies but doesn't execute | Trust users to manage their systems |
| **rustup** | Fast check, shows version + command, channel support | Good UX, clear messaging |
| **docker** | No built-in checking, relies on "docker version" command | Too hands-off for desktop tools |

**Key insight:** gh CLI is the gold standard. Users familiar with gh expect:
1. Instant startup (async check)
2. Clear version diff in status bar
3. Copy-paste upgrade command
4. Never blocks workflow

wakadash should match this UX.

## Edge Cases to Handle

### Network Conditions
- **No internet**: Silently fail, show no notice
- **GitHub API down**: Same as no internet
- **Slow connection**: Timeout after 5s, show no notice
- **Intermittent DNS**: Same as timeout
- **Rate limited (403)**: Silently fail, try again in 24h

**Never:** Show error messages for network issues. This is enhancement, not critical feature.

### Version Scenarios
- **Current = latest**: No notice shown
- **Current > latest** (dev build): No notice shown (not an error)
- **Invalid version format**: Silently fail (log error for debugging)
- **No releases exist**: Silently fail (API returns 404)
- **Pre-release marked "latest"**: Show notice (GitHub API behavior)

### Installation Scenarios
- **Homebrew install**: Show notice with "brew upgrade wakadash" ✅
- **Go install**: No notice (out of scope for v1) ⏸️
- **Binary download**: No notice (out of scope for v1) ⏸️
- **Development build**: No notice if version = "dev" 🤔

**Open question:** Should dev builds show "(development build, skipping update check)" or just silently skip? Probably silently skip—devs know they're on dev build.

## Testing Strategy

### Unit Tests
```go
TestSemverComparison()
    - v1.2.3 < v1.2.4 ✅
    - v1.2.4 < v1.3.0 ✅
    - v1.9.0 < v1.10.0 ✅ (not string comparison)
    - v1.2.3 == v1.2.3 ✅
    - v2.0.0 > v1.9.9 ✅
    - Invalid versions return error ✅

TestGitHubAPIResponse()
    - Valid JSON parses correctly ✅
    - Missing tag_name returns error ✅
    - Network timeout returns error ✅
    - 404 returns error ✅
    - Rate limit (403) returns error ✅

TestCacheLogic()
    - Cache hit within 24h skips API ✅
    - Cache miss after 24h calls API ✅
    - Version change clears cache ✅
    - Invalid cache file recreates ✅
```

### Integration Tests
```go
TestVersionCheckCommand()
    - Returns message when newer version ✅
    - Returns nil when current is latest ✅
    - Returns nil on network error ✅
    - Completes within timeout ✅
    - Respects cache (doesn't call API twice) ✅
```

### Manual Testing
- Start app with slow network → dashboard appears instantly
- Start app with no network → dashboard works, no error shown
- Start app with older version → notice appears in status bar
- Start app with current version → no notice appears
- Version notice shows correct diff and command
- Second start within 24h → no API call (verify with network monitor)
- Start after 24h → API call happens, cache updated

## Display Design

### Status Bar Notice (when update available)

```
┌─────────────────────────────────────────────────────────┐
│ Dashboard content...                                     │
│                                                           │
├─────────────────────────────────────────────────────────┤
│ ⬆ Update available: v1.2.3 → v1.2.4                      │
│ Run: brew upgrade wakadash                               │
│ Release: https://github.com/b00y0h/wakadash/releases/... │
└─────────────────────────────────────────────────────────┘
```

**Styling (lipgloss):**
- Accent color border (green/blue - not red/alarm color)
- Icon: ⬆ or 🔔 or ⚡ (optional, keep minimal)
- Bold version numbers
- Monospace font for command
- Subtle background color to distinguish from main content

**Position:** Bottom of screen (status bar area, doesn't push content up)

**Size:** 2-3 lines max (compact, non-intrusive)

**Dismissal:** Not in v1 (always visible if update available and cache valid)

## Success Metrics

How to measure if feature provides value:

**Qualitative:**
- Users report seeing update notices and upgrading
- No complaints about slow startup
- No GitHub issues about "version check broke my dashboard"
- Positive mentions in reviews/feedback

**Quantitative (if telemetry existed, which we DON'T do):**
- Update notices shown: X per week
- Network failures: Y per week (expect 10-20%, acceptable)
- Check duration: Average <2s, p95 <5s

**Proxy metrics (without telemetry):**
- GitHub release download counts increase after new version
- No performance-related issues filed
- Feature gets mentioned in user reviews/tweets as helpful
- Homebrew analytics (if available) show update adoption

## Confidence Assessment

| Topic | Level | Source |
|-------|-------|--------|
| Check frequency (24h standard) | HIGH | Multiple sources (update-notifier, Homebrew docs, Salesforce CLI examples) |
| Background/async pattern | HIGH | update-notifier architecture, UX best practices, gh CLI behavior |
| GitHub Releases API | HIGH | Standard practice, multiple library implementations |
| Non-intrusive display | HIGH | UX research, CLI tool conventions, Bubble Tea patterns |
| Message format | MEDIUM | Limited CLI examples found, extrapolated from patterns |
| Bubble Tea integration | MEDIUM | Framework docs available, statusbar component exists |
| Caching strategy | HIGH | go-update-checker library, update-notifier patterns |
| Timeout recommendations | MEDIUM | Best practices extrapolated from general guidance |
| User annoyance thresholds | MEDIUM | UX research (10% disable rate), not CLI-specific |

## Sources

### Update Notification Libraries & Patterns
- [sindresorhus/update-notifier](https://github.com/sindresorhus/update-notifier) — Node.js library with non-intrusive patterns (HIGH confidence)
- [vercel/update-check](https://github.com/vercel/update-check) — Minimalistic CLI update notifications
- [Christian1984/go-update-checker](https://github.com/Christian1984/go-update-checker) — Go library for GitHub release checking with caching

### CLI Tools Research
- [GitHub CLI (gh) repository](https://github.com/cli/cli) — Reference implementation
- [GitHub CLI discussions: How to upgrade](https://github.com/cli/cli/discussions/4630)
- [GitHub CLI issue: Disable update notifications](https://github.com/cli/cli/issues/743)
- [Salesforce CLI update notification issue](https://github.com/forcedotcom/cli/issues/1260) — Example message format
- [OpenAI Codex CLI feature request](https://github.com/openai/codex/issues/2806) — User expectations

### Homebrew Integration
- [Automating Homebrew Tap Updates with GitHub Actions](https://builtfast.dev/blog/automating-homebrew-tap-updates-with-github-actions/)
- [Homebrew Documentation: How to Create and Maintain a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [Homebrew Documentation: Taps](https://docs.brew.sh/Taps)
- [homebrew-autoupdate tap](https://github.com/DomT4/homebrew-autoupdate) — 24h default interval
- [Enabling Auto-updates in Homebrew](https://easyosx.net/2024/01/29/enabling-auto-updates-in-homebrew/)
- [tap-release GitHub App](https://github.com/toolmantim/tap-release) — Automate tap updates on releases

### UX Best Practices & Notification Design
- [Notification Design: Practical dos and don'ts (Webflow)](https://webflow.com/blog/notification-ux)
- [Design Guidelines For Better Notifications UX (Smashing Magazine)](https://www.smashingmagazine.com/2025/07/design-guidelines-better-notifications-ux/)
- [A Comprehensive Guide to Notification Design (Toptal)](https://www.toptal.com/designers/ux/notification-design)
- [CLI UX best practices (Evil Martians)](https://evilmartians.com/chronicles/cli-ux-best-practices-3-patterns-for-improving-progress-displays)
- [Adobe: Notification Overload Best Practices](https://theblog.adobe.com/notification-overload-best-practices-for-designing-notifications-with-respect-for-users/)
- [Indicators, Validations, and Notifications (Nielsen Norman Group)](https://www.nngroup.com/articles/indicators-validations-notifications/)

### Bubble Tea Framework & Go TUI Development
- [Bubble Tea TUI framework](https://github.com/charmbracelet/bubbletea) — Core framework
- [Bubbles component library](https://github.com/charmbracelet/bubbles) — Includes list with NewStatusMessage
- [Bubble Tea statusbar component](https://pkg.go.dev/github.com/noahgorstein/jqp/tui/bubbles/statusbar) — Custom statusbar example
- [Building Terminal UI with Bubble Tea (blog post)](https://sngeth.com/go/terminal/ui/bubble-tea/2025/08/17/building-terminal-ui-with-bubble-tea/)
- [TUI Components in Go with Bubble Tea](https://applegamer22.github.io/posts/go/bubbletea/)
- [Injecting messages from outside program loop (issue)](https://github.com/charmbracelet/bubbletea/issues/25)

### Go CLI Development
- [Go CLI solutions page](https://go.dev/solutions/clis) — Official Go guidance
- [Writing Go CLIs With Just Enough Architecture](https://blog.carlana.net/post/2020/go-cli-how-to-and-advice/)
- [mitchellh/cli](https://github.com/mitchellh/cli) — CLI framework library
- [urfave/cli](https://github.com/urfave/cli) — Alternative CLI framework

### Additional Context
- Docker Desktop update notifications — Auto-update by default pattern
- kubectl version compatibility — Client/server version skew handling
- Push notification frequency research — 10% disable rate per week finding
