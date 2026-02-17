# wakafetch Homebrew Distribution

## What This Is

Release automation for the wakafetch Go CLI that enables macOS and Linux users to install via Homebrew. Uses GoReleaser with GitHub Actions to build multi-platform binaries and auto-publish to a dedicated Homebrew tap.

## Core Value

Users can install wakafetch with `brew tap b00y0h/wakafetch && brew install wakafetch` — no manual downloads or Go toolchain required.

## Current State (v1.0)

**Shipped:** 2026-02-17

**Repositories:**
- **Fork:** b00y0h/wakafetch (GoReleaser + GitHub Actions)
- **Tap:** b00y0h/homebrew-wakafetch (Homebrew cask)
- **Upstream:** sahaj-b/wakafetch (Go CLI)

**Release pipeline:**
- Tag v*.*.* triggers GitHub Actions
- GoReleaser builds darwin/linux × amd64/arm64
- Cask auto-published to tap with quarantine removal hook

**Current release:** v0.1.0

## Requirements

### Validated

- ✓ Fork b00y0h/wakafetch from upstream — v1.0
- ✓ GoReleaser multi-platform builds (darwin/linux × amd64/arm64) — v1.0
- ✓ GitHub Actions workflow on version tags — v1.0
- ✓ Homebrew tap (b00y0h/homebrew-wakafetch) — v1.0
- ✓ Cross-repo PAT for formula publishing — v1.0
- ✓ End-to-end release with v0.1.0 — v1.0
- ✓ macOS quarantine removal (no Gatekeeper warnings) — v1.0

### Active

(None — milestone complete)

### Out of Scope

- App code changes — only release/CI infrastructure
- Windows builds — Homebrew is macOS/Linux focused
- Additional package managers (apt, yum, etc.) — Homebrew only
- brew.sh discoverability — requires homebrew-core submission with upstream cooperation
- Automated dependency updates — manual releases only

## Constraints

- **No secrets in code**: Tokens via GitHub Actions secrets only
- **No app modifications**: Only release infrastructure (.goreleaser.yaml, .github/workflows/)
- **PAT rotation**: Fine-grained token expires Feb 2027

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Fork instead of tap-only | Full control over releases; upstream has no releases | ✓ Good |
| GoReleaser over manual | Industry standard for Go release automation | ✓ Good |
| Fine-grained PAT | More secure than classic PAT with broad repo scope | ✓ Good |
| homebrew_casks over brews | Modern GoReleaser v2.10+ syntax, v3-ready | ✓ Good |
| xattr quarantine removal | Prevents macOS Gatekeeper warnings | ✓ Good |
| Keep personal tap only | brew.sh requires homebrew-core + upstream cooperation | — Accepted |

---
*Last updated: 2026-02-17 after v1.0 milestone*
