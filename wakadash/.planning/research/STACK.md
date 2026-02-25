# Technology Stack: Version Update Check

**Project:** wakadash (v2.2 milestone)
**Researched:** 2026-02-25
**Confidence:** HIGH

## Recommended Stack

### Semantic Version Comparison
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| golang.org/x/mod/semver | v0.33.0 | Parse and compare semver strings | Official Go team package, zero external dependencies beyond stdlib, designed for Go module system, requires "v" prefix which matches GitHub tags |

**Rationale:** Use the official Go team's semver package because:
1. **Zero heavyweight dependencies** - Only depends on Go stdlib
2. **Maintenance guarantee** - Maintained by Go team, guaranteed compatibility
3. **Perfect fit** - Requires "v" prefix (e.g., "v1.2.3") which matches GitHub release tags exactly
4. **Battle-tested** - Used by 1,821+ packages including Go module tooling itself
5. **Simple API** - Just `semver.Compare(v1, v2)` returns -1/0/+1

**Alternative considered:** `github.com/Masterminds/semver/v3` - More features (constraints, ranges) but unnecessary complexity for simple version comparison. Adds external dependency without benefit.

**Installation:**
```bash
go get golang.org/x/mod/semver@latest
```

### GitHub API Client
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| net/http (stdlib) + encoding/json (stdlib) | Go 1.24.2 | Fetch latest release from GitHub API | Already available, no new dependencies, sufficient for single API endpoint |

**Rationale:** Use standard library HTTP client because:
1. **Already in use** - Existing `internal/api/client.go` demonstrates pattern
2. **Minimal requirements** - Only need one endpoint: `GET /repos/{owner}/{repo}/releases/latest`
3. **No authentication needed** - GitHub releases API is public (no rate limit concerns for read-only)
4. **Proven pattern** - Project already uses `http.Client` with timeout, just extend it
5. **Avoid bloat** - `google/go-github` is well-maintained but adds 2 dependencies for 1 endpoint

**Alternative considered:** `github.com/google/go-github/v83` (latest: v83.0.0, Feb 2026) - Full-featured GitHub API client. Only adds 2 lightweight dependencies (`go-cmp`, `go-querystring`) and provides type-safe API. **Consider if** project needs multiple GitHub API features in future, but overkill for single endpoint now.

**Installation:**
```bash
# No installation needed - use stdlib
```

### Async Execution
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| goroutines + channels (stdlib) | Go 1.24.2 | Non-blocking version check | Built-in concurrency, integrates with Bubble Tea's Cmd pattern |

**Rationale:** Use goroutines because:
1. **Already integrated** - Bubble Tea uses `tea.Cmd` which wraps goroutines
2. **Project pattern** - Existing code in `internal/tui/commands.go` shows async WakaTime API calls
3. **Zero overhead** - No additional libraries needed
4. **Timeout control** - Use `context.WithTimeout` for 5-second max check duration

**Pattern from existing codebase:**
```go
// See internal/tui/commands.go for reference
func checkVersionCmd() tea.Msg {
    // Network call in goroutine, return result as message
}
```

## Implementation Approach

### Minimal Stdlib Implementation (RECOMMENDED)

**What to build:**
```go
package version

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "golang.org/x/mod/semver"
)

type Release struct {
    TagName string `json:"tag_name"`  // e.g., "v1.2.3"
}

func CheckUpdate(currentVersion, owner, repo string) (newer bool, latest string, err error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return false, "", err  // Fail silently - don't block startup
    }
    defer resp.Body.Close()

    var release Release
    if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
        return false, "", err
    }

    // Both versions must have "v" prefix for semver.Compare
    if !semver.IsValid(currentVersion) || !semver.IsValid(release.TagName) {
        return false, "", fmt.Errorf("invalid version format")
    }

    if semver.Compare(release.TagName, currentVersion) > 0 {
        return true, release.TagName, nil
    }

    return false, currentVersion, nil
}
```

**Integration points:**
1. **Existing HTTP pattern** - Mirrors `internal/api/client.go` timeout/error handling
2. **Bubble Tea command** - Add to `internal/tui/commands.go` alongside `fetchStatsCmd`
3. **Message handling** - Add `versionCheckMsg` type to `internal/tui/messages.go`
4. **Status bar** - Display in existing status bar when update available

### What NOT to Add

❌ **go-github** - 83.0.0 is current (Feb 2026), lightweight (2 deps), but unnecessary for 1 endpoint
❌ **tcnksm/go-latest** - Unmaintained since 2015, adds abstraction without benefit
❌ **Masterminds/semver** - Feature-rich but overkill for simple A > B comparison
❌ **blang/semver** - Extra strictness not needed, golang.org/x/mod/semver sufficient
❌ **OAuth/auth client** - Public API, no auth needed for releases
❌ **Retry logic** - Single attempt, fail gracefully if network unavailable

## Configuration

### Constants
```go
const (
    GitHubOwner = "b00y0h"
    GitHubRepo  = "wakadash"
    VersionCheckTimeout = 5 * time.Second
)
```

### Current Version Embedding
```go
// Use Go 1.18+ build-time version injection
var Version = "dev"  // Override with -ldflags "-X main.Version=v1.2.3" at build time
```

## Error Handling Strategy

**Fail silently** - Version check failures should never prevent dashboard from working:

```go
// In Bubble Tea Update()
case versionCheckMsg:
    if msg.err != nil {
        // Log error but don't show to user
        // Dashboard continues normally
        return m, nil
    }
    if msg.newer {
        m.updateAvailable = true
        m.latestVersion = msg.version
    }
    return m, nil
```

## Network Best Practices Applied

Based on 2026 HTTP client best practices:

1. **Always set timeout** - Never use `http.DefaultClient` (infinite timeout)
   - Client timeout: 10 seconds
   - Context timeout: 5 seconds (extra safety layer)

2. **Use context for cancellation** - `context.WithTimeout` for request-level control

3. **Reuse client** - Create once, reuse (connection pooling)

4. **Don't block UI** - Run in goroutine via `tea.Cmd`

5. **Background context pattern** - Use `context.Background()` not request context (this runs async after startup)

## Sources

**HIGH confidence** - Official documentation and current packages:
- [golang.org/x/mod/semver documentation](https://pkg.go.dev/golang.org/x/mod/semver) - Official Go team semver package
- [GitHub Releases API](https://docs.github.com/en/rest/releases/releases) - `/repos/{owner}/{repo}/releases/latest` endpoint
- [google/go-github v83.0.0](https://github.com/google/go-github) - Alternative if multi-endpoint needs arise
- [Go HTTP client timeout best practices (Feb 2026)](https://oneuptime.com/blog/post/2026-02-01-go-http-client-timeouts/view) - Current recommendations
- [Go context timeout patterns (Feb 2026)](https://oneuptime.com/blog/post/2026-02-06-fix-context-canceled-errors-otel-go-background-work/view) - Background work patterns

**MEDIUM confidence** - Community resources:
- [tcnksm/go-latest](https://github.com/tcnksm/go-latest) - Unmaintained but demonstrates pattern
- [Go HTTP timeouts guide](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) - Comprehensive timeout explanation

## Dependencies Summary

**New dependencies required:**
```go
require (
    golang.org/x/mod v0.33.0  // For semver comparison only
)
```

**Total impact:** +1 dependency (zero external deps, stdlib extension)

**Existing capabilities to leverage:**
- `net/http` - HTTP client (already used in `internal/api/client.go`)
- `encoding/json` - JSON parsing (already used)
- `time` - Timeouts and durations (already used)
- `context` - Request cancellation (add for version check)
- Bubble Tea's `tea.Cmd` pattern - Async execution (already established)

## Anti-Patterns to Avoid

1. **Blocking startup** - Version check MUST be async, don't delay dashboard display
2. **Over-engineering** - Don't add full GitHub client for 1 endpoint
3. **Complex retry** - Single attempt sufficient, this is UX enhancement not critical path
4. **Version without "v" prefix** - Ensure build tags include "v" to match `golang.org/x/mod/semver` requirements
5. **Request context reuse** - Don't pass Bubble Tea message context to background goroutine (causes cancellation issues)
6. **Ignoring timeouts** - Always set both client and context timeouts
7. **Error surfacing** - Don't show network errors to user, fail gracefully

## Implementation Notes

### Bubble Tea Integration Pattern

```go
// Initial command on startup
func (m model) Init() tea.Cmd {
    return tea.Batch(
        fetchStatsCmd(m.client, m.range),
        checkForUpdateCmd(Version, GitHubOwner, GitHubRepo),  // New: async version check
    )
}

// Message type
type versionCheckCompleteMsg struct {
    newer   bool
    version string
    err     error
}

// Command (runs in goroutine)
func checkForUpdateCmd(current, owner, repo string) tea.Cmd {
    return func() tea.Msg {
        newer, latest, err := CheckUpdate(current, owner, repo)
        return versionCheckCompleteMsg{newer, latest, err}
    }
}

// Update handler
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case versionCheckCompleteMsg:
        if msg.err == nil && msg.newer {
            m.updateAvailable = true
            m.latestVersion = msg.version
        }
        // Always succeed - don't block on errors
        return m, nil
    }
}
```

### Status Bar Display

```go
// In status bar rendering (when m.updateAvailable == true)
fmt.Sprintf(
    "Update available: %s → %s | Run: brew upgrade wakadash",
    currentVersion, m.latestVersion,
)
```

This follows the gh CLI pattern: non-intrusive notification with actionable command.
