# Phase 2: Release Automation - Research

**Researched:** 2026-02-13
**Domain:** Go release automation with GoReleaser and GitHub Actions
**Confidence:** HIGH

## Summary

Phase 2 implements automated multi-platform Go binary releases using GoReleaser and GitHub Actions. The standard approach is well-established: GoReleaser handles cross-compilation, archiving, checksums, and changelog generation, while GitHub Actions provides the CI/CD infrastructure triggered by semantic version tags.

**Key technology stack:**
- GoReleaser v2 (current: v2.13.3 as of Jan 2026) for build automation
- GitHub Actions with goreleaser-action v6 for CI/CD
- Fine-grained Personal Access Token (PAT) for cross-repository Homebrew tap publishing
- Semantic versioning with `v*.*.*` tag pattern

**Primary recommendation:** Use GoReleaser's default configuration as the foundation, customizing only the builds (target platforms), archives (naming), and brews (Homebrew tap) sections. The default checksum (SHA256) and changelog (git-based) implementations work out-of-the-box without configuration.

**Critical requirements:**
1. Full git history (`fetch-depth: 0`) is mandatory for changelog generation
2. Cross-repository PAT with `repo` permissions required for Homebrew tap publishing (GITHUB_TOKEN cannot write to external repos)
3. Workflow needs `contents: write` permission for GitHub release creation

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| GoReleaser | v2 (~> v2.13) | Multi-platform build automation, archiving, checksums, changelog | Industry standard for Go releases; handles all release tasks declaratively |
| goreleaser-action | v6 | GitHub Actions integration for GoReleaser | Official GoReleaser action; handles setup and execution |
| actions/checkout | v4 | Git repository checkout in CI | Standard GitHub Actions checkout with full history support |
| actions/setup-go | v5 | Go toolchain setup in CI | Official Go team action for version management |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| GitHub fine-grained PAT | n/a | Cross-repository authentication | Required for Homebrew tap publishing (separate repo) |
| Semantic versioning | n/a | Version tagging convention | Use `v*.*.*` pattern for release tags |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| GoReleaser | Manual goreleaser build + gh CLI | GoReleaser handles edge cases, naming conventions, checksums automatically |
| goreleaser-action | Install GoReleaser manually in workflow | Action handles caching, version management, proper setup |
| GitHub Actions | GitLab CI, CircleCI | GitHub Actions integrates natively with GitHub releases and secrets |

**Installation:**
GoReleaser is installed via goreleaser-action in CI. For local development:
```bash
# macOS
brew install goreleaser/tap/goreleaser

# Linux
curl -sfL https://goreleaser.com/static/run | bash
```

## Architecture Patterns

### Recommended Project Structure
```
wakafetch/
├── .github/
│   └── workflows/
│       └── release.yml       # Release workflow triggered by tags
├── .goreleaser.yaml          # GoReleaser configuration
├── cmd/
│   └── wakafetch/
│       └── main.go           # Application entry point
├── internal/                 # Private packages
├── go.mod
└── go.sum
```

**Why this structure:**
- `cmd/` directory is Go convention for applications with entry points
- `.goreleaser.yaml` at root is standard GoReleaser convention
- GitHub Actions workflows in `.github/workflows/` is platform standard

### Pattern 1: Tag-Triggered Release Workflow

**What:** GitHub Actions workflow that triggers on semantic version tags and orchestrates the full release process.

**When to use:** Always - this is the standard pattern for automated Go releases.

**Example:**
```yaml
# Source: https://goreleaser.com/ci/actions/
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
          HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
```

**Critical elements:**
- `fetch-depth: 0` - GoReleaser requires full history for changelog
- `contents: write` - Permission to create GitHub releases
- `HOMEBREW_TAP_TOKEN` - Cross-repo PAT for tap publishing (not GITHUB_TOKEN)

### Pattern 2: Multi-Platform Build Configuration

**What:** GoReleaser builds section defining target platforms and architectures.

**When to use:** Always - specify exactly which platforms to support.

**Example:**
```yaml
# Source: https://goreleaser.com/customization/builds/go/
builds:
  - id: wakafetch
    main: ./cmd/wakafetch
    binary: wakafetch
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
```

**Key decisions:**
- `CGO_ENABLED=0` for static binaries (portable, no libc dependency)
- `ldflags -s -w` strips debug info (smaller binaries)
- Version injection via ldflags for `--version` flag

### Pattern 3: Archive Naming Convention

**What:** Consistent tar.gz naming with project, version, OS, and architecture.

**When to use:** Always - Homebrew and users expect this format.

**Example:**
```yaml
# Source: https://goreleaser.com/customization/archive/
archives:
  - id: wakafetch
    format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    files:
      - README.md
      - LICENSE
```

**Default behavior:**
- README and LICENSE automatically included
- Format is tar.gz by default (Windows typically gets zip via format_overrides)
- Name template follows Homebrew conventions

### Pattern 4: Homebrew Tap Publishing

**What:** GoReleaser automatically creates/updates Homebrew formula in separate tap repository.

**When to use:** For CLI tools targeting macOS/Linux users.

**Example:**
```yaml
# Source: https://goreleaser.com/customization/homebrew/
brews:
  - name: wakafetch
    repository:
      owner: username
      name: homebrew-wakafetch
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    directory: Formula
    homepage: "https://github.com/username/wakafetch"
    description: "WakaTime activity fetcher CLI"
    license: "MIT"
    test: |
      system "#{bin}/wakafetch --version"
    install: |
      bin.install "wakafetch"
```

**Critical notes:**
- Token MUST be cross-repo PAT, not GITHUB_TOKEN
- Formula auto-generated and committed to tap repo
- Tap repo must exist beforehand (goreleaser doesn't create it)

### Anti-Patterns to Avoid

- **Using GITHUB_TOKEN for Homebrew tap:** GITHUB_TOKEN is scoped only to the workflow's repository and cannot write to external tap repositories. Always use a dedicated PAT.

- **Shallow git clones:** `fetch-depth: 1` breaks changelog generation. GoReleaser needs full history to compute changes since last tag.

- **Hardcoding versions in ldflags:** Use GoReleaser templates (`{{.Version}}`, `{{.Commit}}`) instead of hardcoded values.

- **Testing releases with real tags:** Use `goreleaser release --snapshot --clean` locally to test configuration without creating releases.

- **Committing dist/ directory:** The `dist/` directory contains build artifacts and should be gitignored. GoReleaser creates it during releases.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Multi-platform compilation | Custom build scripts with GOOS/GOARCH loops | GoReleaser builds section | Handles matrix builds, naming conventions, ldflags injection, and parallel builds automatically |
| Archive creation | tar/zip commands in scripts | GoReleaser archives section | Manages consistent naming, file inclusion, format overrides per OS |
| SHA256 checksum generation | shasum commands and file writing | GoReleaser checksum section | Auto-generates checksums file for all artifacts, works out-of-the-box |
| Changelog from git commits | git log parsing scripts | GoReleaser changelog section | Filters commits, groups by type, formats consistently, handles missing tags |
| GitHub Release creation | gh CLI with manual artifact uploads | GoReleaser release + GitHub Actions | Handles uploads, release notes, checksums, retries, and API rate limits |
| Homebrew formula generation | Manual formula writing | GoReleaser brews section | Generates formula from templates, handles SHA256 calculation, commits to tap repo |

**Key insight:** GoReleaser handles dozens of edge cases (platform-specific naming, checksum verification, changelog filtering, concurrent uploads) that are tedious and error-prone to implement manually. A 50-line `.goreleaser.yaml` replaces hundreds of lines of bash scripting.

## Common Pitfalls

### Pitfall 1: Shallow Git Clone Breaks Changelog

**What goes wrong:** Changelog contains only the latest commit or is empty, even though multiple commits exist since last tag.

**Why it happens:** GitHub Actions defaults to `fetch-depth: 1` (shallow clone) for performance. GoReleaser's changelog requires full history to compute commits since previous tag.

**How to avoid:**
```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0  # CRITICAL: Full history required
```

**Warning signs:**
- Changelog shows "No history" or single commit
- GoReleaser warnings about missing tags
- Previous releases not visible in git log

### Pitfall 2: Cross-Repository Authentication Failure

**What goes wrong:** Workflow succeeds at building and creating GitHub release, but fails when publishing Homebrew formula with "403 Forbidden" or "Resource not accessible by integration."

**Why it happens:** GITHUB_TOKEN is scoped only to the repository containing the workflow. Homebrew tap is a separate repository requiring its own authentication.

**How to avoid:**
1. Create fine-grained PAT with `Contents: Read and write` for tap repository
2. Store PAT as repository secret (e.g., `HOMEBREW_TAP_TOKEN`)
3. Reference in GoReleaser config: `token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"`
4. Pass to workflow: `env: HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}`

**Warning signs:**
- Error: "Resource not accessible by integration"
- Error: "refusing to allow a Personal Access Token to create or update workflow"
- Homebrew tap formula not updated after successful release

### Pitfall 3: Missing Workflow Permissions

**What goes wrong:** Workflow fails with "Resource not accessible by integration" when trying to create GitHub release, even though GITHUB_TOKEN is provided.

**Why it happens:** GitHub Actions restricts GITHUB_TOKEN permissions by default. Creating releases requires explicit `contents: write` permission.

**How to avoid:**
```yaml
permissions:
  contents: write  # Required for creating GitHub releases
```

**Warning signs:**
- Error when uploading artifacts to release
- Error: "Resource not accessible by integration"
- Build succeeds but release creation fails

### Pitfall 4: Incorrect Archive Naming Causes Conflicts

**What goes wrong:** Archives overwrite each other, or Homebrew can't find the correct archive for the platform.

**Why it happens:** Archive name template doesn't include OS/Arch, causing all platforms to generate the same filename.

**How to avoid:**
```yaml
archives:
  - name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    # NOT: "{{ .ProjectName }}_{{ .Version }}"
```

**Warning signs:**
- Only one archive appears in release (should be 4: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64)
- Homebrew formula SHA256 mismatch errors
- Users download wrong architecture

### Pitfall 5: Testing with Real Tags

**What goes wrong:** Creating test tags pollutes release history, and failed releases leave incomplete GitHub releases.

**Why it happens:** Running `goreleaser release` locally or pushing test tags to try the workflow.

**How to avoid:**
```bash
# Local testing - no tags, no uploads
goreleaser release --snapshot --clean

# Verify config before tagging
goreleaser check
```

**Warning signs:**
- Git tags like v0.0.1-test, v0.0.0, test-release
- Incomplete GitHub releases marked as "draft"
- Confusion about which release is "real"

### Pitfall 6: Sensitive Token Exposure

**What goes wrong:** Personal Access Token appears in logs, commit history, or public repositories.

**Why it happens:** Hardcoding token in `.goreleaser.yaml` or accidentally logging secrets.

**How to avoid:**
- Always use environment variables: `token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"`
- Never commit tokens to git
- Store tokens in GitHub Secrets, not workflow files
- GitHub auto-masks known secrets in logs, but be cautious with derived values

**Warning signs:**
- Token visible in `.goreleaser.yaml`
- Token in git history (requires token rotation)
- Security alert from GitHub about exposed token

## Code Examples

Verified patterns from official sources:

### Complete .goreleaser.yaml for Multi-Platform CLI

```yaml
# Source: https://goreleaser.com/quick-start/
version: 2

builds:
  - id: wakafetch
    main: ./cmd/wakafetch
    binary: wakafetch
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
  - id: wakafetch
    format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    files:
      - README.md
      - LICENSE

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

brews:
  - name: wakafetch
    repository:
      owner: username
      name: homebrew-wakafetch
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    directory: Formula
    homepage: "https://github.com/username/wakafetch"
    description: "WakaTime activity fetcher CLI"
    license: "MIT"
    test: |
      system "#{bin}/wakafetch --version"
    install: |
      bin.install "wakafetch"
```

### Version Information in Go

```go
// Source: https://goreleaser.com/cookbooks/using-main.version/
package main

import "fmt"

var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    fmt.Printf("wakafetch %s, commit %s, built at %s\n", version, commit, date)
}
```

GoReleaser injects values via ldflags, replacing the default "dev" with actual version from git tag.

### Local Testing Commands

```bash
# Source: https://goreleaser.com/quick-start/
# Validate configuration
goreleaser check

# Test full release locally (no upload, no tag required)
goreleaser release --snapshot --clean

# Build for single platform (fast iteration)
GOOS=linux GOARCH=amd64 goreleaser build --single-target

# See what would be built without building
goreleaser build --skip=build
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `brews.tap` | `brews.repository` | GoReleaser v2 | Old syntax deprecated; use repository.owner/name instead |
| `archives.format` | `archives.formats` (array) | GoReleaser v2.6 | Support multiple formats per archive (e.g., tar.gz + zip for Windows) |
| `--skip-publish` | `--skip=publish` | GoReleaser v2 | New flag syntax for consistency; old flags still work but deprecated |
| Long-lived PATs | Fine-grained PATs with expiration | GitHub 2023+ | Fine-grained PATs limit scope to specific repositories and require regular rotation |
| `main.Date` with `{{.Date}}` | `{{.CommitDate}}` for reproducible builds | GoReleaser v1.14+ | Reproducible builds use commit timestamp, not build timestamp |

**Deprecated/outdated:**
- `brews.github` - Use `brews.repository` instead
- `archives.format` (singular) - Use `archives.formats` (array)
- `--skip-publish`, `--skip-docker` - Use `--skip=publish,docker`
- Classic PATs for new tokens - Use fine-grained PATs with repository scope

**Current best practices (2026):**
- Fine-grained PATs with 1-year expiration (security requirement)
- OIDC tokens for cloud deployments (eliminates long-lived secrets)
- Environment-based secret scoping (production vs staging)
- Regular secret rotation (30-90 days for critical systems)

## Open Questions

1. **Go version compatibility**
   - What we know: GoReleaser supports Go 1.18+ (for ldflags features)
   - What's unclear: Whether wakafetch has minimum Go version requirements
   - Recommendation: Use `go-version: stable` in workflow (gets latest stable Go)

2. **Binary size optimization**
   - What we know: `ldflags -s -w` strips debug info, `CGO_ENABLED=0` avoids libc
   - What's unclear: Whether UPX compression is needed (GoReleaser supports it)
   - Recommendation: Start without UPX; add only if binary size is problematic

3. **Windows support**
   - What we know: Phase requirements specify darwin and linux only
   - What's unclear: Future Windows support plans
   - Recommendation: Add `format_overrides` for Windows zip when needed

4. **Release artifact retention**
   - What we know: GitHub releases store artifacts indefinitely
   - What's unclear: Whether old release cleanup is needed
   - Recommendation: No action needed; GitHub handles storage

## Sources

### Primary (HIGH confidence)
- [GoReleaser Quick Start](https://goreleaser.com/quick-start/) - Initial setup and basic workflow
- [GoReleaser GitHub Actions Documentation](https://goreleaser.com/ci/actions/) - Official CI/CD integration guide
- [GoReleaser Archives Configuration](https://goreleaser.com/customization/archive/) - Archive naming and formatting
- [GoReleaser Checksum Configuration](https://goreleaser.com/customization/checksum/) - SHA256 checksum generation
- [GoReleaser Changelog Configuration](https://goreleaser.com/customization/changelog/) - Git-based changelog
- [GoReleaser Go Builds](https://goreleaser.com/customization/builds/go/) - Multi-platform build configuration
- [goreleaser-action GitHub Repository](https://github.com/goreleaser/goreleaser-action) - Official GitHub Action
- [GitHub Actions Automatic Token Authentication](https://docs.github.com/en/actions/security-guides/automatic-token-authentication) - GITHUB_TOKEN limitations
- [GitHub Actions Workflow Permissions](https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/controlling-permissions-for-github_token) - Permissions reference

### Secondary (MEDIUM confidence)
- [GoReleaser Homebrew Taps](https://goreleaser.com/customization/homebrew/) - Tap publishing configuration (page had redirect issues, verified via search results)
- [Best Practices for Managing Secrets in GitHub Actions](https://www.blacksmith.sh/blog/best-practices-for-managing-secrets-in-github-actions) - 2026 secrets management guidance
- [8 GitHub Actions Secrets Management Best Practices](https://www.stepsecurity.io/blog/github-actions-secrets-management-best-practices) - Security hardening
- [GoReleaser Issue #2026 - PAT Security Concerns](https://github.com/goreleaser/goreleaser/issues/2026) - Cross-repo authentication discussion
- [Using ldflags to Set Version Information](https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications) - Version injection pattern
- [How to Release to Homebrew with GoReleaser](https://billyhadlow.com/blog/how-to-release-to-homebrew/) - Complete workflow example

### Tertiary (LOW confidence, marked for validation)
- [GoReleaser Lessons Learned](https://carlosbecker.com/posts/goreleaser-lessons-learned/) - Creator's reflections (useful for context, not authoritative for current best practices)
- Community discussions on GitHub about deploy keys vs PATs (active discussions, not finalized solutions)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - GoReleaser is the de facto standard for Go releases; patterns verified from official docs
- Architecture: HIGH - Workflow structure verified from official GitHub Actions and GoReleaser documentation
- Pitfalls: HIGH - Common issues documented in official troubleshooting, GitHub issues, and multiple sources
- Security practices: MEDIUM - Rapidly evolving area; fine-grained PATs and OIDC are current (2026) but practices continue to evolve

**Research date:** 2026-02-13
**Valid until:** ~2026-04-13 (60 days - GoReleaser is stable, but GitHub Actions features and security practices evolve quarterly)

**Research notes:**
- GoReleaser v2 is current and stable; v3 not planned yet
- GitHub fine-grained PATs became GA in 2023 and are now standard
- OIDC for cloud deployments (AWS, Azure, GCP) is 2026 best practice but not needed for GitHub-to-GitHub operations
- Homebrew tap cross-repo authentication remains a known pain point; community wants deploy key support but PATs are current solution
