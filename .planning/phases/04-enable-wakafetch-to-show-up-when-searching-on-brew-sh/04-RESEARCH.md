# Phase 4: Enable wakafetch to show up when searching on brew.sh - Research

**Researched:** 2026-02-17
**Domain:** Homebrew package discoverability, homebrew-core formula submission
**Confidence:** HIGH

## Summary

The core question for this phase is: "What does it take to get wakafetch to appear when someone searches on brew.sh?" The answer is unambiguous: **formulae.brew.sh (brew.sh's search) indexes only homebrew-core and homebrew-cask**, not personal taps. Personal taps like `b00y0h/homebrew-wakafetch` are invisible to brew.sh search.

To appear on brew.sh search, wakafetch must be submitted to the official `Homebrew/homebrew-core` repository as a **formula** (not a cask). This is a significant process: write a Ruby formula file that builds wakafetch from source, submit a PR to homebrew-core, and pass the maintainers' review. This is entirely feasible for wakafetch because:

1. The project uses only the Go standard library — zero external dependencies
2. The upstream sahaj-b/wakafetch has 91 stars (exceeds the 75-star threshold)
3. The license is MIT (acceptable under Debian Free Software Guidelines)
4. GoReleaser already produces tagged releases with source tarballs

However, a critical complication exists: the project currently publishes a **cask** (pre-built binaries via GoReleaser's `homebrew_casks` section), but Homebrew policy requires open-source CLI tools to be submitted as **formulae** (built from source). Additionally, the formula should point to the upstream `sahaj-b/wakafetch`, not the `b00y0h` fork, since Homebrew does not accept fork-based formulae unless the fork is the officially designated successor. The upstream has no releases — that must be addressed first.

**Primary recommendation:** Submit `wakafetch` to `Homebrew/homebrew-core` as a formula that builds from source, pointing to the upstream `sahaj-b/wakafetch` repository. This requires coordinating with the upstream maintainer to cut a proper release, then writing and submitting a formula PR.

## Standard Stack

### Core

| Component | Version/Detail | Purpose | Why Standard |
|-----------|---------------|---------|--------------|
| Homebrew Ruby formula | `.rb` file in homebrew-core | Package definition | Required format for homebrew-core inclusion |
| `brew create` | Current Homebrew | Scaffolds formula template | Homebrew's official scaffolding tool |
| `std_go_args` | Homebrew built-in | Go build helper | Standard approach for all Go formulae in core |
| `brew audit --strict` | Current Homebrew | Validates formula before PR | Required pre-submission check |
| `brew test` | Current Homebrew | Tests formula install | Required pre-submission validation |

### Supporting

| Component | Version/Detail | Purpose | When to Use |
|-----------|---------------|---------|-------------|
| `brew bump-formula-pr` | Current Homebrew | Automates future version updates | After initial acceptance, for version bumps |
| GoReleaser source archives | Already configured | Provides tarball for formula URL | Already working from Phase 3 |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Submitting to homebrew-core | Keep personal tap | Tap stays invisible on brew.sh — does not achieve phase goal |
| Formula (builds from source) | Cask (pre-built binary) | Casks for open-source CLI tools are explicitly rejected by Homebrew |
| Upstream sahaj-b/wakafetch | b00y0h/wakafetch fork | Forks rejected unless designated official successor |

## Architecture Patterns

### Homebrew Formula Structure for Go CLI (No External Dependencies)

```ruby
# Source: https://formulae.brew.sh/formula/wakatime-cli (reference pattern)
class Wakafetch < Formula
  desc "Terminal dashboard for WakaTime/Wakapi coding activity"
  homepage "https://github.com/sahaj-b/wakafetch"
  url "https://github.com/sahaj-b/wakafetch/archive/refs/tags/v0.X.Y.tar.gz"
  sha256 "..."
  license "MIT"

  bottle do
    sha256 cellar: :any_skip_relocation, arm64_sequoia: "..."
    sha256 cellar: :any_skip_relocation, arm64_sonoma:  "..."
    sha256 cellar: :any_skip_relocation, sonoma:        "..."
    sha256 cellar: :any_skip_relocation, ventura:       "..."
    sha256 cellar: :any_skip_relocation, x86_64_linux:  "..."
  end

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X main.version=#{version}
    ]
    system "go", "build", *std_go_args(ldflags:)
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/wakafetch --version 2>&1")
  end
end
```

### How brew.sh Search Works

- `formulae.brew.sh` is regenerated from `Homebrew/homebrew-core` and `Homebrew/homebrew-cask` repositories automatically
- It does NOT index third-party taps
- Once a formula is merged to homebrew-core, it appears on brew.sh search within hours
- `brew search wakafetch` (local CLI) only searches taps the user has already added

### PR Submission Flow

```
1. Fork Homebrew/homebrew-core on GitHub
2. Run: brew create https://github.com/sahaj-b/wakafetch/archive/refs/tags/vX.Y.Z.tar.gz
3. Edit the generated .rb file
4. Run: brew install --build-from-source wakafetch
5. Run: brew test wakafetch
6. Run: brew audit --strict --new-formula wakafetch
7. Submit PR to Homebrew/homebrew-core
8. Await review (typically 2-7 days for new formulae)
```

### Anti-Patterns to Avoid

- **Submitting a cask for an open-source CLI tool**: Homebrew explicitly rejects this — "submit it first to homebrew/core as a formula"
- **Pointing formula at the fork (b00y0h/wakafetch)**: Homebrew rejects forks unless officially designated successor
- **Using `go get` in the build step**: homebrew-core requires offline-capable builds; however, with zero external dependencies this is a non-issue
- **Submitting without a proper upstream tagged release**: Formula requires a stable `url` pointing to a versioned tarball

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Bottle generation | Manual binary builds | Homebrew CI (BrewTestBot) | Automatically generates bottles for all platforms on PR merge |
| Formula template | Write from scratch | `brew create <tarball-url>` | Scaffolds correct Ruby structure |
| Formula validation | Manual review | `brew audit --strict --new-formula` | Catches >50 common issues automatically |
| Version bump automation | Manual PR | `brew bump-formula-pr --version X.Y.Z` | Handles checksum, commit, push in one command |

**Key insight:** Homebrew's toolchain handles nearly all the boilerplate — the main work is writing the install block and test block correctly.

## Common Pitfalls

### Pitfall 1: Submitting the Fork Instead of Upstream

**What goes wrong:** PR gets rejected with "We don't accept forks unless they're the official successor"
**Why it happens:** The project is a fork of sahaj-b/wakafetch but the formula would need to reference sahaj-b's repo
**How to avoid:** The formula should point to `sahaj-b/wakafetch` as the canonical upstream. If the goal is to distribute b00y0h's version specifically, that changes the strategy entirely.
**Warning signs:** PR description mentions "fork of" without also showing official succession

### Pitfall 2: No Upstream Release Tag

**What goes wrong:** Formula cannot be written without a tagged release + tarball URL
**Why it happens:** Upstream `sahaj-b/wakafetch` currently has zero releases on GitHub
**How to avoid:** Either (a) work with upstream to cut a release, or (b) if b00y0h's fork IS the formula target (with appropriate justification), b00y0h already has releases
**Warning signs:** `brew create` fails because there's no stable URL to point to

### Pitfall 3: Cask vs Formula Confusion

**What goes wrong:** Phase 3's GoReleaser config publishes to personal tap as a `homebrew_casks` block. If you try to submit this cask to `homebrew-cask`, it will be rejected because wakafetch is an open-source CLI tool.
**Why it happens:** GoReleaser's `homebrew_casks` section creates Homebrew casks, but Homebrew's policy is that open-source CLI tools belong in homebrew-core as formulae
**How to avoid:** Submit a **formula** (Ruby class using `go build`) to homebrew-core, not a cask
**Warning signs:** Any attempt to submit to `Homebrew/homebrew-cask` repo instead of `Homebrew/homebrew-core`

### Pitfall 4: Notability Threshold

**What goes wrong:** PR gets rejected for not meeting the popularity bar
**Why it happens:** Homebrew requires >=30 forks, >=30 watchers, or >=75 stars for niche software
**Current status:**
- `sahaj-b/wakafetch`: 91 stars (PASSES threshold of 75)
- `b00y0h/wakafetch`: 0 stars (FAILS threshold)
**How to avoid:** Submit formula pointing to `sahaj-b/wakafetch` which already meets the bar

### Pitfall 5: GoReleaser Module Path Mismatch

**What goes wrong:** Build fails because `go.mod` says `module github.com/sahaj-b/wakafetch` but b00y0h's fork uses same module path
**Why it happens:** The fork didn't update go.mod module path
**How to avoid:** Formula points to sahaj-b's repo where the module path matches. If using b00y0h's fork, the module path may need updating.
**Warning signs:** `go build` errors about module path mismatches

### Pitfall 6: Version Flag Injection

**What goes wrong:** `brew test wakafetch` fails because `--version` doesn't output expected string
**Why it happens:** The current GoReleaser config injects version via `-X main.version={{.Version}}` but formula needs to replicate this
**How to avoid:** Formula install block must include the same ldflags: `-X main.version=#{version}`
**Warning signs:** `wakafetch --version` outputs empty or "dev" instead of the version number

## Code Examples

### Reference: wakatime-cli Formula (Similar Go CLI in homebrew-core)

```ruby
# Source: https://github.com/Homebrew/homebrew-core/blob/master/Formula/w/wakatime-cli.rb
def install
  arch = Hardware::CPU.intel? ? "amd64" : Hardware::CPU.arch.to_s
  ldflags = %W[
    -s -w
    -X github.com/wakatime/wakatime-cli/pkg/version.Arch=#{arch}
    -X github.com/wakatime/wakatime-cli/pkg/version.BuildDate=#{time.iso8601}
    -X github.com/wakatime/wakatime-cli/pkg/version.Commit=#{Utils.git_head(length: 7)}
    -X github.com/wakatime/wakatime-cli/pkg/version.OS=#{OS.kernel_name.downcase}
    -X github.com/wakatime/wakatime-cli/pkg/version.Version=v#{version}
  ].join(" ")
  system "go", "build", *std_go_args(ldflags:)
end
```

### Wakafetch-specific formula install block

```ruby
# Based on go.mod: module github.com/sahaj-b/wakafetch
# GoReleaser injects: -X main.version={{.Version}}
def install
  ldflags = %W[
    -s -w
    -X main.version=#{version}
  ]
  system "go", "build", *std_go_args(ldflags:)
end

test do
  # Test that binary runs; actual API calls require credentials
  assert_match version.to_s, shell_output("#{bin}/wakafetch --version 2>&1", 0)
end
```

### Checking formula before PR submission

```bash
# Install and test locally
brew install --build-from-source Formula/w/wakafetch.rb
brew test wakafetch
brew audit --strict --new-formula wakafetch

# Check for common issues
brew style wakafetch
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| GoReleaser `brews:` section | GoReleaser `homebrew_casks:` section | GoReleaser v2.10 (2024) | Pre-built binaries now correctly go as casks |
| formulae only on macOS | formulae + casks on macOS AND Linux | Homebrew/brew#19121 | Casks now work on Linux too |
| Self-submit any formula | Must meet notability threshold | Enforced since ~2020 | New formulae need 75+ stars or community proof |

**Deprecated/outdated:**
- GoReleaser `brews:` section: Replaced by `homebrew_casks:` for pre-built binaries. Disabled as of GoReleaser 2.x. If anyone reads old tutorials using `brews:`, that approach is wrong now.

## Open Questions

1. **Who does the formula reference — sahaj-b or b00y0h?**
   - What we know: Homebrew rejects fork formulae. The upstream (sahaj-b) has the stars. The fork (b00y0h) has the releases and the GoReleaser setup.
   - What's unclear: Does the project owner want to push this upstream, or does b00y0h want their fork to be the "official" distribution even if it means a harder path?
   - Recommendation: Submit formula pointing to `sahaj-b/wakafetch`. This requires the upstream to cut a tagged release. **Alternatively**, if b00y0h wants to own the homebrew-core submission, they need to establish the fork as the official successor (difficult) or contribute the release back upstream.

2. **Does sahaj-b/wakafetch have a `--version` flag?**
   - What we know: GoReleaser injects `-X main.version={{.Version}}` into b00y0h's releases
   - What's unclear: Whether the test block can rely on `--version` working, and what flag name is used
   - Recommendation: Check `wakafetch --help` or source code before finalizing the test block

3. **What is the formula name — `wakafetch` or something else?**
   - What we know: No formula named `wakafetch` currently exists in homebrew-core
   - What's unclear: Whether Homebrew maintainers would prefer a different name
   - Recommendation: `wakafetch` is the natural name; verify with `brew search wakafetch` before submitting

4. **Timeline and review uncertainty**
   - What we know: Homebrew maintainers "typically respond within a couple days, up to a week"
   - What's unclear: Whether the formula will pass on first review or require multiple iterations
   - Recommendation: Plan for 1-3 weeks from PR submission to merge

## Sources

### Primary (HIGH confidence)

- https://docs.brew.sh/Acceptable-Formulae - Full acceptance criteria, fork policy, notability requirements
- https://docs.brew.sh/Acceptable-Casks - CLI tool cask policy ("submit to core first")
- https://formulae.brew.sh/formula/wakatime-cli - Reference Go CLI formula in homebrew-core
- https://github.com/Homebrew/homebrew-core/blob/master/Formula/w/wakatime-cli.rb - Concrete Ruby formula example

### Secondary (MEDIUM confidence)

- https://goreleaser.com/customization/homebrew_casks/ - GoReleaser cask vs formula documentation
- https://github.com/orgs/goreleaser/discussions/5563 - Why GoReleaser switched to casks
- https://docs.brew.sh/How-To-Open-a-Homebrew-Pull-Request - PR submission process
- https://github.com/b00y0h/wakafetch - 0 stars/forks (fails notability alone)
- https://github.com/sahaj-b/wakafetch - 91 stars, 5 forks (passes notability)

### Tertiary (LOW confidence)

- https://github.com/orgs/Homebrew/discussions/4697 - Community discussion on tap discoverability (confirms brew.sh doesn't index personal taps)

## Metadata

**Confidence breakdown:**
- formulae.brew.sh only indexes homebrew-core/cask: HIGH — confirmed by official Homebrew docs and community discussion
- Fork policy (forks rejected): HIGH — stated in official Acceptable Formulae docs
- Notability requirements (75 stars): HIGH — stated in official Acceptable Formulae docs
- wakafetch upstream stars (91): HIGH — directly observed on GitHub
- Zero external Go dependencies: HIGH — confirmed by reading go.mod (no require block)
- Formula structure/Ruby syntax: HIGH — verified against live wakatime-cli formula
- Review timeline estimate: MEDIUM — from official docs, actual experience varies
- Whether upstream will cut a release: LOW — unknown, requires coordination

**Research date:** 2026-02-17
**Valid until:** 2026-05-17 (stable domain — Homebrew policies change slowly)
