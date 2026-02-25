# Requirements: wakadash

**Defined:** 2026-02-25
**Core Value:** Real-time visibility into coding activity without leaving the terminal

## v2.2 Requirements

Requirements for version update check feature.

### Version Checking

- [ ] **VER-01**: App checks GitHub Releases API on startup (async, non-blocking)
- [ ] **VER-02**: App compares current version with latest release using semantic versioning
- [ ] **VER-03**: Network errors during version check fail silently (no error messages)
- [ ] **VER-04**: Version check times out after 5 seconds maximum

### Status Bar Display

- [ ] **DISP-01**: Status bar shows update notice when newer version available
- [ ] **DISP-02**: Update notice displays version diff (e.g., v2.2.0 → v2.3.0)
- [ ] **DISP-03**: Update notice shows brew upgrade command
- [ ] **DISP-04**: Update notice uses bordered box styling (gh CLI style)

## Future Requirements

Deferred to future release. Tracked but not in current roadmap.

### Caching

- **CACHE-01**: Cache version check results for 24 hours
- **CACHE-02**: Skip API call if checked within cache period

### Configuration

- **CFG-01**: User can disable version checks via config
- **CFG-02**: User can configure check frequency

### Multi-Platform

- **PLAT-01**: Detect installation method (Homebrew, go install, binary)
- **PLAT-02**: Show appropriate upgrade command per platform

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Auto-update | Keep manual like gh CLI — user controls when to update |
| Modal popups | Anti-pattern — status bar is non-intrusive |
| Release notes preview | Complexity, defer to future |
| Pre-release notifications | Focus on stable releases only |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| VER-01 | Phase 13 | Pending |
| VER-02 | Phase 13 | Pending |
| VER-03 | Phase 13 | Pending |
| VER-04 | Phase 13 | Pending |
| DISP-01 | Phase 14 | Pending |
| DISP-02 | Phase 14 | Pending |
| DISP-03 | Phase 14 | Pending |
| DISP-04 | Phase 14 | Pending |

**Coverage:**
- v2.2 requirements: 8 total
- Mapped to phases: 8
- Unmapped: 0 ✓

---
*Requirements defined: 2026-02-25*
*Last updated: 2026-02-25 after roadmap creation*
