# Requirements: wakafetch Homebrew Distribution

**Defined:** 2026-02-13
**Core Value:** Users can install wakafetch with `brew tap b00y0h/wakafetch && brew install wakafetch`

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Release Infrastructure

- [ ] **REL-01**: GoReleaser builds binaries for darwin/amd64, darwin/arm64, linux/amd64, linux/arm64
- [ ] **REL-02**: GoReleaser creates tar.gz archives with OS, arch, and version in filename
- [ ] **REL-03**: GoReleaser generates SHA256 checksums for all archives
- [ ] **REL-04**: GoReleaser generates changelog from git commit history

### Distribution

- [ ] **DIST-01**: Homebrew tap repository (b00y0h/homebrew-wakafetch) exists with README
- [ ] **DIST-02**: GoReleaser publishes Homebrew cask to tap on release
- [ ] **DIST-03**: Homebrew cask includes proper install block for binary
- [ ] **DIST-04**: Users can install via `brew tap b00y0h/wakafetch && brew install wakafetch`

### CI/CD

- [ ] **CI-01**: GitHub Actions workflow exists at .github/workflows/release.yml
- [ ] **CI-02**: Workflow triggers on push of v*.*.* tags
- [ ] **CI-03**: Workflow uses goreleaser-action with full git history (fetch-depth: 0)
- [ ] **CI-04**: Cross-repo PAT (HOMEBREW_TAP_TOKEN) configured for tap publishing

### Setup

- [ ] **SETUP-01**: Fork b00y0h/wakafetch from upstream sahaj-b/wakafetch
- [ ] **SETUP-02**: Create b00y0h/homebrew-wakafetch repository
- [ ] **SETUP-03**: Create fine-grained PAT with Contents read/write on tap repo
- [ ] **SETUP-04**: Add PAT as HOMEBREW_TAP_TOKEN secret in wakafetch repo

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Enhanced Distribution

- **DIST-05**: Shell completions for bash, zsh, fish
- **DIST-06**: Man page generation
- **DIST-07**: Docker image publishing

### Advanced Release

- **REL-05**: Pre-release handling (skip tap update for RCs)
- **REL-06**: Binary signatures with Cosign
- **REL-07**: SBOM generation
- **REL-08**: Universal macOS binaries

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| App code changes | Only release/CI infrastructure, no modifications to wakafetch itself |
| Windows builds | Homebrew is macOS/Linux focused; Windows users can use Go install |
| Other package managers | apt, yum, scoop etc. deferred; Homebrew only for v1 |
| Homebrew core submission | Requires 1.0 stable + significant adoption; personal tap sufficient |
| macOS notarization | Only needed for App Store distribution; CLI tools don't require it |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| REL-01 | TBD | Pending |
| REL-02 | TBD | Pending |
| REL-03 | TBD | Pending |
| REL-04 | TBD | Pending |
| DIST-01 | TBD | Pending |
| DIST-02 | TBD | Pending |
| DIST-03 | TBD | Pending |
| DIST-04 | TBD | Pending |
| CI-01 | TBD | Pending |
| CI-02 | TBD | Pending |
| CI-03 | TBD | Pending |
| CI-04 | TBD | Pending |
| SETUP-01 | TBD | Pending |
| SETUP-02 | TBD | Pending |
| SETUP-03 | TBD | Pending |
| SETUP-04 | TBD | Pending |

**Coverage:**
- v1 requirements: 16 total
- Mapped to phases: 0
- Unmapped: 16 ⚠️

---
*Requirements defined: 2026-02-13*
*Last updated: 2026-02-13 after initial definition*
