# Roadmap: wakafetch Homebrew Distribution

## Overview

This roadmap delivers automated Homebrew distribution for the wakafetch Go CLI in three phases. Phase 1 establishes GitHub infrastructure (fork, tap repo, authentication). Phase 2 implements GoReleaser automation for multi-platform builds and releases. Phase 3 completes Homebrew integration so users can install with `brew tap b00y0h/wakafetch && brew install wakafetch`.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Repository Setup** - GitHub infrastructure ready for automation
- [ ] **Phase 2: Release Automation** - GoReleaser builds and publishes releases
- [ ] **Phase 3: Homebrew Distribution** - Users can install via brew

## Phase Details

### Phase 1: Repository Setup
**Goal**: GitHub infrastructure is ready for release automation
**Depends on**: Nothing (first phase)
**Requirements**: SETUP-01, SETUP-02, SETUP-03, SETUP-04
**Success Criteria** (what must be TRUE):
  1. Fork b00y0h/wakafetch exists from upstream sahaj-b/wakafetch
  2. Tap repository b00y0h/homebrew-wakafetch exists with README
  3. Fine-grained PAT exists with Contents read/write on tap repo
  4. HOMEBREW_TAP_TOKEN secret configured in wakafetch repository
**Plans**: 3 plans in 2 waves

Plans:
- [ ] 01-01-PLAN.md - Fork upstream repository and configure settings
- [ ] 01-02-PLAN.md - Create tap repository and initialize structure
- [ ] 01-03-PLAN.md - Create PAT and configure repository secret (has checkpoint)

### Phase 2: Release Automation
**Goal**: Automated multi-platform releases work end-to-end
**Depends on**: Phase 1
**Requirements**: REL-01, REL-02, REL-03, REL-04, CI-01, CI-02, CI-03, CI-04
**Success Criteria** (what must be TRUE):
  1. Pushing a version tag triggers GitHub Actions workflow
  2. GoReleaser builds binaries for darwin/amd64, darwin/arm64, linux/amd64, linux/arm64
  3. GitHub Release is created with tar.gz archives and SHA256 checksums
  4. Release includes auto-generated changelog from commit history
  5. Workflow authenticates to tap repository using HOMEBREW_TAP_TOKEN
**Plans**: TBD

Plans:
- [ ] TBD (will be defined during phase planning)

### Phase 3: Homebrew Distribution
**Goal**: Users can install wakafetch via Homebrew
**Depends on**: Phase 2
**Requirements**: DIST-01, DIST-02, DIST-03, DIST-04
**Success Criteria** (what must be TRUE):
  1. GoReleaser automatically publishes Homebrew cask to tap repository on release
  2. Cask includes proper install block that places wakafetch binary in PATH
  3. User can run `brew tap b00y0h/wakafetch` successfully
  4. User can run `brew install wakafetch` and execute `wakafetch` command
  5. Cask includes test block for installation verification
**Plans**: TBD

Plans:
- [ ] TBD (will be defined during phase planning)

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Repository Setup | 0/3 | Planned | - |
| 2. Release Automation | 0/TBD | Not started | - |
| 3. Homebrew Distribution | 0/TBD | Not started | - |
