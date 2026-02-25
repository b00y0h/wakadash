# Architecture Patterns: Version Update Check Integration

**Domain:** Go TUI application with Bubble Tea framework
**Researched:** 2026-02-25
**Confidence:** HIGH

## Executive Summary

Version checking in Bubble Tea TUI applications follows the standard Elm Architecture async pattern: a command (tea.Cmd) launches a goroutine to check GitHub releases, sends the result back via a custom message type, and the Update handler processes it to update the model state. The View function then conditionally renders the notification in the status bar.

The key architectural insight: **version checking is just another async data fetch**, structurally identical to the existing stats fetching pattern already implemented in wakadash. The existing code demonstrates the exact pattern needed.

## Recommended Architecture

### Component Structure

```
┌─────────────────────────────────────────────────────────────┐
│                         main.go                              │
│  • Parse flags                                               │
│  • Load config                                               │
│  • Create API client                                         │
│  • Create version checker (NEW)                              │
│  • Create TUI model                                          │
│  • Run Bubble Tea program                                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    internal/version/                         │
│                     checker.go (NEW)                         │
│                                                              │
│  type Checker struct {                                       │
│    currentVersion string                                     │
│    githubRepo     string  // "b00y0h/wakadash"             │
│    httpClient     *http.Client                              │
│  }                                                           │
│                                                              │
│  func (c *Checker) CheckLatestRelease()                      │
│         (*ReleaseInfo, error)                                │
│                                                              │
│  type ReleaseInfo struct {                                   │
│    TagName  string  // "v2.2.0"                             │
│    URL      string  // GitHub release page                  │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     internal/tui/                            │
│                      model.go                                │
│                                                              │
│  type Model struct {                                         │
│    // ... existing fields ...                                │
│    updateAvailable *UpdateInfo  // NEW: nil or update info  │
│  }                                                           │
│                                                              │
│  type UpdateInfo struct {                                    │
│    CurrentVersion string                                     │
│    LatestVersion  string                                     │
│    ReleaseURL     string                                     │
│  }                                                           │
│                                                              │
│  func (m Model) Init() tea.Cmd {                             │
│    return tea.Batch(                                         │
│      fetchStatsCmd(...),           // Existing              │
│      fetchDurationsCmd(...),       // Existing              │
│      fetchSummaryCmd(...),         // Existing              │
│      checkVersionCmd(versionChecker), // NEW                │
│      m.spinner.Tick,               // Existing              │
│      tickEverySecond(),            // Existing              │
│    )                                                         │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     internal/tui/                            │
│                    messages.go (MODIFY)                      │
│                                                              │
│  // ADD new message types:                                   │
│                                                              │
│  type versionCheckCompleteMsg struct {                       │
│    updateInfo *UpdateInfo  // nil if no update              │
│  }                                                           │
│                                                              │
│  type versionCheckFailedMsg struct {                         │
│    err error  // Log but don't display to user              │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     internal/tui/                            │
│                   commands.go (MODIFY)                       │
│                                                              │
│  func checkVersionCmd(checker *version.Checker) tea.Cmd {    │
│    return func() tea.Msg {                                   │
│      defer func() {                                          │
│        if r := recover(); r != nil {                         │
│          return versionCheckFailedMsg{                       │
│            err: fmt.Errorf("panic: %v", r)                   │
│          }                                                   │
│        }                                                     │
│      }()                                                     │
│                                                              │
│      releaseInfo, err := checker.CheckLatestRelease()        │
│      if err != nil {                                         │
│        return versionCheckFailedMsg{err: err}                │
│      }                                                       │
│                                                              │
│      if releaseInfo != nil && isNewer(releaseInfo.TagName) { │
│        return versionCheckCompleteMsg{                       │
│          updateInfo: &UpdateInfo{                            │
│            CurrentVersion: checker.currentVersion,           │
│            LatestVersion:  releaseInfo.TagName,              │
│            ReleaseURL:     releaseInfo.URL,                  │
│          },                                                  │
│        }                                                     │
│      }                                                       │
│                                                              │
│      return versionCheckCompleteMsg{updateInfo: nil}         │
│    }                                                         │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              internal/tui/model.go Update()                  │
│                                                              │
│  func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {  │
│    switch msg := msg.(type) {                                │
│    // ... existing cases ...                                 │
│                                                              │
│    case versionCheckCompleteMsg:                             │
│      m.updateAvailable = msg.updateInfo                      │
│      return m, nil                                           │
│                                                              │
│    case versionCheckFailedMsg:                               │
│      // Log silently, don't show to user                     │
│      // (network errors shouldn't interrupt dashboard)       │
│      return m, nil                                           │
│    }                                                         │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│         internal/tui/model.go renderStatusBar()              │
│                                                              │
│  func (m Model) renderStatusBar() string {                   │
│    var lines []string                                        │
│                                                              │
│    // Primary status line (existing logic)                   │
│    statusLine := m.buildPrimaryStatus()                      │
│    lines = append(lines, statusLine)                         │
│                                                              │
│    // Update notification (NEW)                              │
│    if m.updateAvailable != nil {                             │
│      updateLine := m.buildUpdateNotification()               │
│      lines = append(lines, updateLine)                       │
│    }                                                         │
│                                                              │
│    return lipgloss.JoinVertical(lipgloss.Left, lines...)     │
│  }                                                           │
│                                                              │
│  func (m Model) buildUpdateNotification() string {           │
│    line1 := fmt.Sprintf(                                     │
│      "A new version is available: %s → %s",                  │
│      m.updateAvailable.CurrentVersion,                       │
│      m.updateAvailable.LatestVersion,                        │
│    )                                                         │
│    line2 := "Run: brew upgrade wakadash"                     │
│                                                              │
│    // Style with theme's accent color                        │
│    styled := lipgloss.NewStyle().                            │
│      Foreground(m.theme.Accent).                             │
│      Render(line1 + "\n" + line2)                            │
│                                                              │
│    return styled                                             │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
```

## Data Flow for Async Version Check

### Sequence Diagram

```
main.go                 model.Init()         checkVersionCmd()        GitHub API           model.Update()        model.View()
   │                         │                      │                      │                      │                   │
   │ Create checker          │                      │                      │                      │                   │
   ├────────────────────────►│                      │                      │                      │                   │
   │                         │                      │                      │                      │                   │
   │ Start program           │                      │                      │                      │                   │
   ├────────────────────────►│                      │                      │                      │                   │
   │                         │                      │                      │                      │                   │
   │                         │ Launch goroutine     │                      │                      │                   │
   │                         ├─────────────────────►│                      │                      │                   │
   │                         │                      │                      │                      │                   │
   │                         │                      │ GET /releases/latest │                      │                   │
   │                         │                      ├─────────────────────►│                      │                   │
   │                         │                      │                      │                      │                   │
   │                         │ (TUI renders without update notification)  │                      │                   │
   │                         │                                            │                      │                   │
   │                         │                      │   200 + JSON        │                      │                   │
   │                         │                      │◄─────────────────────┤                      │                   │
   │                         │                      │                      │                      │                   │
   │                         │                      │ Compare versions     │                      │                   │
   │                         │                      │ (semver check)       │                      │                   │
   │                         │                      │                      │                      │                   │
   │                         │                      │ Send versionCheckCompleteMsg               │                   │
   │                         │                      ├────────────────────────────────────────────►│                   │
   │                         │                      │                      │                      │                   │
   │                         │                      │                      │    Update model:    │                   │
   │                         │                      │                      │  m.updateAvailable  │                   │
   │                         │                      │                      │    = &UpdateInfo    │                   │
   │                         │                      │                      │                      │                   │
   │                         │                      │                      │                      │  Trigger refresh  │
   │                         │                      │                      │                      ├──────────────────►│
   │                         │                      │                      │                      │                   │
   │                         │                      │                      │                      │ Render status bar │
   │                         │                      │                      │                      │ with update notice│
   │                         │                      │                      │                      │                   │
```

### Timing Characteristics

| Phase | Duration | Notes |
|-------|----------|-------|
| **Dashboard startup** | ~50-200ms | Config load, client init, TUI setup |
| **Version check launch** | ~1ms | Goroutine spawn (non-blocking) |
| **Initial render** | Immediate | Dashboard shows without update notification |
| **GitHub API call** | ~200-800ms | Network latency + API response time |
| **Version comparison** | <1ms | Semantic version parsing |
| **Model update** | <1ms | Set `updateAvailable` field |
| **Re-render** | <10ms | Status bar updated with notification |

**Total time to notification:** 200-800ms (async, doesn't block startup)

## Integration Points

### New Components

| Component | File | Responsibility |
|-----------|------|----------------|
| **Version Checker** | `internal/version/checker.go` | HTTP client for GitHub Releases API |
| **Release Info Type** | `internal/version/checker.go` | Release metadata (version, URL) |
| **Version Command** | `internal/tui/commands.go` | Bubble Tea command wrapper |
| **Version Messages** | `internal/tui/messages.go` | Success/failure message types |
| **Update Info Type** | `internal/tui/model.go` | Model state for update notification |

### Modified Components

| Component | File | Modification |
|-----------|------|--------------|
| **Main** | `cmd/wakadash/main.go` | Create version checker, pass to model |
| **Model** | `internal/tui/model.go` | Add `updateAvailable` field |
| **Init** | `internal/tui/model.go` | Add `checkVersionCmd` to batch |
| **Update** | `internal/tui/model.go` | Handle version check messages |
| **Status Bar** | `internal/tui/model.go` | Render multi-line with update notice |

### External Dependencies

| Dependency | Purpose | Why Needed |
|------------|---------|------------|
| None (use stdlib) | GitHub Releases API | Simple HTTP GET, JSON parsing with `encoding/json` |
| `github.com/hashicorp/go-version` | Semantic version comparison | Robust SemVer parsing (e.g., "v2.2.0" > "v2.1.5") |

**Note:** The existing codebase already uses `net/http` and `encoding/json` for API calls. Version checking follows the same pattern.

## Patterns to Follow

### Pattern 1: Async Command with Panic Recovery

**What:** Launch goroutine via `tea.Cmd`, wrap in panic recovery, return result via custom message.

**When:** Any I/O operation (network, disk) that might fail or take time.

**Why:** Keeps UI responsive, prevents terminal corruption from panics.

**Example:**
```go
func checkVersionCmd(checker *version.Checker) tea.Cmd {
	return func() tea.Msg {
		defer func() {
			if r := recover(); r != nil {
				return versionCheckFailedMsg{err: fmt.Errorf("panic: %v", r)}
			}
		}()

		result, err := checker.CheckLatestRelease()
		if err != nil {
			return versionCheckFailedMsg{err: err}
		}
		return versionCheckCompleteMsg{result: result}
	}
}
```

**Existing usage in codebase:**
- `fetchStatsCmd()` in `internal/tui/commands.go` lines 58-81
- `fetchDurationsCmd()` in `internal/tui/commands.go` lines 101-124
- `fetchSummaryCmd()` in `internal/tui/commands.go` lines 128-151

### Pattern 2: Silent Failure for Non-Critical Operations

**What:** Log errors from version check but don't display to user.

**When:** Operation is informational, failure doesn't affect core functionality.

**Why:** Network failures shouldn't interrupt the dashboard experience.

**Example:**
```go
case versionCheckFailedMsg:
	// Version check failed - log but don't show error to user
	// Dashboard continues working normally
	if msg.err != nil {
		log.Printf("version check failed (silent): %v", msg.err)
	}
	return m, nil
```

**Contrast with critical failures:**
- Stats fetch errors ARE shown (dashboard exists to show stats)
- Version check errors are NOT shown (dashboard works without version info)

### Pattern 3: Conditional Rendering Based on Model State

**What:** Check model field in View(), render additional UI elements if present.

**When:** Optional UI components that may or may not be shown.

**Why:** Declarative rendering - View() is pure function of model state.

**Example:**
```go
func (m Model) renderStatusBar() string {
	var lines []string

	// Always show primary status
	lines = append(lines, m.buildPrimaryStatus())

	// Conditionally show update notification
	if m.updateAvailable != nil {
		lines = append(lines, m.buildUpdateNotification())
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
```

**Existing usage in codebase:**
- Panel visibility toggles: `if m.showLanguages { ... }` (model.go lines 447-449)
- Help overlay: `if m.showHelp { return m.renderHelp() }` (model.go line 385)
- Theme picker: `if m.showPicker { return m.picker.View() }` (model.go lines 358-360)

### Pattern 4: Batch Multiple Init Commands

**What:** Use `tea.Batch()` to launch multiple async operations on startup.

**When:** Multiple independent async operations needed at init.

**Why:** All operations start immediately in parallel.

**Example:**
```go
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchStatsCmd(m.client, m.rangeStr),    // Stats API
		fetchDurationsCmd(m.client),            // Durations API
		fetchSummaryCmd(m.client),              // Summary API
		checkVersionCmd(m.versionChecker),      // NEW: GitHub API
		m.spinner.Tick,                         // Animation
		tickEverySecond(),                      // Timer
	)
}
```

**Existing usage:** `model.go` lines 142-158 (all async commands launched in parallel)

## Anti-Patterns to Avoid

### Anti-Pattern 1: Blocking Init with Synchronous Network Call

**What goes wrong:**
```go
// DON'T DO THIS
func (m Model) Init() tea.Cmd {
	// This blocks the entire TUI startup!
	updateInfo, _ := checkVersionSync(m.checker)  // BLOCKS
	m.updateAvailable = updateInfo
	return m.spinner.Tick
}
```

**Why bad:** Dashboard won't appear until GitHub API responds (200-800ms delay, or longer if network is slow).

**Consequences:**
- Poor UX: blank screen on startup
- Timeout issues: if GitHub is unreachable, startup fails
- Violates Bubble Tea async pattern

**Instead:**
```go
// DO THIS
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		checkVersionCmd(m.checker),  // Async - returns immediately
		// ... other commands
	)
}
```

### Anti-Pattern 2: Showing Version Check Errors to User

**What goes wrong:**
```go
// DON'T DO THIS
case versionCheckFailedMsg:
	m.err = msg.err  // Shows error in main dashboard
	return m, nil
```

**Why bad:** Network failures are common and expected. User can't fix them.

**Consequences:**
- Dashboard shows error state for non-critical feature
- User sees "network error" when they just want to see stats
- Clutters UI with irrelevant failures

**Instead:**
```go
// DO THIS
case versionCheckFailedMsg:
	// Silent - version check is informational, not critical
	return m, nil
```

### Anti-Pattern 3: Retrying Failed Version Checks

**What goes wrong:**
```go
// DON'T DO THIS
case versionCheckFailedMsg:
	// Retry on failure
	return m, tea.Batch(
		checkVersionCmd(m.checker),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return retryVersionCheckMsg{}
		}),
	)
```

**Why bad:**
- GitHub API rate limits (60 requests/hour unauthenticated)
- Wastes network bandwidth
- Version check runs once at startup - no need to retry

**Consequences:**
- Hits rate limits
- Drains mobile data
- Adds complexity without value

**Instead:**
```go
// DO THIS
case versionCheckFailedMsg:
	// Single attempt only - user can restart app if needed
	return m, nil
```

### Anti-Pattern 4: Complex Version String Parsing

**What goes wrong:**
```go
// DON'T DO THIS
func isNewer(latest string) bool {
	// Fragile string manipulation
	latestParts := strings.Split(strings.TrimPrefix(latest, "v"), ".")
	currentParts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	// ... manual comparison logic
}
```

**Why bad:**
- Doesn't handle pre-release versions (v2.2.0-beta.1)
- Doesn't handle build metadata (v2.2.0+20210101)
- Edge cases: "v2.10.0" vs "v2.2.0" (string comparison fails)

**Consequences:**
- Incorrect version comparisons
- False positives/negatives
- SemVer violations

**Instead:**
```go
// DO THIS - use hashicorp/go-version
func isNewer(latest string) bool {
	current, _ := version.NewVersion(currentVersion)
	latestVer, _ := version.NewVersion(latest)
	return latestVer.GreaterThan(current)
}
```

## GitHub Releases API Integration

### Endpoint

```
GET https://api.github.com/repos/{owner}/{repo}/releases/latest
```

**Example:** `https://api.github.com/repos/b00y0h/wakadash/releases/latest`

### Request

```go
req, err := http.NewRequest("GET", url, nil)
req.Header.Set("Accept", "application/vnd.github+json")
req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
// No authentication needed for public repos (60 req/hour)
```

### Response (relevant fields)

```json
{
  "tag_name": "v2.2.0",
  "name": "Release v2.2.0",
  "html_url": "https://github.com/b00y0h/wakadash/releases/tag/v2.2.0",
  "published_at": "2026-02-25T10:00:00Z"
}
```

### Rate Limiting

| Auth Status | Limit | Window | Notes |
|-------------|-------|--------|-------|
| **Unauthenticated** | 60 requests | 1 hour | Sufficient for single check at startup |
| **Authenticated** | 5,000 requests | 1 hour | Not needed for this use case |

**Strategy:** Single unauthenticated request at startup. No retry logic.

### Error Handling

| Status Code | Meaning | Action |
|-------------|---------|--------|
| 200 | Success | Parse response, compare versions |
| 403 | Rate limit exceeded | Silent failure, try next app launch |
| 404 | Repo/release not found | Silent failure (repo might be private) |
| 500-599 | GitHub server error | Silent failure, transient issue |
| Network timeout | No connection | Silent failure, offline or slow network |

**All errors are silent** - version check is informational, not critical.

## Suggested Build Order

### Phase 1: Core Version Checker (No UI)

**Goal:** Verify GitHub API integration works.

**Components:**
1. `internal/version/checker.go` - HTTP client, API call, JSON parsing
2. `internal/version/checker_test.go` - Unit tests with mock HTTP
3. Manual testing: verify API response parsing

**Why first:** Isolate GitHub API integration from TUI complexity. Can test independently.

**Exit criteria:**
- `CheckLatestRelease()` returns correct version for real repo
- Tests verify JSON parsing, error handling, version comparison

### Phase 2: Bubble Tea Integration (Messages & Commands)

**Goal:** Wire version checker into Bubble Tea lifecycle.

**Components:**
1. Add message types to `internal/tui/messages.go`
2. Add command wrapper to `internal/tui/commands.go`
3. Add `updateAvailable` field to `internal/tui/model.go`
4. Add message handlers to `Update()` in `internal/tui/model.go`
5. Call `checkVersionCmd()` from `Init()` in `internal/tui/model.go`

**Why second:** Establishes data flow before UI rendering.

**Exit criteria:**
- Version check launches at startup (verified with debug logging)
- Model state updates when newer version detected
- No impact on dashboard if check fails

### Phase 3: Status Bar UI (Multi-line Rendering)

**Goal:** Display update notification to user.

**Components:**
1. Refactor `renderStatusBar()` to support multi-line output
2. Add `buildUpdateNotification()` helper
3. Style with theme colors (accent color for visibility)
4. Test with different terminal widths

**Why third:** UI layer depends on data flow working correctly.

**Exit criteria:**
- Update notification appears in status bar when newer version available
- Multi-line rendering doesn't break layout
- Notification styled consistently with theme

### Phase 4: Main Integration & Testing

**Goal:** End-to-end functionality.

**Components:**
1. Modify `cmd/wakadash/main.go` to create version checker
2. Pass checker to `NewModel()`
3. Integration testing:
   - Test with real GitHub API (current version)
   - Test with mock "newer" version (simulate update available)
   - Test with network failure (verify silent failure)
   - Test with rate limit (verify graceful handling)

**Why fourth:** Complete integration after all pieces tested individually.

**Exit criteria:**
- Running `wakadash` shows update notification when newer version exists
- Dashboard works normally when version check fails
- No blocking delays at startup

### Dependency Graph

```
Phase 1: Version Checker
    │
    ▼
Phase 2: Bubble Tea Integration
    │
    ▼
Phase 3: Status Bar UI
    │
    ▼
Phase 4: Main Integration
```

**Total estimated complexity:** 4-6 hours of development + testing

## Scalability Considerations

### At Launch (0-1K users)

**Concern:** GitHub API rate limits (60 requests/hour unauthenticated)

**Impact:**
- Single check at startup = 1 request per app launch
- 60 users/hour can check for updates
- Sufficient for current scale

**Approach:** No authentication, single check at startup.

### At 10K users

**Concern:** If 10K users launch app simultaneously, all hit GitHub API.

**Impact:**
- GitHub can handle 10K requests (enterprise infrastructure)
- Individual users might hit rate limit if restarting frequently
- No action needed - rate limits are per-IP, not per-repo

**Approach:** Continue current strategy. Monitor GitHub API status.

### At 100K+ users

**Concern:** High request volume, potential rate limiting for individual users.

**Mitigation options:**
1. **Authenticated requests** (5,000/hour) - but requires OAuth flow
2. **Caching layer** - CDN with 5-minute TTL for latest release JSON
3. **Skip check** if last check was <24 hours ago (store in `~/.wakadash/last_version_check`)

**Recommendation:** Implement option 3 (local caching) when user base reaches 50K+.

**Implementation:**
```go
// Check last version check timestamp
lastCheck, _ := readLastCheckTime("~/.wakadash/last_version_check")
if time.Since(lastCheck) < 24*time.Hour {
	// Skip check - too recent
	return versionCheckSkippedMsg{}
}

// Otherwise, perform check and update timestamp
```

### Future Enhancements (Out of Scope for v2.2)

| Enhancement | Value | Complexity | Notes |
|-------------|-------|------------|-------|
| **Cache check timestamp** | Reduce API calls | Low | Store last check in `~/.wakadash/` |
| **Dismiss notification** | User control | Low | Add "d" key to dismiss, store in config |
| **Show release notes** | User awareness | Medium | Fetch release body from API, show in help overlay |
| **Auto-update** | Convenience | High | Security concerns, platform differences |

**Recommendation:** Add caching + dismissal in v2.3 if user feedback requests it.

## Sources

### Bubble Tea Patterns
- [Bubble Tea Official Repo](https://github.com/charmbracelet/bubbletea) - TUI framework architecture
- [How to Build Command Line Tools with Bubbletea in Go](https://oneuptime.com/blog/post/2026-01-30-how-to-build-command-line-tools-with-bubbletea-in-go/view) - Async patterns and event loop
- [The Bubbletea (TUI) State Machine pattern](https://zackproser.com/blog/bubbletea-state-machine) - State machine for robust async operations

### Version Checking Libraries
- [go-github-selfupdate (rhysd)](https://github.com/rhysd/go-github-selfupdate) - Self-update mechanism with GitHub Releases
- [go-selfupdate (creativeprojects)](https://github.com/creativeprojects/go-selfupdate) - Enhanced fork with GitLab support
- [go-update-checker](https://github.com/Christian1984/go-update-checker) - Version comparison with caching
- [go-latest](https://github.com/tcnksm/go-latest) - Version checking from multiple sources
- [hashicorp/go-version](https://github.com/hashicorp/go-version) - Semantic version parsing and comparison

### GitHub API
- [google/go-github Library](https://github.com/google/go-github) - Official Go client for GitHub API
- [repos_releases.go](https://github.com/google/go-github/blob/master/github/repos_releases.go) - Release-related API methods
- [GitHub CLI](https://github.com/cli/cli) - Official CLI reference implementation

### Documentation Quality
- **HIGH confidence:** Bubble Tea async patterns (official docs + code examples)
- **HIGH confidence:** GitHub Releases API (tested in production by many projects)
- **MEDIUM confidence:** Scale considerations (extrapolated from rate limits, not tested at scale)
