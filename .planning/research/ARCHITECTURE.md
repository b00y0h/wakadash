# Architecture Research: Go CLI Release Pipeline with Homebrew Tap

**Domain:** Go CLI Release Automation with Homebrew Distribution
**Researched:** 2026-02-13
**Confidence:** HIGH

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Source Repository (wakafetch)                   │
├─────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌───────────────────┐  ┌──────────────────────┐  │
│  │   Go Source  │  │  .goreleaser.yaml │  │  .github/workflows/  │  │
│  │     Code     │  │  Configuration    │  │    release.yml       │  │
│  └──────┬───────┘  └─────────┬─────────┘  └──────────┬───────────┘  │
│         │                    │                        │              │
├─────────┴────────────────────┴────────────────────────┴──────────────┤
│                           Git Tag Push (vX.Y.Z)                      │
│                                    ↓                                 │
├─────────────────────────────────────────────────────────────────────┤
│                        GitHub Actions Workflow                       │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────────────┐    │
│  │  Checkout   │→ │  Setup Go   │→ │  Run GoReleaser Action   │    │
│  │ (full hist) │  │             │  │  (distribution: OSS/Pro) │    │
│  └─────────────┘  └─────────────┘  └──────────┬───────────────┘    │
│                                                │                     │
├────────────────────────────────────────────────┴─────────────────────┤
│                         GoReleaser Pipeline                          │
├─────────────────────────────────────────────────────────────────────┤
│  Phase 1: Before Hooks                                               │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │  Execute custom commands (tests, linters, etc.)             │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                 ↓                                    │
│  Phase 2: Builds                                                     │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │  Build matrix: GOOS × GOARCH combinations                   │    │
│  │  Output to: dist/{BuildID}_{BuildTarget}/binary             │    │
│  │  Common targets: linux/amd64, darwin/amd64, darwin/arm64    │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                 ↓                                    │
│  Phase 3: Archives                                                   │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │  Package binaries: .tar.gz (Unix), .zip (Windows)           │    │
│  │  Include extras: README, LICENSE, completions, manpages     │    │
│  │  Output: dist/wakafetch_v1.0.0_linux_amd64.tar.gz          │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                 ↓                                    │
│  Phase 4: Checksums                                                  │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │  Generate SHA256 checksums for all artifacts                │    │
│  │  Output: dist/checksums.txt                                 │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                 ↓                                    │
│  Phase 5: Signing (optional)                                         │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │  Sign checksums with GPG or Cosign                          │    │
│  │  Output: checksums.txt.sig, checksums.txt.sigstore.json     │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                 ↓                                    │
│  Phase 6: GitHub Release                                             │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │  Create GitHub release with generated changelog             │    │
│  │  Upload: binaries, archives, checksums, signatures          │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                 ↓                                    │
│  Phase 7: Homebrew Cask Generation                                  │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │  Generate Ruby Cask definition from template                │    │
│  │  Include: download URLs, SHA256, installation steps         │    │
│  │  Commit to tap repo: Casks/wakafetch.rb                     │    │
│  └─────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
                                   ↓
┌─────────────────────────────────────────────────────────────────────┐
│              Tap Repository (homebrew-wakafetch)                     │
├─────────────────────────────────────────────────────────────────────┤
│  Casks/                                                              │
│  └── wakafetch.rb          # Auto-generated cask definition         │
│                                                                      │
│  README.md                 # User installation instructions          │
└─────────────────────────────────────────────────────────────────────┘
                                   ↓
                   User: brew tap b00y0h/wakafetch
                   User: brew install wakafetch
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **Source Repository** | Holds Go source code, build config, CI workflow | GitHub repository with Go modules |
| **GoReleaser Config** | Defines build matrix, archives, distributions | `.goreleaser.yaml` in project root |
| **GitHub Actions Workflow** | Triggers on tags, orchestrates release | `.github/workflows/release.yml` |
| **GoReleaser Engine** | Executes build pipeline phases sequentially | goreleaser/goreleaser-action@v6 |
| **dist/ Directory** | Temporary storage for all build artifacts | Generated locally/in CI, git-ignored |
| **GitHub Release** | Hosts binaries and archives for download | Created via GITHUB_TOKEN |
| **Tap Repository** | Homebrew package registry for your software | Separate GitHub repo: homebrew-{name} |
| **Homebrew Cask** | Ruby DSL defining how to install your CLI | Auto-generated, committed to tap repo |

## Recommended Project Structure

### Source Repository (b00y0h/wakafetch)

```
wakafetch/
├── .github/
│   └── workflows/
│       └── release.yml          # GitHub Actions workflow for releases
├── cmd/
│   └── wakafetch/
│       └── main.go              # CLI entrypoint
├── internal/                     # Internal packages
├── .goreleaser.yaml             # GoReleaser configuration
├── .gitignore                   # MUST include: dist/
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── LICENSE                      # Required for Homebrew
├── README.md                    # Documentation (referenced by cask)
└── dist/                        # Generated by GoReleaser (git-ignored)
    ├── wakafetch_darwin_amd64/
    ├── wakafetch_linux_amd64/
    ├── wakafetch_v1.0.0_linux_amd64.tar.gz
    ├── checksums.txt
    └── artifacts.json
```

### Tap Repository (b00y0h/homebrew-wakafetch)

```
homebrew-wakafetch/
├── Casks/
│   └── wakafetch.rb            # Auto-generated by GoReleaser
├── README.md                    # Installation instructions for users
└── .github/                     # Optional: CI for cask validation
    └── workflows/
        └── test.yml             # brew audit --cask wakafetch
```

### Structure Rationale

- **Separate repositories:** Homebrew convention requires tap repos with `homebrew-` prefix
- **Casks/ directory:** Homebrew expects casks here (not Formula/ which is for source builds)
- **dist/ in .gitignore:** Build artifacts should never be committed to source
- **.goreleaser.yaml in root:** GoReleaser convention and GitHub Actions default path
- **LICENSE file required:** Homebrew won't accept packages without clear licensing

## Architectural Patterns

### Pattern 1: GitHub Actions Tag-Triggered Release

**What:** Workflow triggers only on semantic version tags (v1.0.0, v2.3.1-beta, etc.)

**When to use:** Standard for all Go CLI projects using GoReleaser

**Trade-offs:**
- **Pro:** Clean separation between development and release processes
- **Pro:** Automatic versioning from git tags
- **Con:** Requires discipline in tag management (can't easily "undo" a tag push)

**Example:**
```yaml
name: release
on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write  # Required for GitHub releases

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # CRITICAL: GoReleaser needs full history

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      - uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
```

### Pattern 2: Cross-Repository Homebrew Publishing with Custom PAT

**What:** Use separate GitHub Personal Access Token for pushing to tap repository

**When to use:** When tap repo is different from source repo (almost always the case)

**Trade-offs:**
- **Pro:** GITHUB_TOKEN can't write to other repos, PAT enables cross-repo commits
- **Pro:** Fine-grained tokens can limit permissions to just the tap repo
- **Con:** Requires creating and securing an additional secret
- **Con:** PAT has broader permissions than ideal (security consideration)

**Example:**
```yaml
# .goreleaser.yaml
homebrew_casks:
  - name: wakafetch
    repository:
      owner: b00y0h
      name: homebrew-wakafetch
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    homepage: https://github.com/b00y0h/wakafetch
    description: "Wakatime stats fetcher CLI"
    license: MIT
```

### Pattern 3: Multi-Platform Build Matrix

**What:** Define GOOS/GOARCH combinations for cross-compilation

**When to use:** Always, unless targeting only a single platform

**Trade-offs:**
- **Pro:** One CI run produces binaries for all platforms
- **Pro:** Go's excellent cross-compilation makes this nearly free
- **Con:** More artifacts to upload/download increases release time
- **Con:** Testing on all target platforms requires separate infrastructure

**Example:**
```yaml
# .goreleaser.yaml
builds:
  - id: wakafetch
    binary: wakafetch
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    # Results in: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, windows/arm64
```

### Pattern 4: Archive Customization with Extras

**What:** Include additional files (README, LICENSE, completions) in release archives

**When to use:** When users need documentation/tools beyond the binary

**Trade-offs:**
- **Pro:** Self-contained archives with everything users need
- **Pro:** Shell completions improve UX significantly
- **Con:** Slightly larger download sizes
- **Con:** Requires generating completions as part of build

**Example:**
```yaml
# .goreleaser.yaml
archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    files:
      - LICENSE
      - README.md
      - completions/*
      - manpages/*
```

### Pattern 5: Homebrew Cask Over Deprecated Brews

**What:** Use `homebrew_casks` instead of deprecated `brews` section

**When to use:** Always for Go CLI tools (as of GoReleaser v2.10+)

**Trade-offs:**
- **Pro:** Correct Homebrew semantics (casks for pre-built binaries)
- **Pro:** Supported on both macOS and Linux
- **Pro:** Simpler installation (no source building)
- **Con:** Requires migration if using old `brews` config
- **Con:** May need quarantine removal hook for unsigned binaries

**Example:**
```yaml
# .goreleaser.yaml
homebrew_casks:
  - name: wakafetch
    repository:
      owner: b00y0h
      name: homebrew-wakafetch
    directory: Casks
    homepage: https://github.com/b00y0h/wakafetch
    description: "Wakatime stats fetcher CLI"
    license: MIT

    # For unsigned binaries (development phase)
    hooks:
      post:
        install: |
          if OS.mac?
            system_command "/usr/bin/xattr", args: ["-dr", "com.apple.quarantine", "#{staged_path}/wakafetch"]
          end
```

## Data Flow

### Release Trigger Flow

```
Developer → git tag v1.0.0 → git push origin v1.0.0
    ↓
GitHub detects tag push
    ↓
Triggers .github/workflows/release.yml
    ↓
GitHub Actions runner starts
```

### Build Artifact Flow

```
Source Code (main.go)
    ↓
GoReleaser Build Phase
    ↓ (for each GOOS/GOARCH)
Compiled Binaries → dist/{BuildID}_{BuildTarget}/wakafetch
    ↓
Archive Phase
    ↓ (tar.gz for Unix, zip for Windows)
Packaged Archives → dist/wakafetch_v1.0.0_{OS}_{ARCH}.tar.gz
    ↓
Checksum Phase
    ↓
SHA256 Checksums → dist/checksums.txt
    ↓
GitHub Release Upload
    ↓
GitHub Release Assets (public URLs)
```

### Homebrew Distribution Flow

```
GitHub Release Assets (URLs + SHA256s)
    ↓
GoReleaser Homebrew Cask Phase
    ↓
Generate wakafetch.rb (Ruby DSL)
    ↓ (contains download URL, SHA256, install instructions)
Commit to b00y0h/homebrew-wakafetch/Casks/wakafetch.rb
    ↓ (using HOMEBREW_TAP_TOKEN)
Push to tap repository
    ↓
User: brew tap b00y0h/wakafetch
    ↓ (Homebrew clones tap repo locally)
User: brew install wakafetch
    ↓ (Homebrew reads Casks/wakafetch.rb)
Download from GitHub Release URL
    ↓ (verify SHA256)
Extract archive
    ↓
Install binary to /usr/local/bin/ (Intel) or /opt/homebrew/bin/ (Apple Silicon)
```

### Key Data Flow Elements

1. **Version propagation:** Git tag → GoReleaser → Archive names → Cask version → User installation
2. **Integrity chain:** Binary → SHA256 → checksums.txt → Cask definition → Homebrew verification
3. **Authorization chain:** GITHUB_TOKEN (release) → HOMEBREW_TAP_TOKEN (tap commit) → Public access (user install)

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| **0-1k users** | Basic setup is sufficient: GitHub releases + Homebrew tap. Single-arch builds acceptable. |
| **1k-10k users** | Add multi-arch support (arm64), sign binaries with GPG/Cosign for trust, consider Scoop (Windows), AUR (Arch Linux). |
| **10k-100k users** | Add CDN/mirror for release artifacts, implement SBOM generation for supply chain transparency, add attestations with Sigstore, consider Docker images for containerized users. |
| **100k+ users** | Consider package managers: Snap, Flatpak, system package repos (APT, RPM), mirror artifacts to multiple CDNs, implement proper notarization for macOS (required for Gatekeeper). |

### Scaling Priorities

1. **First bottleneck: Download speed from GitHub Releases**
   - **Symptom:** Slow downloads for users far from GitHub's servers
   - **Fix:** Use GoReleaser's upload integrations (S3, GCS) as mirrors
   - **Timing:** When you start seeing users outside North America/Europe

2. **Second bottleneck: macOS Gatekeeper blocking unsigned binaries**
   - **Symptom:** macOS users get security warnings, have to bypass Gatekeeper
   - **Fix:** Code sign with Apple Developer cert + notarize via GoReleaser Pro
   - **Timing:** As soon as you have any significant macOS user base

3. **Third bottleneck: Trust and supply chain verification**
   - **Symptom:** Security-conscious users can't verify build provenance
   - **Fix:** Add Cosign signing, SBOM generation, GitHub attestations
   - **Timing:** Before pursuing enterprise users

## Anti-Patterns

### Anti-Pattern 1: Using GITHUB_TOKEN for Tap Repository Commits

**What people do:** Try to use default `GITHUB_TOKEN` to commit to tap repo

**Why it's wrong:** GitHub's automatic `GITHUB_TOKEN` is scoped only to the workflow's repository. It cannot push to other repositories, causing tap publishing to fail silently or with permission errors.

**Do this instead:** Create a Personal Access Token (classic or fine-grained) with `repo` or `contents:write` permissions for the tap repository, add it as `HOMEBREW_TAP_TOKEN` secret, and reference it in GoReleaser config:

```yaml
homebrew_casks:
  - repository:
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
```

### Anti-Pattern 2: Committing dist/ Directory to Git

**What people do:** Commit GoReleaser's output directory to version control

**Why it's wrong:** Binary artifacts bloat repository size, pollute git history, create merge conflicts, and belong in releases (not source control). GoReleaser explicitly warns about dirty repos with untracked files.

**Do this instead:** Add `dist/` to `.gitignore` immediately. Always use `--clean` flag in GoReleaser to remove old artifacts. Let GitHub Releases be the artifact storage, not git.

```gitignore
# .gitignore
dist/
```

### Anti-Pattern 3: Using Deprecated brews Section

**What people do:** Copy old GoReleaser examples using `brews:` configuration

**Why it's wrong:** Homebrew formulae are designed for source-based installations. Using them for pre-compiled binaries confuses users and breaks Homebrew conventions. This pattern is deprecated as of GoReleaser v2.10 and will be removed in v3.

**Do this instead:** Use `homebrew_casks:` which properly represents pre-built binary distribution:

```yaml
# WRONG (deprecated)
brews:
  - tap:
      owner: b00y0h
      name: homebrew-wakafetch

# RIGHT (current)
homebrew_casks:
  - repository:
      owner: b00y0h
      name: homebrew-wakafetch
    directory: Casks  # Not Formula!
```

### Anti-Pattern 4: Shallow Git Checkout in CI

**What people do:** Use default `actions/checkout` without `fetch-depth: 0`

**Why it's wrong:** GoReleaser needs full git history to generate changelogs, detect previous tags, and compute version information. Shallow checkouts (depth=1) cause changelog generation to fail or produce incomplete results.

**Do this instead:** Always specify full history in checkout step:

```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0  # REQUIRED for GoReleaser
```

### Anti-Pattern 5: Hardcoding Versions and URLs

**What people do:** Manually update version numbers in config files, hardcode download URLs

**Why it's wrong:** Creates manual work, high chance of errors, defeats the purpose of automation. Version mismatches between git tag and config cause user confusion.

**Do this instead:** Use GoReleaser's templating system:

```yaml
# Uses git tag automatically
version: "{{ .Version }}"

# Auto-generated from GitHub release
url: "https://github.com/b00y0h/wakafetch/releases/download/{{ .Tag }}/{{ .ArtifactName }}"

# Dynamic SHA256 from build
sha256: "{{ .ArtifactChecksum }}"
```

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| **GitHub Releases** | Direct API via GITHUB_TOKEN | Automatic with GoReleaser, requires `contents: write` permission |
| **GitHub (Tap Repo)** | Git push via HOMEBREW_TAP_TOKEN | Requires separate PAT with `repo` or `contents:write` scope |
| **Homebrew Core** | Manual PR submission | Optional: submit to official taps after proving stability |
| **GPG Keyserver** | Cosign/GPG for signatures | Optional but recommended for trust/verification |
| **Docker Registry** | Push via GoReleaser dockers section | Optional: parallel distribution method |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| **Source Code ↔ GoReleaser** | Filesystem read (go.mod, .goreleaser.yaml) | GoReleaser reads source, doesn't modify it |
| **GoReleaser ↔ dist/ Directory** | Filesystem write | Temporary artifact storage, cleaned between runs |
| **GitHub Actions ↔ GoReleaser** | Environment variables (secrets, tokens) | Secrets passed via env, not CLI args |
| **Source Repo ↔ Tap Repo** | Git push via GitHub API | Unidirectional: source never reads tap, only writes |
| **Homebrew Client ↔ Tap Repo** | Git clone, file read | Homebrew clones tap locally, reads Cask definitions |
| **Homebrew Client ↔ GitHub Releases** | HTTPS download | Homebrew fetches binaries from release URLs in Cask |

## Build Order Dependencies

### Phase 1: Repository Setup (can be done in parallel)

```
Create source repo (b00y0h/wakafetch)
    ∥
Create tap repo (b00y0h/homebrew-wakafetch)
    ∥
Generate GitHub PAT with tap repo access
```

### Phase 2: Local Development (sequential)

```
1. Write Go CLI code
    ↓
2. Test locally (go build, go test)
    ↓
3. Create .goreleaser.yaml config
    ↓
4. Test GoReleaser locally (goreleaser release --snapshot)
    ↓
5. Verify dist/ artifacts look correct
```

### Phase 3: CI Configuration (sequential, depends on Phase 1)

```
1. Add PAT as HOMEBREW_TAP_TOKEN secret
    ↓
2. Create .github/workflows/release.yml
    ↓
3. Configure permissions (contents: write)
    ↓
4. Test with dry-run or pre-release tag
```

### Phase 4: First Release (sequential, depends on Phases 2-3)

```
1. Commit all changes to main
    ↓
2. Create and push git tag (git tag v0.1.0 && git push origin v0.1.0)
    ↓
3. Monitor GitHub Actions workflow
    ↓
4. Verify GitHub Release created
    ↓
5. Verify Cask committed to tap repo
    ↓
6. Test installation (brew tap b00y0h/wakafetch && brew install wakafetch)
```

### Critical Dependencies

- **GoReleaser config MUST exist** before first tag push (or release will fail)
- **HOMEBREW_TAP_TOKEN MUST be set** before enabling tap publishing (or tap commits fail)
- **fetch-depth: 0 MUST be set** before GoReleaser runs (or changelog fails)
- **Tap repository MUST exist** before GoReleaser tries to commit (or push fails)
- **GitHub Release MUST complete** before Homebrew cask generation (downloads need URLs)

## Sources

### Official Documentation (HIGH confidence)
- [GoReleaser Introduction](https://goreleaser.com/intro/)
- [GoReleaser GitHub Actions](https://goreleaser.com/ci/actions/)
- [GoReleaser Dist Folder](https://goreleaser.com/customization/dist/)
- [GoReleaser Homebrew Formulas (deprecated)](https://goreleaser.com/customization/homebrew_formulas/)
- [GoReleaser Homebrew Casks](https://goreleaser.com/customization/homebrew_casks/)
- [GoReleaser v2.10 Announcement](https://goreleaser.com/blog/goreleaser-v2.10/)
- [GoReleaser Checksums](https://goreleaser.com/customization/checksum/)
- [GoReleaser Archives](https://goreleaser.com/customization/archive/)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Ruby API: Formula](https://rubydoc.brew.sh/Formula.html)
- [Homebrew Ruby API: Tap](https://rubydoc.brew.sh/Tap.html)

### GitHub Repositories (HIGH confidence)
- [goreleaser/goreleaser-action](https://github.com/goreleaser/goreleaser-action)
- [goreleaser/goreleaser releases](https://github.com/goreleaser/goreleaser/releases)
- [GitHub Issue #5594: Brew packages should be casks](https://github.com/goreleaser/goreleaser/issues/5594)

### Community Resources (MEDIUM confidence)
- [Creating Your First Homebrew Tap (Kristoffer.dev)](https://kristoffer.dev/blog/guide-to-creating-your-first-homebrew-tap/)
- [From Go Code to Homebrew Tap (Applied Go)](https://appliedgo.net/whisper-cli/)
- [Creating Homebrew Formulas with GoReleaser (Bindplane)](https://bindplane.com/blog/creating-homebrew-formulas-with-goreleaser)
- [How to release to Homebrew with GoReleaser (Billy Hadlow)](https://billyhadlow.com/blog/how-to-release-to-homebrew/)
- [DeepWiki: Homebrew/brew Core Package Management](https://deepwiki.com/Homebrew/brew/3-formula-and-cask-system)

---
*Architecture research for: Go CLI Release Pipeline with Homebrew Tap*
*Researched: 2026-02-13*
