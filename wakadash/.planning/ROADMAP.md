# Roadmap: wakadash v2.2

**Milestone:** v2.2 Version Update Check
**Depth:** Standard
**Coverage:** 8/8 requirements mapped

## Phases

- [ ] **Phase 13: Version Checking Backend** - GitHub API integration and semantic versioning
- [ ] **Phase 14: Status Bar Display Integration** - Update notice rendering in TUI
- [ ] **Phase 15: Caching & Production Polish** - Rate limiting and production readiness
- [ ] **Phase 16: Integration Testing & Validation** - End-to-end testing and edge cases

## Phase Details

### Phase 13: Version Checking Backend
**Goal**: Implement GitHub Releases API integration with semantic version comparison
**Depends on**: Nothing (first phase)
**Requirements**: VER-01, VER-02, VER-03, VER-04
**Success Criteria** (what must be TRUE):
  1. App fetches latest version from GitHub Releases API without blocking startup
  2. Version comparison correctly identifies when update is available (v1.10.0 > v1.9.0)
  3. Network failures and timeouts fail silently without error messages to user
  4. Version check completes or times out within 5 seconds maximum
**Plans**: TBD

### Phase 14: Status Bar Display Integration
**Goal**: Display update notification in status bar with version diff and upgrade command
**Depends on**: Phase 13
**Requirements**: DISP-01, DISP-02, DISP-03, DISP-04
**Success Criteria** (what must be TRUE):
  1. Status bar shows update notice when newer version is available
  2. Update notice displays clear version diff format (e.g., v2.2.0 → v2.3.0)
  3. Update notice shows actionable brew upgrade command
  4. Update notice uses bordered box styling consistent with gh CLI patterns
**Plans**: TBD

### Phase 15: Caching & Production Polish
**Goal**: Implement check throttling and cache management for production reliability
**Depends on**: Phase 14
**Requirements**: (Enables production deployment of VER and DISP requirements)
**Success Criteria** (what must be TRUE):
  1. Version check results are cached for 24 hours to avoid API rate limits
  2. Cached data is validated on startup and reused within cache period
  3. Update notice does not appear on every startup (respects cache interval)
  4. Corrupt cache files are handled gracefully without crashing
**Plans**: TBD

### Phase 16: Integration Testing & Validation
**Goal**: Validate end-to-end functionality with real GitHub API and edge cases
**Depends on**: Phase 15
**Requirements**: (Validates all VER and DISP requirements)
**Success Criteria** (what must be TRUE):
  1. Version check works with real wakadash GitHub repository
  2. Dashboard handles network disconnection gracefully without errors
  3. Update notice displays correctly on narrow terminals (80 columns minimum)
  4. All version checking scenarios tested (no update, update available, network failure, cache hit)
**Plans**: TBD

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 13. Version Checking Backend | 0/? | Not started | - |
| 14. Status Bar Display Integration | 0/? | Not started | - |
| 15. Caching & Production Polish | 0/? | Not started | - |
| 16. Integration Testing & Validation | 0/? | Not started | - |

## Coverage Map

**Phase 13: Version Checking Backend**
- VER-01: App checks GitHub Releases API on startup (async, non-blocking)
- VER-02: App compares current version with latest release using semantic versioning
- VER-03: Network errors during version check fail silently (no error messages)
- VER-04: Version check times out after 5 seconds maximum

**Phase 14: Status Bar Display Integration**
- DISP-01: Status bar shows update notice when newer version available
- DISP-02: Update notice displays version diff (e.g., v2.2.0 → v2.3.0)
- DISP-03: Update notice shows brew upgrade command
- DISP-04: Update notice uses bordered box styling (gh CLI style)

**Phase 15: Caching & Production Polish**
- (No direct requirements - enables production deployment)

**Phase 16: Integration Testing & Validation**
- (No direct requirements - validates all requirements)

**Total coverage:** 8/8 v2.2 requirements mapped ✓

---
*Roadmap created: 2026-02-25*
*Phase numbering: Continues from v2.1 (ended at phase 12)*
