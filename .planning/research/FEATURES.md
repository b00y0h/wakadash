# Feature Landscape

**Domain:** Go CLI Release Automation with Homebrew Distribution
**Researched:** 2026-02-13
**Confidence:** HIGH

## Table Stakes

Features users expect. Missing = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Multi-platform binaries | Go CLI users run macOS, Linux, Windows | Low | GoReleaser handles via cross-compilation automatically |
| Multi-architecture builds | Apple Silicon (ARM64) + Intel (AMD64) required for macOS | Low | Must support arm64, amd64 at minimum |
| GitHub Release creation | Standard distribution method for Go tools | Low | Automated via GoReleaser + GitHub Actions |
| Homebrew tap | macOS users expect `brew install` for CLI tools | Low | GoReleaser auto-updates tap formula/cask |
| Checksums | Security-conscious users verify downloads | Low | GoReleaser generates automatically |
| Automated changelog | Users expect to see what changed | Low | Generated from git commits or conventional commits |
| Archive packaging | Users need `.tar.gz` (Unix) and `.zip` (Windows) | Low | GoReleaser handles automatically |
| Semantic versioning | Industry standard (v1.2.3 format) | Low | Tag-based releases with git tags |
| README/docs in release | Users need usage instructions | Low | Include via GoReleaser extra_files |

## Differentiators

Features that set product apart. Not expected, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Binary signatures (GPG/Cosign) | Enterprise security requirements | Medium | Verifies authenticity, prevents tampering |
| SBOM generation | Supply chain security compliance | Medium | Required for enterprise/government users |
| macOS Universal Binaries | Single binary for both Intel + ARM Macs | Medium | Better UX than separate downloads |
| AI-enhanced release notes | Polished, user-friendly changelogs | Low | GoReleaser Pro v2.6+ feature (requires API key) |
| Docker image distribution | Alternative to Homebrew for containers | Medium | Parallel distribution channel |
| Homebrew Cask (vs Formula) | Proper pre-compiled binary distribution | Low | GoReleaser v2.10+ standard (formula deprecated) |
| Cross-platform CI matrix | Parallel builds speed up releases | Low | GitHub Actions matrix strategy |
| Draft releases for review | Preview before publishing | Low | PR-based Homebrew updates, GitHub draft releases |
| Automated version bumping | Commit-based semantic versioning | Medium | Tools like go-semantic-release, svu |
| Multiple Homebrew taps | Stable vs beta channels | Medium | Separate taps for different release tracks |

## Anti-Features

Features to explicitly NOT build.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Custom download server | Maintenance burden, trust issues | Use GitHub Releases (free, trusted, CDN) |
| Manual changelog editing | Error-prone, delays releases | Auto-generate from conventional commits |
| Version numbers in code | Out of sync, manual updates | Inject via `-ldflags` at build time |
| Homebrew core submission (initially) | Slow review, high barrier | Start with personal tap, migrate later if popular |
| Building on release server | Reproducibility issues | Build in CI, release server only publishes |
| All-in-one mega binary | Complexity, size bloat | Single focused CLI, not a Swiss Army knife |
| Custom update mechanism | Reinventing wheel, security risk | Let Homebrew/package managers handle updates |
| Windows MSI/EXE installers (initially) | Complex signing, overhead | Provide .zip archives first, add installers if demand exists |
| Nightly/continuous releases | Version fatigue, unstable perception | Use semantic versioning with pre-release tags (v1.2.3-beta.1) |

## Feature Dependencies

```
GitHub Release
    └──requires──> Git tags (semantic versioning)
    └──requires──> Multi-platform binaries
    └──requires──> Checksums file

Homebrew Tap
    └──requires──> GitHub Release
    └──requires──> Archive packaging (.tar.gz)
    └──enhances──> macOS Universal Binaries (better UX)

Binary Signatures
    └──requires──> Checksums file
    └──requires──> GPG key or Cosign setup

SBOM Generation
    └──requires──> Built binaries (not source)
    └──enhances──> Binary Signatures (sign SBOM)

macOS Universal Binaries
    └──requires──> Both arm64 and amd64 builds
    └──conflicts──> Homebrew Cask single-arch requirement (use separate archives)

Docker Distribution
    └──requires──> Multi-platform binaries
    └──optional──> SBOM for image metadata

Automated Version Bumping
    └──requires──> Conventional Commits format
    └──enables──> Automated changelog generation
```

### Dependency Notes

- **Homebrew requires GitHub Release**: Tap formulas/casks download from GitHub Release assets
- **Checksums before signatures**: Sign the checksums file, not every individual binary
- **Universal binaries are optional**: Provide separate arm64/amd64 archives for broader compatibility
- **SBOM from binaries, not source**: More accurate dependency tree from `go version -m <binary>`
- **Homebrew Cask vs Formula**: Casks are now standard for pre-compiled binaries (formula deprecated in GoReleaser v2.10)

## MVP Recommendation

### Launch With (v1)

Minimum viable release automation:

- [x] **Multi-platform binaries** (macOS arm64/amd64, Linux arm64/amd64, Windows amd64) — Table stakes
- [x] **GitHub Release automation** — Standard distribution method
- [x] **Homebrew Cask in personal tap** — Primary macOS install method
- [x] **Checksums file** — Security baseline
- [x] **Basic changelog** (git log-based) — Users need to know what changed
- [x] **Archive packaging** (.tar.gz, .zip) — Standard formats
- [x] **GitHub Actions workflow** — Automate on tag push

**Rationale**: This covers the absolute minimum users expect. Without Homebrew support, macOS users won't adopt. Without multi-platform builds, Linux/Windows users excluded. Without checksums, security-conscious users won't trust it.

### Add After Validation (v1.x)

Features to add once core is working and user feedback arrives:

- [ ] **Binary signatures (Cosign)** — Add when enterprise users request it (trigger: first security audit request)
- [ ] **macOS Universal Binaries** — Add if users complain about confusion between arm64/amd64 (trigger: 3+ GitHub issues)
- [ ] **Conventional Commits + better changelog** — Improve when changelog quality complaints arise (trigger: first "what changed?" issue)
- [ ] **SBOM generation** — Add when compliance requirements surface (trigger: first enterprise deployment)
- [ ] **Docker image distribution** — Add if container users request it (trigger: 5+ requests)

**Rationale**: These improve the product but aren't blockers for initial adoption. Add based on actual user demand, not speculation.

### Future Consideration (v2+)

Features to defer until product-market fit established:

- [ ] **AI-enhanced release notes** — Requires GoReleaser Pro subscription, benefit unclear
- [ ] **Automated semantic versioning** — Adds complexity to workflow, manual tags work fine initially
- [ ] **Beta/stable tap separation** — Only needed when significant user base exists
- [ ] **Windows MSI installers** — Add if Windows adoption exceeds 20% of user base
- [ ] **Homebrew core submission** — Only after 1000+ GitHub stars and stable 1.0 release

**Rationale**: These are "nice to have" improvements that add complexity. Don't invest time until core value is proven and user base demands them.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Multi-platform binaries | HIGH | LOW | P1 |
| GitHub Release automation | HIGH | LOW | P1 |
| Homebrew Cask | HIGH | LOW | P1 |
| Checksums | HIGH | LOW | P1 |
| Basic changelog | MEDIUM | LOW | P1 |
| Archive packaging | HIGH | LOW | P1 |
| GitHub Actions CI | HIGH | LOW | P1 |
| Binary signatures | MEDIUM | MEDIUM | P2 |
| macOS Universal Binaries | MEDIUM | MEDIUM | P2 |
| Conventional Commits changelog | MEDIUM | LOW | P2 |
| SBOM generation | LOW | MEDIUM | P2 |
| Docker distribution | LOW | MEDIUM | P2 |
| AI release notes | LOW | MEDIUM | P3 |
| Automated version bumping | LOW | MEDIUM | P3 |
| Separate beta tap | LOW | LOW | P3 |
| Windows installers | LOW | HIGH | P3 |
| Homebrew core | MEDIUM | HIGH | P3 |

**Priority key:**
- P1: Must have for launch (wakafetch v0.1.0)
- P2: Should have, add when demand surfaces
- P3: Nice to have, future consideration

## Competitor Feature Analysis

Comparing against popular Go CLI tools distributed via Homebrew:

| Feature | kubectl | gh (GitHub CLI) | hugo | wakafetch Plan |
|---------|---------|-----------------|------|----------------|
| Homebrew tap | Official tap | Official tap | Official tap | Personal tap → core later |
| Multi-arch macOS | Universal binary | Universal binary | Separate binaries | Separate binaries initially |
| Checksums | ✓ SHA256 | ✓ SHA256 | ✓ SHA256 | ✓ SHA256 |
| Binary signatures | ✓ Cosign | ✓ GPG | ✗ None | ✗ Initially, add if requested |
| SBOM | ✓ Included | ✗ None | ✗ None | ✗ Initially, add if compliance needed |
| Docker images | ✓ Official | ✓ Official | ✓ Official | ✗ Defer unless requested |
| Auto-update check | ✓ Built-in | ✓ Built-in | ✗ None | ✗ Let Homebrew handle |
| Changelog quality | Generated | High quality | Generated | Start generated, improve based on feedback |

**Key insights:**
- **Universal binaries aren't universal**: Even major tools (hugo) use separate binaries
- **Security features vary**: kubectl has full SBOM/signatures, hugo has neither — depends on user base
- **Docker is common but optional**: All three have official images, but it's not a blocker
- **Auto-update is rare**: Most rely on package manager updates
- **Homebrew core takes time**: All started with taps, migrated to core after adoption

## Sources

### GoReleaser Features and Capabilities
- [GoReleaser Official Documentation](https://goreleaser.com/)
- [GoReleaser Quick Start](https://goreleaser.com/quick-start/)
- [GoReleaser GitHub Actions Integration](https://goreleaser.com/ci/actions/)
- [GoReleaser Homebrew Casks Documentation](https://goreleaser.com/customization/homebrew_casks/)
- [GoReleaser Homebrew Formulas (Deprecated)](https://goreleaser.com/customization/homebrew_formulas/)
- [GoReleaser Changelog Generation](https://goreleaser.com/customization/changelog/)
- [GoReleaser v2.10 Release - Cask Migration](https://goreleaser.com/blog/goreleaser-v2.10/)
- [GoReleaser Supply Chain Security](https://goreleaser.com/blog/supply-chain-security/)
- [GoReleaser Signing Documentation](https://goreleaser.com/customization/sign/)
- [GoReleaser macOS Universal Binaries](https://goreleaser.com/customization/universalbinaries/)

### Go CLI Distribution Best Practices
- [Go Official Documentation - CLIs](https://go.dev/solutions/clis)
- [Building and distributing a command-line app in Go](https://mauricio.github.io/2022/04/01/building-and-distributing-command-line-app-in-go.html)
- [From Go Code to Homebrew Tap: Writing and Deploying a Whisper CLI with GoReleaser](https://appliedgo.net/whisper-cli/)
- [CLI tools FTW (or: how to distribute your CLI tools with goreleaser)](https://appliedgo.net/release/)
- [Deploying Go CLI Applications](https://medium.com/@ben.lafferty/deploying-go-cli-applications-316e9cca16a4)

### Release Automation and GitHub Actions
- [goreleaser/goreleaser-action](https://github.com/goreleaser/goreleaser-action)
- [How to Configure GitHub Actions for Release Automation (2026)](https://oneuptime.com/blog/post/2026-02-02-github-actions-release-automation/view)
- [How to Automate Releases with GitHub Actions (2026)](https://oneuptime.com/blog/post/2026-01-25-automate-releases-github-actions/view)
- [wangyoucao577/go-release-action](https://github.com/wangyoucao577/go-release-action)

### Multi-Architecture and Cross-Platform
- [Creating Multi-architecture GitHub Releases for Go Binaries](https://ningbozhao.github.io/create-multi-arch-github-release-go-binary/)
- [Go on ARM and Beyond](https://go.dev/blog/ports)
- [How to Build Multi-Architecture Docker Images (ARM64 + AMD64) (2026)](https://oneuptime.com/blog/post/2026-01-06-docker-multi-architecture-images/view)
- [A Detailed Guide to Golang Cross-Platform Cross-Compilation Technology and Practice](https://www.oreateai.com/blog/a-detailed-guide-to-golang-crossplatform-crosscompilation-technology-and-practice/4fb21cf3401260e66b46cd33c4f9cd18)

### Security Features
- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)
- [GoReleaser SBOM Generation Discussion](https://github.com/goreleaser/goreleaser/issues/2597)
- [GitHub Artifact Naming Conventions - Kairos](https://kairos.io/docs/reference/artifacts/)

### CLI Best Practices and User Expectations
- [CLI tool software release process](https://medium.com/@henvic/cli-tool-software-release-1da0ef664323)
- [Command Line Interface Guidelines](https://clig.dev/)
- [UX patterns for CLI tools](https://www.lucasfcosta.com/blog/ux-patterns-cli-tools)

### Semantic Versioning and Automation
- [go-semantic-release/semantic-release](https://github.com/go-semantic-release/semantic-release)
- [Semantic Release - GoReleaser Cookbook](https://goreleaser.com/cookbooks/semantic-release/)
- [How to Implement Semantic Versioning Automation (2026)](https://oneuptime.com/blog/post/2026-01-25-semantic-versioning-automation/view)

### Homebrew Distribution
- [Homebrew Cask vs Formula Discussion](https://github.com/orgs/goreleaser/discussions/5563)
- [Creating Homebrew Formulas With GoReleaser](https://dzone.com/articles/creating-homebrew-formulas-with-goreleaser)
- [Packaging a project release (goreleaser part 2)](https://appliedgo.net/release2/)

---
*Feature research for: Go CLI Release Automation with Homebrew Distribution*
*Researched: 2026-02-13*
