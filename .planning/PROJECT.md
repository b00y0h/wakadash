# wakafetch Homebrew Distribution

## What This Is

Release automation for the wakafetch Go CLI that enables macOS and Linux users to install via Homebrew. This adds GoReleaser and GitHub Actions to a fork of the upstream project, with a dedicated Homebrew tap for formula distribution.

## Core Value

Users can install wakafetch with `brew tap b00y0h/wakafetch && brew install wakafetch` — no manual downloads or Go toolchain required.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Fork b00y0h/wakafetch from upstream sahaj-b/wakafetch
- [ ] GoReleaser config for multi-platform builds (darwin/linux, amd64/arm64)
- [ ] GitHub Actions workflow triggered on version tags (v*.*.*)
- [ ] Homebrew tap repo (b00y0h/homebrew-wakafetch) with README
- [ ] Cross-repo PAT setup for formula publishing
- [ ] Working end-to-end release flow with v0.1.0

### Out of Scope

- App code changes — only release/CI infrastructure
- Windows builds — Homebrew is macOS/Linux focused
- Additional package managers (apt, yum, etc.) — Homebrew only for v1
- Automated dependency updates — manual releases only

## Context

**Upstream:** sahaj-b/wakafetch (Go CLI tool)
**Fork:** b00y0h/wakafetch (will contain release automation)
**Tap:** b00y0h/homebrew-wakafetch (Homebrew formula distribution)

GoReleaser will:
- Build binaries for 4 platform/arch combinations
- Create GitHub releases with tar.gz archives
- Push formula to the tap repo automatically

Cross-repo publishing requires a Personal Access Token with Contents read/write on the tap repo.

## Constraints

- **No secrets in code**: All tokens via GitHub Actions secrets, never hardcoded
- **No app modifications**: Only add .goreleaser.yml and .github/workflows/release.yml
- **Minimal config**: Idiomatic GoReleaser, no unnecessary features
- **First release version**: v0.1.0

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Fork instead of tap-only | Full control over releases; upstream may not have releases | — Pending |
| GoReleaser over manual | Industry standard for Go release automation | — Pending |
| Fine-grained PAT | More secure than classic PAT with broad repo scope | — Pending |

---
*Last updated: 2026-02-13 after initialization*
