# Project Research Summary

**Project:** wakafetch Homebrew Distribution
**Domain:** Go CLI Release Automation with Homebrew Tap Distribution
**Researched:** 2026-02-13
**Confidence:** HIGH

## Executive Summary

This project involves automating the release and distribution of wakafetch, a Go CLI tool, through Homebrew on macOS and Linux. The standard approach is to use GoReleaser (v2.13.3+) with GitHub Actions for automated multi-platform builds, releases, and Homebrew tap updates. GoReleaser handles the entire release pipeline: cross-compilation for multiple platforms, archive creation, checksum generation, GitHub Release creation, and automatic Homebrew cask generation.

The recommended approach uses a tag-triggered GitHub Actions workflow that runs GoReleaser to build binaries for macOS (arm64/amd64), Linux (arm64/amd64), and Windows (amd64). GoReleaser then creates a GitHub Release and automatically commits an updated Homebrew cask to a separate tap repository (homebrew-wakafetch). This approach is well-established with extensive documentation and strong community support. The key enabler is GoReleaser's built-in integration with Homebrew taps using the modern `homebrew_casks` configuration (the older `brews` approach is deprecated as of v2.10).

The primary risk is cross-repository authentication. The default GitHub Actions GITHUB_TOKEN cannot write to the separate tap repository. This requires creating a dedicated Personal Access Token (PAT) or GitHub App token with repo permissions for the tap repository. Security best practice demands using a bot account (not a personal account) to minimize blast radius if the token is compromised. Other critical pitfalls include using deprecated configurations, missing shallow clone configuration (fetch-depth: 0), and accidentally publishing pre-release versions to the main tap. Following established patterns and running goreleaser check before releases mitigates these risks.

## Key Findings

### Recommended Stack

The stack centers on GoReleaser as the complete release automation solution, integrated with GitHub Actions for CI/CD orchestration. GoReleaser is the industry standard for Go CLI releases with 15,557+ GitHub stars and active development (latest release January 10, 2026).

**Core technologies:**
- **GoReleaser v2.13.3+**: Complete release automation handling builds, packaging, signing, and publishing — recommended because it consolidates multi-platform builds, Homebrew tap generation, checksums, signing, and GitHub releases into one actively-maintained tool
- **Go 1.24+ (stable)**: Build toolchain required by GoReleaser — use "stable" in GitHub Actions for latest stable release
- **GitHub Actions (v4+ checkout, v5+ setup-go, v6+ goreleaser-action)**: CI/CD automation — native GitHub integration, free for public repos, well-maintained actions with tight GoReleaser integration
- **Homebrew Tap (separate repository)**: Binary distribution for macOS/Linux — standard package manager convention, GoReleaser generates and updates cask automatically, requires repository name like `homebrew-wakafetch`
- **SHA-256 checksums**: Artifact verification — default in GoReleaser, automatically generated for all release artifacts

**Authentication tokens:**
- **GITHUB_TOKEN**: Built-in token for creating releases in same repository
- **Personal Access Token (PAT) or GitHub App Token**: Required for cross-repository publishing to separate Homebrew tap, needs `repo` or `contents:write` permissions, should be generated from dedicated bot account

**Optional security enhancements:**
- Cosign for keyless signing (modern alternative to GPG)
- macOS notarization (requires $99/year Apple Developer Account, only needed for App Store or enterprise distribution)

### Expected Features

Based on competitive analysis (kubectl, gh, hugo) and user expectations for Go CLI tools, features fall into three tiers.

**Must have (table stakes):**
- Multi-platform binaries (macOS arm64/amd64, Linux arm64/amd64, Windows amd64) — users expect CLIs to work on their platform
- GitHub Release automation — standard distribution method
- Homebrew cask in personal tap — macOS users expect `brew install` for CLI tools
- SHA-256 checksums — security-conscious users verify downloads
- Automated changelog — users expect to see what changed
- Archive packaging (.tar.gz for Unix, .zip for Windows) — standard formats
- Semantic versioning (v1.2.3 format) — industry standard

**Should have (competitive):**
- Binary signatures (GPG/Cosign) — enterprise security requirements (add when first requested)
- Conventional Commits for better changelogs — improves when changelog quality complaints arise
- macOS Universal Binaries — single binary for both Intel and ARM Macs (add if users complain about confusion)
- SBOM generation — supply chain security compliance (add when enterprise users request)

**Defer (v2+):**
- AI-enhanced release notes — requires GoReleaser Pro subscription, benefit unclear
- Automated semantic versioning — adds workflow complexity, manual tags work fine initially
- Docker image distribution — add if container users request it (5+ requests)
- Beta/stable tap separation — only needed when significant user base exists
- Homebrew core submission — only after 1000+ GitHub stars and stable 1.0 release

**Anti-features (explicitly do NOT build):**
- Custom download server — use GitHub Releases instead (free, trusted, CDN)
- Manual changelog editing — auto-generate from conventional commits
- Version numbers in code — inject via `-ldflags` at build time
- Custom update mechanism — let Homebrew handle updates
- Building on release server — build in CI for reproducibility

### Architecture Approach

The architecture follows a linear pipeline pattern triggered by git tag pushes. The pipeline executes seven sequential phases: before hooks (tests/linting), builds (cross-compilation matrix), archives (packaging with extras), checksums (SHA-256 generation), optional signing (GPG/Cosign), GitHub Release creation, and Homebrew cask generation/publishing. This pattern is standard across the Go CLI ecosystem and well-supported by tooling.

**Major components:**
1. **Source Repository (wakafetch)** — holds Go source code, `.goreleaser.yaml` config, and `.github/workflows/release.yml` GitHub Actions workflow
2. **GitHub Actions Workflow** — triggers on semantic version tags (v*), orchestrates checkout with full git history, sets up Go toolchain, and runs GoReleaser action
3. **GoReleaser Engine** — executes seven-phase pipeline producing multi-platform binaries, archives, checksums, and GitHub Release assets
4. **dist/ Directory** — temporary storage for build artifacts (git-ignored, cleaned between runs)
5. **GitHub Release** — hosts binaries and archives for public download via GITHUB_TOKEN
6. **Tap Repository (homebrew-wakafetch)** — separate repository containing `Casks/wakafetch.rb` auto-generated by GoReleaser
7. **Homebrew Cask** — Ruby DSL defining download URLs, SHA-256 checksums, and installation instructions

**Key architectural patterns:**
- **Tag-triggered releases**: Workflow triggers only on semantic version tags, providing clean separation between development and release
- **Cross-repository publishing**: Uses separate PAT/GitHub App token for pushing to tap repository (GITHUB_TOKEN cannot write to other repos)
- **Multi-platform build matrix**: Go's cross-compilation builds all GOOS/GOARCH combinations in one CI run
- **Homebrew casks over deprecated brews**: Modern approach for pre-compiled binaries (as of GoReleaser v2.10+)

**Data flow:**
Git tag push → GitHub Actions trigger → GoReleaser build pipeline → GitHub Release creation → Homebrew cask generation → commit to tap repository → user installation via `brew tap` and `brew install`

### Critical Pitfalls

The research identified 12 pitfalls ranging from critical (project-breaking) to moderate (quality-impacting). The top five critical pitfalls all occur during initial setup and can prevent successful releases.

1. **Wrong GitHub Token for Cross-Repository Publishing** — GoReleaser fails with 404 errors when using default GITHUB_TOKEN to push to tap repository. Prevention: Create separate PAT or GitHub App token with `repo` permissions, add as HOMEBREW_TAP_TOKEN secret, reference in GoReleaser config.

2. **Personal Access Token Security Risk** — Using personal user's PAT exposes all user repositories if token is compromised. Prevention: Create dedicated bot GitHub account, generate PAT from bot account, give bot only push access to tap repository, rotate PAT periodically.

3. **Formulas vs Casks Confusion** — Using deprecated `brews` configuration instead of `homebrew_casks` violates Homebrew semantics and faces deprecation. Prevention: Use `homebrew_casks` section in `.goreleaser.yaml`, use Casks/ directory in tap repository, run `goreleaser check` to detect deprecated fields.

4. **Missing fetch-depth: 0 in GitHub Actions** — Shallow clone prevents GoReleaser from generating changelogs and detecting version information. Prevention: Always include `fetch-depth: 0` in `actions/checkout@v4` step.

5. **Pre-release Version Overwriting Production** — Tagging pre-release versions (v1.0.0-rc1) updates main cask, replacing stable version unexpectedly. Prevention: Set `skip_upload: auto` in homebrew_casks configuration to automatically skip pre-release publishing.

**Additional moderate pitfalls:**
- Insufficient GitHub Actions workflow permissions (missing `contents: write`)
- Formula naming convention violations (incorrect class name capitalization)
- Missing or inadequate test blocks in cask definitions
- Multi-architecture archive conflicts (multiple OS/arch combinations without explicit configuration)

## Implications for Roadmap

Based on research, suggested phase structure prioritizes getting the basic release pipeline working end-to-end before adding enhancements. The architecture dictates a sequential approach where each phase builds on working infrastructure from the previous phase.

### Phase 1: Core Release Pipeline Setup

**Rationale:** Must establish working end-to-end pipeline before any features. All critical pitfalls occur in this phase. Cannot validate approach without successful release. Foundation for all subsequent work.

**Delivers:**
- Working GoReleaser configuration (.goreleaser.yaml)
- GitHub Actions release workflow (.github/workflows/release.yml)
- Separate Homebrew tap repository (homebrew-wakafetch)
- Bot account with properly scoped PAT
- First successful test release (v0.1.0)

**Addresses (from FEATURES.md):**
- Multi-platform binaries (macOS/Linux/Windows with arm64/amd64)
- GitHub Release creation
- Checksums (SHA-256)
- Archive packaging (.tar.gz/.zip)
- Semantic versioning

**Avoids (from PITFALLS.md):**
- Wrong GitHub token (Critical Pitfall 1) — configure HOMEBREW_TAP_TOKEN from start
- Personal PAT security risk (Critical Pitfall 2) — create bot account immediately
- Formulas vs Casks confusion (Critical Pitfall 3) — use homebrew_casks from start
- Missing fetch-depth (Critical Pitfall 6) — configure in initial workflow
- Multi-arch conflicts (Critical Pitfall 5) — configure architecture targets clearly

**Critical path items:**
1. Create homebrew-wakafetch tap repository first (GoReleaser needs it to exist)
2. Create bot account and generate PAT before testing pipeline
3. Add PAT as HOMEBREW_TAP_TOKEN secret before first test release
4. Configure .goreleaser.yaml with homebrew_casks (not deprecated brews)
5. Test with snapshot release before pushing first tag

### Phase 2: Homebrew Tap Publication

**Rationale:** Once pipeline works, focus on Homebrew distribution quality. Adds user-facing polish and prevents common installation issues. Validates that end-users can actually install and use the tool.

**Delivers:**
- Homebrew cask with proper test block
- Pre-release handling (skip_upload: auto)
- Automated changelog from commit messages
- README in tap repository with usage instructions
- Verification that `brew tap` and `brew install` work end-to-end

**Addresses (from FEATURES.md):**
- Homebrew tap publication (table stakes)
- Automated changelog (table stakes)
- README/docs in release (table stakes)

**Avoids (from PITFALLS.md):**
- Pre-release overwriting production (Critical Pitfall 4) — configure skip_upload: auto
- Missing test block (Moderate Pitfall 10) — add functional test verification
- Insufficient workflow permissions (Moderate Pitfall 8) — verify contents:write permission
- Formula naming violations (Moderate Pitfall 9) — run brew audit --strict

**Uses (from STACK.md):**
- GoReleaser homebrew_casks configuration
- GitHub Actions workflow permissions
- Conventional commit format for changelog

**Implements (from ARCHITECTURE.md):**
- Homebrew cask generation (Phase 7 of GoReleaser pipeline)
- Tap repository structure (Casks/ directory, README)
- User installation flow verification

### Phase 3: Release Process Documentation

**Rationale:** Codify release process for repeatability and future maintainers. Establishes operational patterns before they're forgotten. Prevents mistakes during subsequent releases.

**Delivers:**
- Release runbook (RELEASING.md)
- Pre-release checklist
- Token rotation procedure
- Troubleshooting guide for common failures
- Test procedure for M1/Intel Macs

**Addresses (from FEATURES.md):**
- None directly, but enables sustainable long-term operation

**Avoids (from PITFALLS.md):**
- Deprecated configuration fields (Moderate Pitfall 7) — document how to check
- Recovery strategies for all critical pitfalls — document response procedures
- Token security mistakes — document rotation policy

### Phase 4: Enhanced Security Features (Optional)

**Rationale:** Add when enterprise users or security-conscious users request. Not needed for MVP. Can defer until proven demand.

**Delivers:**
- Binary signatures with Cosign (keyless signing via GitHub OIDC)
- SBOM generation for supply chain transparency
- GitHub attestations for build provenance

**Addresses (from FEATURES.md):**
- Binary signatures (should have, add when requested)
- SBOM generation (should have, add for enterprise)

**Uses (from STACK.md):**
- Cosign (optional, modern alternative to GPG)
- GoReleaser signing configuration
- GitHub Actions OIDC for keyless signing

### Phase Ordering Rationale

- **Sequential dependency chain**: Cannot test Homebrew publication (Phase 2) without working release pipeline (Phase 1). Cannot document process (Phase 3) without stable working pipeline.
- **Risk mitigation first**: All critical pitfalls addressed in Phase 1 before adding features. Foundation must be solid before building on it.
- **User value priority**: Phases 1-2 deliver complete end-user value (working Homebrew installation). Phase 3 is operational excellence. Phase 4 is optional enhancement.
- **Validation gates**: Each phase includes end-to-end validation before proceeding. Phase 1 must produce successful release. Phase 2 must complete successful `brew install`. Phase 3 must have runbook tested by following it.
- **Defer optional complexity**: Security features (Phase 4) are valuable but not blocking. Add based on actual user requests, not speculation.

### Research Flags

**Phases with standard patterns (skip research-phase):**
- **Phase 1: Core Release Pipeline Setup** — extremely well-documented, GoReleaser official docs are comprehensive, hundreds of example repositories available, established patterns
- **Phase 2: Homebrew Tap Publication** — standard Homebrew cask patterns, clear documentation, examples in ARCHITECTURE.md research
- **Phase 3: Release Process Documentation** — internal documentation, no new technical research needed

**Phases NOT needing deeper research:**
All phases use well-established patterns with HIGH confidence research. GoReleaser and Homebrew documentation is excellent. No phases require `/gsd:research-phase` during planning.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Official GoReleaser documentation, active project (latest release Jan 2026), verified version compatibility, clear migration path from deprecated patterns |
| Features | HIGH | Competitive analysis of major tools (kubectl, gh, hugo), clear table stakes vs differentiators, strong community consensus on MVP scope |
| Architecture | HIGH | Official architecture documentation, well-established patterns across Go CLI ecosystem, clear integration points and data flow |
| Pitfalls | MEDIUM | Based on GitHub issues, community tutorials, and official deprecation notices. Some findings need validation during implementation, but core pitfalls are well-documented |

**Overall confidence:** HIGH

The stack choice (GoReleaser) is industry standard with comprehensive documentation. The architecture patterns are proven across thousands of Go CLI projects. The main uncertainty is in pitfall recovery strategies which may evolve as GoReleaser versions change.

### Gaps to Address

Areas where research was inconclusive or needs validation during implementation:

- **macOS notarization requirements post-2025** — Apple Developer documentation may have changed. Validation: Check current Apple notarization requirements if/when pursuing enterprise distribution. Impact: LOW (optional feature, defer to Phase 4).

- **GitHub Actions pricing changes (Jan 1, 2026)** — May affect cost for high-volume projects. Validation: Review current GitHub Actions pricing during Phase 1 setup. Impact: LOW (free tier sufficient for typical release cadence).

- **GoReleaser Pro features** — Research covered OSS version. Pro features (AI release notes, advanced SBOM) may provide value. Validation: Evaluate Pro subscription after MVP if changelog quality becomes concern. Impact: LOW (not blocking).

- **Fine-grained PAT vs GitHub App tokens** — Research mentions GitHub Apps as "recommended alternative" but GoReleaser documentation primarily shows PAT examples. Validation: Test GitHub App token support during Phase 1 setup. Impact: MEDIUM (affects security posture, but both approaches work).

- **Homebrew core submission requirements** — Research notes "after 1000+ stars" but exact requirements may change. Validation: Review Homebrew core submission guidelines before pursuing. Impact: LOW (defer to post-v1.0).

**Handling during planning:**
- Document assumptions in roadmap
- Include validation checkpoints in Phase 1 and Phase 2 tasks
- Add "verify current requirements" tasks for items that may have changed
- Plan for alternative approaches where uncertainty exists (e.g., PAT vs GitHub App tokens)

## Sources

### Primary (HIGH confidence)
- [GoReleaser Official Documentation](https://goreleaser.com/) — Core features, configuration, CI/CD integration
- [GoReleaser GitHub Releases](https://github.com/goreleaser/goreleaser/releases) — Version compatibility, latest features
- [GoReleaser GitHub Actions](https://goreleaser.com/ci/actions/) — Official CI/CD patterns
- [Homebrew Documentation](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap) — Tap structure, cask conventions
- [Go Release History](https://go.dev/doc/devel/release) — Version requirements
- [GitHub Actions Documentation](https://docs.github.com/en/actions) — Workflow configuration, permissions

### Secondary (MEDIUM confidence)
- [Applied Go: Whisper CLI Tutorial](https://appliedgo.net/whisper-cli/) — March 2025 current best practices
- [Medium: Automating Golang Releases](https://medium.com/@wprimadi/automating-golang-project-releases-with-goreleaser-8ccba7cd2a9e) — May 2025 patterns
- [GitHub Issues and Discussions](https://github.com/goreleaser/goreleaser/issues) — Real-world pitfalls and solutions
- Community tutorials from 2025-2026 — Current implementation patterns

### Tertiary (LOW confidence)
- macOS notarization requirements — May have changed post-2025, needs validation
- GitHub Actions pricing — Effective Jan 1, 2026 changes, needs verification
- GoReleaser Pro features — Limited public documentation for Pro-only features

---
*Research completed: 2026-02-13*
*Ready for roadmap: yes*
