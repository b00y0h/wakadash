# Phase 8: Theme Foundation - Context

**Gathered:** 2026-02-19
**Status:** Ready for planning

<domain>
## Phase Boundary

Users can select and persist color themes. Includes theme picker UI, persistence to ~/.wakatime.cfg, and applying theme colors consistently across all existing panels. Theme presets are fixed (6 options). Custom themes or theme editing are out of scope.

</domain>

<decisions>
## Implementation Decisions

### First-run preview
- Full-screen picker before dashboard loads (not an overlay)
- Single theme preview at a time, arrow keys to browse
- Preview shows mini dashboard with sample data (scaled-down version of actual panels)
- Enter key confirms selection (no number key shortcuts)

### Theme switching
- Press 't' from dashboard to open theme picker
- Same full-screen picker experience as first-run
- Dashboard resumes instantly with new theme applied (no confirmation message)
- 't - Change theme' shown in help overlay ('?' screen)

### Themed elements
- Full theming: borders, backgrounds, text all follow theme palette
- Heatmap uses theme's accent color gradient (not GitHub green)
- Header/title bar fully themed (title, refresh indicator, status)

### Claude's Discretion
- Language bar chart colors: Claude decides whether to use theme palette or keep GitHub Linguist colors (Go=cyan, Python=yellow, etc.) based on visual balance
- Exact mini dashboard layout in preview
- Theme palette structure (how many colors per theme, naming)
- Handling terminals with limited color support

</decisions>

<specifics>
## Specific Ideas

- Preview should feel instant and responsive as user browses themes
- Keep picker minimal - no progress indicator (e.g., "3 of 6"), just theme name as subtle label
- Sample data for preview (no API calls during theme selection)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 08-theme-foundation*
*Context gathered: 2026-02-19*
