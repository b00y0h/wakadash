---
phase: 07-distribution
plan: 02
subsystem: distribution
tags: [homebrew-core, formula, source-build, pr-submission]

# Dependency graph
requires:
  - phase: 07-01
    provides: Working personal tap with verified installation
provides:
  - homebrew-core formula ready for resubmission
  - Source-build formula using Go toolchain
  - Test block with version and config error validation
affects: [distribution, discoverability]

# Tech tracking
tech-stack:
  added: [homebrew-formula, std_go_args]
  patterns: [source-build-formula, homebrew-test-block]

key-files:
  created: [Formula/w/wakadash.rb (in fork)]
  modified: []

key-decisions:
  - "Submit early with transparency note about popularity thresholds"
  - "Use std_go_args helper for Go source builds"
  - "Include test for graceful config error handling"
  - "Target main branch for future resubmission (not master)"

patterns-established:
  - "Pattern 1: homebrew-core requires >=30 forks, >=30 watchers, or >=75 stars for self-submitted software"
  - "Pattern 2: Personal tap provides distribution while building community adoption"
  - "Pattern 3: Formula test block should validate both version output and error handling"

# Metrics
duration: 15min
completed: 2026-02-19
---

# Phase 7 Plan 2: Homebrew Core Formula Summary

**Source-build formula submitted to homebrew-core; closed pending popularity threshold (expected for new projects)**

## Performance

- **Duration:** 15 min
- **Started:** 2026-02-19T21:25:00Z
- **Completed:** 2026-02-19T21:41:00Z
- **Tasks:** 7 (5 automated + 2 human checkpoints)
- **PR:** #268434 (closed)

## Outcome

**PR Status:** CLOSED by maintainer @chenrui333

**Reason:** Project doesn't meet homebrew-core popularity thresholds:
- Required: ≥30 forks, ≥30 watchers, or ≥75 stars
- wakadash: New project (created 2026-02-19), thresholds not yet met

**This was expected** — the plan included a transparency note acknowledging early submission.

## Accomplishments

- Created source-build formula at `Formula/w/wakadash.rb`
- Formula uses `std_go_args` helper with Go toolchain
- Test block validates version output and graceful config error
- Submitted PR #268434 to Homebrew/homebrew-core
- Formula ready for resubmission when popularity thresholds are met

## Task Commits

1. **Task 0: Fork Homebrew/homebrew-core** - Fork exists at b00y0h/homebrew-core
2. **Task 1: Set up local dev environment** - Adapted for ClaudeBox (no Homebrew)
3. **Task 2: Get SHA256 of source tarball** - `8cd2bfef0fc8399fb1b1e34557b0a83afecf53c2a907fa8e349f3ca3298988e4`
4. **Task 3: Create wakadash formula** - `fa0cac7`
5. **Task 4: Test formula locally** - Build verified with Go toolchain
6. **Task 5: Commit and push to fork** - Pushed to b00y0h/homebrew-core:wakadash
7. **Task 6: Create PR to homebrew-core** - PR #268434 created
8. **Task 7: Verify PR submission** - PR closed by maintainer (expected)

## Formula Created

```ruby
class Wakadash < Formula
  desc "Live terminal dashboard for WakaTime coding stats"
  homepage "https://github.com/b00y0h/wakadash"
  url "https://github.com/b00y0h/wakadash/archive/refs/tags/v0.2.0.tar.gz"
  sha256 "8cd2bfef0fc8399fb1b1e34557b0a83afecf53c2a907fa8e349f3ca3298988e4"
  license "MIT"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X main.version=#{version}
      -X main.commit=#{tap.user}
      -X main.date=#{time.iso8601}
    ]
    system "go", "build", *std_go_args(ldflags:), "./cmd/wakadash"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/wakadash --version")
    output = shell_output("#{bin}/wakadash 2>&1", 1)
    assert_match(/wakatime|config|api/i, output)
  end
end
```

## Decisions Made

1. **Early submission strategy:** Submit with transparency note to get formula feedback while building adoption
2. **Test block design:** Validate both version output and graceful error handling (no API key required)
3. **Branch targeting:** Future resubmission must target `main` branch (not `master`)

## Deviations from Plan

**[Adapted] ClaudeBox environment without Homebrew**
- Plan assumed local Homebrew installation for `brew audit`, `brew test`
- Adapted: Cloned repo directly, verified build with Go toolchain
- Formula syntax validated with `ruby -c`

**[Noted] Branch targeting issue**
- PR targeted `master` instead of `main`
- Homebrew's default branch is now `main`
- Fixed in formula branch for future resubmission

## Issues Encountered

1. **PR closed for popularity:** Expected outcome for new project
2. **Branch targeting:** PR should have targeted `main` not `master`

## Resubmission Checklist

When wakadash reaches popularity thresholds:

- [ ] Verify ≥30 forks OR ≥30 watchers OR ≥75 stars
- [ ] Update formula SHA256 if new version released
- [ ] Create fresh branch from `upstream/main` (not master)
- [ ] Submit new PR with updated metrics
- [ ] Remove transparency note about early submission

## Current Distribution Status

| Method | Status | Command |
|--------|--------|---------|
| Personal tap | ✓ Working | `brew tap b00y0h/wakadash && brew install wakadash` |
| homebrew-core | Pending popularity | Resubmit when thresholds met |

## Self-Check: PASSED

- ✅ Formula exists in b00y0h/homebrew-core:wakadash branch
- ✅ PR #268434 was submitted (closed as expected)
- ✅ Personal tap working as fallback distribution

---
*Phase: 07-distribution*
*Completed: 2026-02-19*
