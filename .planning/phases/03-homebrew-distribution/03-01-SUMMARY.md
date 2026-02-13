---
phase: 03-homebrew-distribution
plan: 01
subsystem: release-automation
tags:
  - goreleaser
  - homebrew
  - cask
  - distribution
  - macos
dependency_graph:
  requires:
    - 02-01-goreleaser-config
  provides:
    - modern-homebrew-cask-config
    - macos-quarantine-removal
  affects:
    - homebrew-tap-publishing
tech_stack:
  added:
    - homebrew_casks (GoReleaser v2.10+)
  patterns:
    - post-install-hooks
    - macos-quarantine-removal
key_files:
  created: []
  modified:
    - /workspace/wakafetch/.goreleaser.yaml
decisions:
  - migrated-to-homebrew-casks-from-deprecated-brews
  - added-xattr-quarantine-removal-hook
  - using-goreleaserbot-commit-author
metrics:
  duration_seconds: 36
  tasks_completed: 2
  files_modified: 1
  commits: 1
  completed: 2026-02-13T23:20:09Z
---

# Phase 03 Plan 01: Migrate to Homebrew Casks Summary

**One-liner:** Migrated GoReleaser config from deprecated brews to homebrew_casks with macOS quarantine removal using xattr hook

## What Was Built

Updated .goreleaser.yaml to use modern `homebrew_casks` section instead of deprecated `brews` section, preparing for GoReleaser v3 compatibility and enabling proper Homebrew cask distribution.

## Tasks Completed

| Task | Name | Commit | Files Modified |
|------|------|--------|----------------|
| 1 | Update GoReleaser to use homebrew_casks | 0a4853d | .goreleaser.yaml |
| 2 | Commit and push GoReleaser changes | 0a4853d | (push only) |

## Key Changes

### Configuration Migration
- **Replaced:** `brews:` section with `homebrew_casks:`
- **Removed:** `test:` block (not supported by casks)
- **Removed:** `install:` block (casks use different installation mechanism)
- **Added:** `commit_author` with goreleaserbot credentials
- **Added:** `commit_msg_template` for automated formula updates
- **Added:** `hooks.post.install` for macOS quarantine removal

### macOS Quarantine Removal
Added post-install hook to remove macOS quarantine attribute from the binary:
```ruby
if OS.mac?
  system_command "/usr/bin/xattr", args: ["-dr", "com.apple.quarantine", "#{staged_path}/wakafetch"]
end
```

This prevents macOS Gatekeeper from blocking the binary as "unidentified developer" software.

## Deviations from Plan

None - plan executed exactly as written.

## Verification Results

- ✅ .goreleaser.yaml contains `homebrew_casks:` section (not `brews:`)
- ✅ Contains macOS quarantine removal hook with xattr command
- ✅ Configuration syntax validated (manual verification - goreleaser not installed)
- ✅ Changes committed and pushed to b00y0h/wakafetch repository

## Success Criteria Met

- ✅ GoReleaser config updated to use homebrew_casks instead of deprecated brews
- ✅ Post-install hook for quarantine removal present
- ✅ Changes committed (0a4853d) and pushed to b00y0h/wakafetch
- ✅ Ready for version tag to trigger release

## Impact

**Immediate:**
- Configuration now uses modern GoReleaser v2.10+ syntax
- Prepared for GoReleaser v3 (when brews section will be removed)
- macOS users will no longer encounter Gatekeeper warnings

**Future:**
- Next release will publish to homebrew-wakafetch tap using cask format
- Automated formula updates via goreleaserbot
- Improved user experience on macOS systems

## Next Steps

Proceed to 03-02: Test Homebrew cask publishing with an actual release.

## Self-Check: PASSED

**Files created:** None
- N/A - no new files created

**Files modified:**
- ✅ /workspace/wakafetch/.goreleaser.yaml exists and contains homebrew_casks section

**Commits:**
- ✅ Commit 0a4853d exists in git history
