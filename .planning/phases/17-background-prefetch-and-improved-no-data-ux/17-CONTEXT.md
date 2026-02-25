# Phase 17: Background Prefetch and Improved No-Data UX - Context

**Gathered:** 2026-02-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Two UX improvements for historical navigation:
1. Background prefetch of previous week on launch for instant backward navigation
2. Full-screen "end of history" banner when user reaches the oldest available data

</domain>

<decisions>
## Implementation Decisions

### Prefetch strategy
- Prefetch only the previous week (one week back from current)
- Trigger prefetch after main UI loads (dashboard displays first, then quietly fetch)
- Continue predictive loading as user navigates — always prefetch one week ahead of current position
- Silent failure on prefetch errors — user loads on demand when they navigate there

### No-data screen design
- Banner text: "End of history" with the date when archive data started
- Show navigation hints: "Press → or 0 to return"
- Box/border around message — emphasized style, draws attention like a modal
- Trigger: Show immediately when navigating into a week that has no data (don't fetch first)

### Transition behavior
- Non-prefetched weeks: Full loading screen (existing behavior), spinner, then display
- Rapid navigation: Jump to latest target, cancel intermediate fetches
- Instant navigation (prefetched): No feedback — data just appears
- No prefetch indicator in status bar — silent background work
- Today key (0/Home) from end-of-history: Jump directly to today with live data

### Claude's Discretion
- Exact banner styling (colors, borders) within theme system
- Cache management for prefetched data
- Error retry strategy for background prefetch
- Animation timing for transitions

</decisions>

<specifics>
## Specific Ideas

- The current "oldest data" indicator in bottom left is too subtle — user doesn't notice they've hit the end
- Blank screen with centered banner makes it unmistakably clear: no more data exists beyond this point
- Instant navigation for the previous week makes the app feel snappy — users most commonly go back one week

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 17-background-prefetch-and-improved-no-data-ux*
*Context gathered: 2026-02-25*
