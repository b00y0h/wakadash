# Phase 11: Configuration & Validation - Context

**Gathered:** 2026-02-24
**Status:** Ready for planning

<domain>
## Phase Boundary

Add `history_repo` config support in a dedicated `[wakadash]` section. Dashboard starts gracefully when config is missing or invalid. This phase only adds config reading — actually fetching from the repo is Phase 12.

</domain>

<decisions>
## Implementation Decisions

### Config format
- Key: `history_repo` in `[wakadash]` section of ~/.wakatime.cfg
- Value format: `owner/repo` only (e.g., `b00y0h/wakatime-data`)
- Dashboard normalizes to GitHub URL internally

### Section creation
- Auto-create `[wakadash]` section on first run if missing
- Include commented example with link to wakasync repo
- Template: commented `history_repo` line + setup instructions URL

### Claude's Discretion
- Validation behavior (when to validate, warning vs silent fallback)
- Error messaging format
- Fallback behavior when history_repo not configured

</decisions>

<specifics>
## Specific Ideas

- Comment in auto-created section should link to wakasync repo so users know how to set up their archive
- Keep it simple — user uncomments and fills in their repo

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 11-configuration-validation*
*Context gathered: 2026-02-24*
