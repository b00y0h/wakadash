# Phase 9: Stats Panels + Summary - Context

**Gathered:** 2026-02-20
**Status:** Ready for planning

<domain>
## Phase Boundary

Dashboard displays Categories, Editors, Operating Systems, and Machines panels with top 10 items and time labels. Includes Summary panel showing Last 30d total, daily avg, top items, streaks, and counts. Panels arrange in 2-column layout on terminals ≥80 cols, stacking on smaller terminals. Users can toggle panel visibility with keyboard shortcuts. All panels use consistent visual styling with selected theme.

</domain>

<decisions>
## Implementation Decisions

### Panel Content
- Show top 10 items per panel (Categories, Editors, OS, Machines)
- Human-readable time format: "2h 15m" or "45 mins"
- Always show percentages alongside time: "VS Code: 4h 30m (65%)"
- Group unknown/untracked items as "Other" category

### Summary Panel
- Full overview: 30d total, daily avg, top language/project/editor, streaks, active days, item counts
- Highlighted visual treatment — subtle accent border or background (distinct from other panels)
- Position at top — first thing users see, overview before details
- Show both current streak and best streak: "Current: 5 days | Best: 14 days"

### Layout Behavior
- Wide terminals (≥80 cols): Summary spans full width at top, other panels in 2-column grid below
- Narrow terminals (<80 cols): All panels stack vertically, single column, full width each
- Truncate panels when terminal is too short (no scrolling viewport)
- Minimum 3 items per panel when truncating

### Toggle Shortcuts
- Number keys map to visual order: 1=first visible panel, 2=second, etc.
- Include show-all and hide-all shortcuts (e.g., Shift+A, Shift+H)
- Panel visibility does NOT persist across restarts — always start with all panels visible
- Help overlay shows abbreviated hint: "1-9: toggle panels" (not listing each individually)

### Claude's Discretion
- Exact column widths and spacing in 2-column layout
- Which specific keys for show-all/hide-all
- How to visually indicate truncated panels
- Exact accent styling for Summary panel
- Panel ordering within the 2-column grid

</decisions>

<specifics>
## Specific Ideas

- Summary panel should be visually distinct but not overwhelming — subtle emphasis
- "2h 15m" style matches what users expect from time tracking tools
- Visual order mapping for toggles is intuitive — press what you see

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 09-stats-panels-summary*
*Context gathered: 2026-02-20*
