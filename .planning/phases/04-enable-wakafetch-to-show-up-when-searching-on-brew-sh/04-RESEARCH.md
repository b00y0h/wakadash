# Phase 4: Repository Setup - Research

**Researched:** 2026-02-19
**Domain:** Go repository creation, code porting, GoReleaser multi-platform release automation
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Repository structure:**
- Single-binary repository — just wakadash, no multi-tool setup
- Go module: `github.com/b00y0h/wakadash`
- Standard documentation set: README.md, LICENSE, CONTRIBUTING.md, CHANGELOG.md

**Code porting strategy:**
- Clean rewrite of WakaTime API client — don't just copy wakafetch code
- Dashboard-only scope — only port what the TUI needs, no extra features
- Config: Read directly from `~/.wakatime.cfg` (WakaTime's standard location)
- No wakafetch config migration needed — wakadash doesn't have its own config

**CLI interface design:**
- Default action: Launch dashboard immediately (no subcommand required)
- No one-shot mode — wakadash is purely a TUI dashboard tool
- Version flag: `--version` shows version + git commit + build date + Go version

### Claude's Discretion
- Directory layout (flat vs cmd/internal/pkg structure)
- Package naming (generic vs branded)
- Help text verbosity

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

## Summary

Phase 4 creates the `b00y0h/wakadash` GitHub repository from scratch (not a fork), ports the WakaTime API client and config reader from wakafetch via clean rewrite, and establishes GoReleaser + GitHub Actions for tag-triggered multi-platform releases. The end state is a repository where `go build` works, `wakadash --help` shows usage, and pushing a `v*.*.*` tag produces GitHub release artifacts for darwin/linux x amd64/arm64.

The wakafetch codebase at `/workspace/wakafetch` is the source of truth. Its `api.go`, `config.go`, and `types/types.go` contain the patterns to rewrite. The existing architecture research (`STACK.md`, `ARCHITECTURE.md`) establishes that the final tool will use bubbletea — but Phase 4 only needs a minimal stub that compiles and satisfies `--help` and `--version`. The TUI implementation is Phase 5's responsibility.

GoReleaser v2 (current ~v2.13) is the established tool with full patterns already documented in the Phase 2 research. The `.goreleaser.yaml` and GitHub Actions workflow from Phase 2 transfer directly to wakadash with name substitution. The critical delta is: Phase 4 does NOT publish to a Homebrew tap (that's Phase 7), so the `brews:` section is omitted and no `HOMEBREW_TAP_TOKEN` is needed yet.

**Primary recommendation:** Create the repo with `gh repo create`, init the Go module as `github.com/b00y0h/wakadash`, write a minimal working CLI stub with `--help` and `--version`, rewrite the API client and config reader (clean copy-with-attribution from wakafetch patterns), add `.goreleaser.yaml` targeting darwin+linux x amd64+arm64, and add the GitHub Actions release workflow. Verify with a `v0.1.0` tag push.

## Standard Stack

### Core
| Tool/Library | Version | Purpose | Why Standard |
|-------------|---------|---------|--------------|
| Go stdlib only | go 1.24.3 (match wakafetch) | Language runtime, flag package, net/http, encoding | Zero external deps for Phase 4 — bubbletea added in Phase 5 |
| GoReleaser | v2 (~>v2.13) | Multi-platform build automation, checksums, GitHub release | Established in Phase 2; same tool, same version |
| goreleaser-action | v6 | GitHub Actions integration for GoReleaser | Official action, established in Phase 2 |
| actions/checkout | v4 | Git checkout with full history for changelog | Required by GoReleaser changelog |
| actions/setup-go | v5 | Go toolchain in CI | Official Go team action |
| gh CLI | v2.83.2 | Create GitHub repository, configure settings | Already available in environment |

### Supporting
| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| Fine-grained PAT | n/a | Contents read/write for wakadash repo | Only if cross-repo publishing needed — NOT needed in Phase 4 |
| GITHUB_TOKEN | n/a | Create GitHub release and upload artifacts | Sufficient for same-repo release creation |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Go stdlib flag package | cobra, pflag | wakafetch uses stdlib flag — no external deps needed for Phase 4 stub |
| GoReleaser | Manual cross-compilation scripts | GoReleaser handles edge cases, consistent naming, parallel builds |
| Flat directory layout | cmd/internal/pkg structure | Flat is simpler for single-binary tool; cmd/ is worth it once TUI packages added in Phase 5 |

**Go module initialization:**
```bash
mkdir wakadash && cd wakadash
git init
go mod init github.com/b00y0h/wakadash
```

## Architecture Patterns

### Recommended Project Structure for Phase 4

The decision point is flat vs `cmd/` structure. Given that Phase 5 will add a `dashboard/` or `internal/` package, use `cmd/` structure from the start to avoid a refactor.

```
wakadash/
├── .github/
│   └── workflows/
│       └── release.yml         # Tag-triggered release workflow
├── cmd/
│   └── wakadash/
│       └── main.go             # Entry point: --help, --version, stub TUI launch
├── internal/
│   ├── api/
│   │   └── client.go           # WakaTime API client (rewritten from wakafetch api.go)
│   ├── config/
│   │   └── config.go           # ~/.wakatime.cfg reader (rewritten from wakafetch config.go)
│   └── types/
│       └── types.go            # API response types (ported from wakafetch types/types.go)
├── .goreleaser.yaml            # GoReleaser config (no brews section in Phase 4)
├── .gitignore                  # dist/, *.test, etc.
├── go.mod                      # module github.com/b00y0h/wakadash
├── go.sum                      # empty until external deps added
├── LICENSE                     # MIT (with attribution to wakafetch)
├── README.md                   # Basic usage, install from source
├── CONTRIBUTING.md             # Contribution guide
└── CHANGELOG.md               # Initial empty changelog
```

**Why `internal/` for API client and config:** These packages will be consumed by the bubbletea dashboard in Phase 5. The `internal/` convention prevents external import by other modules (appropriate for implementation details) and signals to the planner that these are non-public packages.

**Why `cmd/wakadash/main.go` not root `main.go`:** Follows Go convention for applications. The Phase 2 GoReleaser research established `main: ./cmd/wakafetch` as the pattern. Using `cmd/wakadash/main.go` means `go build ./cmd/wakadash` and GoReleaser `main: ./cmd/wakadash`.

### Pattern 1: Minimal Working CLI Stub (Phase 4 deliverable)

**What:** A `main.go` that satisfies `--help`, `--version`, and prints a placeholder for the dashboard.
**When to use:** Phase 4 only — gives GoReleaser something real to build and tests the release pipeline before Phase 5 implements the actual TUI.

```go
// Source: wakafetch flags.go pattern + GoReleaser version injection cookbook
// File: cmd/wakadash/main.go
package main

import (
    "flag"
    "fmt"
    "os"
    "runtime"
)

// Injected by GoReleaser via ldflags
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    versionFlag := flag.Bool("version", false, "Show version information")
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: wakadash [options]\n\n")
        fmt.Fprintf(os.Stderr, "A live terminal dashboard for WakaTime coding stats.\n\n")
        fmt.Fprintf(os.Stderr, "Options:\n")
        flag.PrintDefaults()
    }
    flag.Parse()

    if *versionFlag {
        fmt.Printf("wakadash %s\n  commit: %s\n  built:  %s\n  go:     %s\n",
            version, commit, date, runtime.Version())
        os.Exit(0)
    }

    // Phase 5 will replace this with tea.NewProgram(...)
    fmt.Println("wakadash: dashboard launching... (not yet implemented)")
    fmt.Println("Run 'wakadash --help' for usage.")
}
```

### Pattern 2: WakaTime API Client (clean rewrite)

**What:** A rewrite of wakafetch's `api.go` scoped to what the TUI needs.
**When to use:** Phase 4 sets up the types and HTTP client; Phase 5 wraps them as `tea.Cmd`.

```go
// File: internal/api/client.go
// Attribution: Rewritten from github.com/sahaj-b/wakafetch (MIT License)
package api

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/b00y0h/wakadash/internal/types"
)

const timeout = 10 * time.Second

// Client holds connection configuration for the WakaTime API.
type Client struct {
    APIKey string
    APIURL string
}

func (c *Client) FetchStats(rangeStr string) (*types.StatsResponse, error) {
    url := c.buildURL(fmt.Sprintf("/v1/users/current/stats/%s", rangeStr))
    return fetchJSON[types.StatsResponse](c.APIKey, url)
}

func (c *Client) FetchSummary(days int) (*types.SummaryResponse, error) {
    today := time.Now()
    start := today.AddDate(0, 0, -days+1).Format("2006-01-02")
    end := today.Format("2006-01-02")
    url := c.buildURL(fmt.Sprintf("/v1/users/current/summaries?start=%s&end=%s", start, end))
    return fetchJSON[types.SummaryResponse](c.APIKey, url)
}

func (c *Client) buildURL(path string) string {
    base := strings.TrimSuffix(c.APIURL, "/")
    if strings.HasSuffix(base, "/v1") {
        // Wakapi compat: strip /v1 to avoid double prefix
        base = strings.TrimSuffix(base, "/v1")
        path = "/v1" + strings.TrimPrefix(path, "/v1")
    }
    return base + path
}

func fetchJSON[T any](apiKey, url string) (*T, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(apiKey)))

    client := &http.Client{Timeout: timeout}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error %s", resp.Status)
    }

    var result T
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }
    return &result, nil
}
```

### Pattern 3: Config Reader (clean rewrite)

**What:** Reads `~/.wakatime.cfg` for `api_url` and `api_key`. Identical semantics to wakafetch `config.go` but scoped to new module.

```go
// File: internal/config/config.go
// Attribution: Rewritten from github.com/sahaj-b/wakafetch (MIT License)
package config

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

type Config struct {
    APIURL string
    APIKey string
}

func Load() (*Config, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("find home directory: %w", err)
    }
    path := filepath.Join(home, ".wakatime.cfg")
    return loadFile(path)
}

func loadFile(path string) (*Config, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("open %s: %w", path, err)
    }
    defer f.Close()

    cfg := &Config{}
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
            continue
        }
        key, val, ok := strings.Cut(line, "=")
        if !ok {
            continue
        }
        switch strings.TrimSpace(key) {
        case "api_url":
            cfg.APIURL = strings.TrimSpace(val)
        case "api_key":
            cfg.APIKey = strings.TrimSpace(val)
        }
    }

    if cfg.APIURL == "" {
        return nil, fmt.Errorf("api_url not found in %s", path)
    }
    if cfg.APIKey == "" {
        return nil, fmt.Errorf("api_key not found in %s", path)
    }

    // Wakapi compat URL normalization (matches wakafetch behavior)
    if cfg.APIURL == "https://wakapi.dev/api" {
        cfg.APIURL = "https://wakapi.dev/api/compat/wakatime"
    }
    cfg.APIURL = strings.TrimSuffix(cfg.APIURL, "/")

    return cfg, nil
}
```

### Pattern 4: GoReleaser Config (Phase 4 — no Homebrew tap)

**What:** `.goreleaser.yaml` for multi-platform builds. No `brews:` section (tap publishing is Phase 7).

```yaml
# Source: Phase 2 research patterns, GoReleaser official docs
# .goreleaser.yaml
version: 2

builds:
  - id: wakadash
    main: ./cmd/wakadash
    binary: wakadash
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.CommitDate}}

archives:
  - id: wakadash
    formats: [tar.gz]
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    files:
      - README.md
      - LICENSE
      - CHANGELOG.md

checksum:
  name_template: "{{ .ProjectName }}_{{ .Version }}_checksums.txt"
  algorithm: sha256

changelog:
  use: git
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^ci:'
      - '^chore:'
```

**Note on ldflags path:** Because `main.go` is in `cmd/wakadash/`, the `version`, `commit`, `date` variables live in `package main` of that file. The ldflag `-X main.version={{.Version}}` correctly targets `main.version` regardless of the subdirectory — Go's linker uses the package path `main`, not the file path.

### Pattern 5: GitHub Actions Release Workflow (Phase 4)

```yaml
# Source: Phase 2 research, GoReleaser CI docs
# .github/workflows/release.yml
name: release

on:
  push:
    tags:
      - 'v*.*.*'

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Note:** No `HOMEBREW_TAP_TOKEN` needed in Phase 4. The `GITHUB_TOKEN` is sufficient for creating releases on the same repo. The tap token will be added in Phase 7.

### Anti-Patterns to Avoid

- **Copying wakafetch code verbatim:** The decision is "clean rewrite with attribution." Copy-paste would carry the `github.com/sahaj-b/wakafetch` import paths into the new module. Always use `github.com/b00y0h/wakadash/internal/...` paths.
- **Implementing one-shot mode:** The CONTEXT.md decision is "no one-shot mode." Do not port wakafetch's `--range`, `--daily`, `--heatmap`, `--full`, `--json` flags. Wakadash is TUI-only.
- **Including bubbletea in Phase 4:** Phase 4 is a stub — no TUI yet. Adding bubbletea before the Phase 5 implementation is premature and creates partially-used dependencies.
- **Root `main.go`:** Putting main.go at the root works but makes `go build ./cmd/wakadash` the natural build command. Using root is inconsistent with the Phase 2 research pattern and makes Phase 5 package addition awkward.
- **Using `archives.format` (singular):** GoReleaser v2 uses `archives.formats` (array). See State of the Art section.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Multi-platform compilation | GOOS/GOARCH shell loops | GoReleaser builds section | Handles matrix builds, ldflags injection, parallel execution |
| SHA256 checksums | shasum commands | GoReleaser checksum section | Auto-generates checksums.txt for all artifacts |
| GitHub release creation | gh release create + manual uploads | GoReleaser + goreleaser-action | Handles uploads, release notes, retries, API rate limits |
| Changelog from git | git log parsing | GoReleaser changelog section | Filters by prefix, sorts, formats consistently |
| INI config parsing | custom parser | bufio.Scanner + strings.Cut | stdlib is sufficient; .wakatime.cfg is simple INI — no need for an INI library |

**Key insight:** The wakafetch config.go already uses `bufio.Scanner` for INI parsing with zero external dependencies. `strings.Cut` (Go 1.18+) replaces `strings.SplitN(line, "=", 2)` cleanly. No external INI library needed.

## Common Pitfalls

### Pitfall 1: Module Path in ldflags

**What goes wrong:** `go build` succeeds locally but GoReleaser fails to inject version — binary reports `version=dev` even after release.
**Why it happens:** When `main.go` is in `cmd/wakadash/`, the Go package is still `package main`. The ldflag `-X main.version=...` is correct. If someone mistakenly writes `-X github.com/b00y0h/wakadash/cmd/wakadash.version=...` it will not work — that's the wrong path for a `var` in `package main`.
**How to avoid:** Use `-X main.version={{.Version}}` exactly. Verify with `goreleaser release --snapshot --clean` and check the binary output.
**Warning signs:** `wakadash --version` outputs `dev` instead of a version number after a tagged release.

### Pitfall 2: Shallow Clone Breaks Changelog

**What goes wrong:** GoReleaser changelog is empty or shows only one commit.
**Why it happens:** GitHub Actions defaults to `fetch-depth: 1`. GoReleaser needs full history.
**How to avoid:** Always include `fetch-depth: 0` in the checkout step.
**Warning signs:** Empty changelog in GitHub release, GoReleaser warnings about tags.

### Pitfall 3: Wakapi URL Normalization Regression

**What goes wrong:** Users with Wakapi self-hosted instances get 404 errors.
**Why it happens:** Wakapi uses `/api/compat/wakatime` path prefix. The wakafetch config.go has special-case handling for `https://wakapi.dev/api` → `https://wakapi.dev/api/compat/wakatime`. This normalization must be preserved in the clean rewrite.
**How to avoid:** Port the normalization logic from `wakafetch/config.go` line 59-61 explicitly.
**Warning signs:** HTTP 404 errors for users with `api_url = https://wakapi.dev/api` in their config.

### Pitfall 4: Archives Format (Singular vs Plural)

**What goes wrong:** GoReleaser v2 rejects `.goreleaser.yaml` with a validation error: `format` field is deprecated.
**Why it happens:** GoReleaser v2.6 changed `archives[].format` (string) to `archives[].formats` (array).
**How to avoid:** Use `formats: [tar.gz]` (array syntax) not `format: tar.gz` (string).
**Warning signs:** `goreleaser check` reports validation errors.

### Pitfall 5: Attribution in LICENSE and README

**What goes wrong:** PR review or legal issues around code provenance.
**Why it happens:** wakadash is a clean rewrite of wakafetch patterns. MIT license requires attribution.
**How to avoid:** The LICENSE file should be a new MIT license with `b00y0h`. The README should credit wakafetch: "Inspired by [wakafetch](https://github.com/sahaj-b/wakafetch) by sahaj-b." The code comments on `internal/api/client.go` and `internal/config/config.go` should say "Attribution: Rewritten from github.com/sahaj-b/wakafetch (MIT License)."
**Warning signs:** No attribution anywhere in repo; identical code to wakafetch without acknowledgment.

### Pitfall 6: Go Version Mismatch

**What goes wrong:** `go.mod` declares `go 1.24.3` but CI uses `go-version: stable` which might be 1.24.x or different.
**Why it happens:** wakafetch uses `go 1.24.3`. The new module should match or use a minimum version.
**How to avoid:** Use `go-version: stable` in CI (gets current stable Go, which will be >= 1.24). Set `go.mod` minimum to match what features are used. `strings.Cut` requires Go 1.18+. Generics (used in `fetchJSON[T]`) require Go 1.18+.
**Warning signs:** CI fails with "requires go >= X.Y.Z" or build failures on Go version mismatch.

### Pitfall 7: Missing `dist/` in .gitignore

**What goes wrong:** GoReleaser build artifacts committed to git.
**Why it happens:** GoReleaser creates `dist/` during local testing. Without `.gitignore`, these get staged.
**How to avoid:** Include `dist/` in `.gitignore` before first `goreleaser release --snapshot --clean` run.
**Warning signs:** `git status` shows `dist/` as untracked after local GoReleaser run.

## Code Examples

Verified patterns from official sources and existing codebase:

### Repository Creation with gh CLI

```bash
# Source: gh repo create --help (gh v2.83.2 verified locally)
gh repo create b00y0h/wakadash \
  --public \
  --description "A live terminal dashboard for WakaTime coding stats" \
  --disable-wiki \
  --disable-projects

# Then clone and initialize
gh repo clone b00y0h/wakadash
cd wakadash
go mod init github.com/b00y0h/wakadash
```

### Go Module Initialization

```bash
# Source: go help mod init
go mod init github.com/b00y0h/wakadash
# Creates go.mod with: module github.com/b00y0h/wakadash / go X.Y
```

### Local GoReleaser Testing (no upload)

```bash
# Source: Phase 2 research, GoReleaser docs
# Validate config
goreleaser check

# Full dry run (builds all platforms, no GitHub release created)
goreleaser release --snapshot --clean

# Verify version injection
./dist/wakadash_linux_amd64_v1/wakadash --version
```

### Tag and Trigger Release

```bash
# Source: GoReleaser quick start pattern
git tag -a v0.1.0 -m "Initial release"
git push origin v0.1.0
# GitHub Actions triggers, GoReleaser creates release with 4 platform binaries
```

### Verify Release Artifacts (success criteria)

After the first tag push, the GitHub release should contain:
```
wakadash_0.1.0_darwin_amd64.tar.gz
wakadash_0.1.0_darwin_arm64.tar.gz
wakadash_0.1.0_linux_amd64.tar.gz
wakadash_0.1.0_linux_arm64.tar.gz
wakadash_0.1.0_checksums.txt
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `archives.format` (string) | `archives.formats` (array) | GoReleaser v2.6 (2024) | Must use array syntax in .goreleaser.yaml |
| `brews.github` | `brews.repository` | GoReleaser v2 | Old key deprecated |
| Root `main.go` | `cmd/<name>/main.go` | Go convention since ~2017 | Enables multiple binaries and cleaner package layout |
| `strings.SplitN(line, "=", 2)` | `strings.Cut(line, "=")` | Go 1.18 (2022) | Cleaner, idiomatic INI line parsing |
| Classic PATs | Fine-grained PATs | GitHub 2023+ | Fine-grained PATs scope to specific repos; needed when Phase 7 adds tap publishing |

**Deprecated/outdated:**
- GoReleaser `brews.tap` key: Use `brews.repository` with `owner`/`name`/`token`
- `--skip-publish`: Use `--skip=publish`
- `archives.format` (singular string): Use `formats` (array)

## Open Questions

1. **What should `wakadash --help` display in the stub?**
   - What we know: The stub needs `--version` and `--help` to satisfy the success criteria. "Running `wakadash --help` shows usage" is explicit.
   - What's unclear: Should the stub also accept (and error on) a flag for the future refresh interval? Or is pure stub with just `--version` sufficient?
   - Recommendation: Keep stub minimal — `--version` only, with `--help` auto-generated by `flag.Usage`. The help text should describe the eventual behavior ("live terminal dashboard") not stub behavior.

2. **Does `go.mod` need `go 1.24.3` or a lower minimum?**
   - What we know: wakafetch uses `go 1.24.3`. The clean rewrite uses `strings.Cut` (Go 1.18) and generics (Go 1.18). No features above 1.18 are strictly required for Phase 4.
   - What's unclear: Whether future bubbletea/lipgloss versions (Phase 5) require a higher minimum.
   - Recommendation: Use `go 1.21` as the minimum (it's the oldest LTS-like version still receiving security updates as of 2026). Bump to 1.24 when Phase 5 dependencies require it. This makes the module more broadly compatible without sacrificing any needed features.

3. **Should the stub attempt to load config and exit gracefully if missing?**
   - What we know: The success criterion is `wakadash --help` shows usage. Config loading failure on launch would make `--help` work only if config exists, which is a bad UX.
   - What's unclear: Whether config loading happens before or after flag parsing.
   - Recommendation: Parse flags first. If `--help` or `--version` is requested, handle it before any config loading. If neither is set, attempt config load and show a helpful error if missing (e.g., "~/.wakatime.cfg not found — see https://wakatime.com/help").

4. **Test workflow trigger: should CI also run `go build` on PRs?**
   - What we know: The success criteria mention tag-triggered releases. A separate CI workflow for PRs is not required but is good practice.
   - What's unclear: Whether Phase 4 should include a `build.yml` CI workflow for push/PR verification.
   - Recommendation: Add a simple `build.yml` that runs `go build ./...` and `go vet ./...` on every push. This is 10 extra lines and catches regressions before tags.

## Sources

### Primary (HIGH confidence)
- `/workspace/wakafetch/api.go` — Read directly; source patterns for API client rewrite
- `/workspace/wakafetch/config.go` — Read directly; source patterns for config reader rewrite
- `/workspace/wakafetch/types/types.go` — Read directly; API response types to port
- `/workspace/wakafetch/flags.go` — Read directly; flag patterns (simplified for Phase 4 stub)
- `/workspace/wakafetch/go.mod` — Read directly; `module github.com/sahaj-b/wakafetch`, `go 1.24.3`
- `/workspace/.planning/phases/02-release-automation/02-RESEARCH.md` — GoReleaser patterns verified in Phase 2
- `/workspace/.planning/research/ARCHITECTURE.md` — Architecture patterns and package structure decisions
- `/workspace/.planning/research/STACK.md` — Technology stack decisions for wakadash
- `gh repo create --help` — gh v2.83.2, verified locally

### Secondary (MEDIUM confidence)
- GoReleaser official docs (https://goreleaser.com) — `archives.formats` array syntax, `version: 2` requirement
- Go stdlib docs — `strings.Cut` (Go 1.18+), `bufio.Scanner` INI parsing pattern

### Tertiary (LOW confidence)
- None — all findings are from direct code reading or established Phase 2 research

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — directly established in Phase 2 research; GoReleaser patterns verified
- Architecture (cmd/ layout): HIGH — Go convention, consistent with Phase 2 research structure
- Code porting patterns: HIGH — source code read directly from /workspace/wakafetch
- Pitfalls: HIGH — pitfalls 1-4 from Phase 2 research (verified); pitfalls 5-7 from direct code reading and CONTEXT.md decisions
- Open questions: MEDIUM — recommendations are well-reasoned but benefit from user confirmation on stub behavior

**Research date:** 2026-02-19
**Valid until:** 2026-05-19 (GoReleaser v2 is stable; Go conventions change slowly)
