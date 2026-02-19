# Project Research Summary

**Project:** wakadash v2.1 — Visual Overhaul + Themes
**Domain:** Terminal dashboard TUI with theme system and comprehensive stats panels
**Researched:** 2026-02-19
**Confidence:** HIGH

## Executive Summary

wakadash v2.1 extends the existing live WakaTime dashboard (built with Bubble Tea + Lipgloss) with a professional theme system and comprehensive stats panels. The research indicates this is a **low-risk enhancement** to an established foundation — not building from scratch, but systematically adding features to working code.

The recommended approach uses **centralized theme abstraction** via a Theme struct with 6 pre-built presets (Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night) sourced from github.com/willyv3/gogh-themes/lipgloss v1.2.0. Stats panels (Categories, Editors, OS, Machines) reuse the existing ntcharts barchart pattern established for Languages/Projects panels — zero new chart dependencies needed. Theme selection persists via ~/.wakatime.cfg extension with simple key=value format.

The **critical risks** are integration pitfalls from mixing new and old code: hardcoded colors breaking the theme system, Lipgloss AdaptiveColor causing startup hangs (fixed in BubbleTea v0.27.1+), and dynamic .Width() styles triggering rendering corruption. These are all preventable with systematic migration: audit hardcoded colors first, force early terminal detection, use .MaxWidth() instead of .Width() for dynamic content. The architecture research provides clear patterns: create theme struct before migration, compute summary stats in Update (not View), calculate responsive layouts in WindowSizeMsg handler.

## Key Findings

### Recommended Stack

The stack additions for v2.1 are minimal — **one new dependency** plus systematic refactoring of existing code. The existing v2.0 foundation (Bubble Tea v1.3.10, Lipgloss v1.1.0, ntcharts v0.4.0) remains unchanged and stable. No framework upgrades needed (v2 versions still in beta).

**Add one dependency:**
- **github.com/willyv3/gogh-themes/lipgloss v1.2.0** — 361 professional terminal color schemes including all 6 required themes (Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night). Colors pre-wrapped as lipgloss.Color, zero manual conversion needed. Released Oct 2025, zero dependencies, themes compiled into binary.

**No updates needed:**
- charmbracelet/bubbletea v1.3.10 (stable, v2 still beta)
- charmbracelet/lipgloss v1.1.0 (stable, v2 still alpha)
- charmbracelet/bubbles v1.0.0 (stable, v2 still beta)
- NimbleMarkets/ntcharts v0.4.0 (latest stable)

**Config mechanism:**
- Standard library `flag` package for CLI override (--theme dracula)
- Extend existing ~/.wakatime.cfg parser with `theme=` key
- Zero heavy config libraries needed (spf13/viper, gookit/config rejected as overkill)

**Stats panels implementation:**
- API data already fetched: Categories, Editors, OperatingSystems, Machines exist in types.StatsData struct
- Rendering uses existing ntcharts barchart.Model (same as Languages/Projects panels)
- Summary stats use existing lipgloss layout (no table libraries needed)

### Expected Features

Research indicates **8 must-have features** (table stakes users expect) and **6 differentiators** (competitive advantages). The anti-features section identifies 8 commonly requested but problematic features to explicitly avoid.

**Must have (table stakes):**
- Multiple stat panels — All competitors (WTF, Sampler) show multiple data categories, not just Languages/Projects
- Top 10 lists — Industry standard for stats dashboards, WakaTime web dashboard shows this
- Summary metrics panel — Dashboard users expect at-a-glance totals (Grafana, WakaTime patterns)
- Time labels on bars — Users need to see actual durations ("5h 23m"), not just relative lengths
- Color coding — Terminal users expect visual distinction between categories
- Responsive layout — TUI must adapt to terminal size changes
- Panel toggles — Users need to hide panels in small terminals or for focus
- Horizontal bar charts — Universal pattern for ranked data visualization

**Should have (competitive advantages):**
- Named theme presets — Familiar themes (Dracula, Nord, Gruvbox, etc.) reduce cognitive load vs custom colors, cult followings
- Theme flag + config — Flexibility via --theme dracula CLI flag OR persistent config file
- Automatic theme persistence — Selected theme remembered across sessions
- Smart 2-column fallback — Panels arrange intelligently: <80 cols = stack vertical, ≥80 = 2 columns
- Visual stats summary — At-a-glance panel showing Last 30d, Totals, Averages, Top items
- Activity heatmap integration — GitHub-style heatmap already implemented, extend theming to it

**Defer (explicitly avoid — anti-features):**
- Custom theme editor in TUI — Complex UI, error-prone, bloats codebase. Config file editing simpler.
- Animated theme transitions — Terminal rendering limitations, flicker, poor UX. Instant switch cleaner.
- Auto-theme by time of day — Assumes user preferences, hard to debug, annoying. Explicit control better.
- 100+ theme pack — Maintenance burden, decision paralysis. 6 curated popular themes optimal.
- Gradient/RGB terminal colors — Terminal compatibility issues, poor accessibility. Stick to 256-color.
- Infinite scrolling panels — Breaks dashboard at-a-glance value. Top 10 with clear cutoff better.

### Architecture Approach

The architecture builds on Bubble Tea's Model-View-Update pattern with **centralized theme abstraction** and **modular panel components**. Key decision: **style registry pattern** with static theme presets (not runtime-switchable) to avoid state management complexity.

**Theme system architecture:**
1. **Centralized Theme struct** — Groups all lipgloss styles (Border, Title, Dim, Error, Warning, chart colors)
2. **Adaptive colors for UI, fixed for charts** — UI elements use lipgloss.AdaptiveColor for light/dark terminal detection; chart colors use fixed palette (assume dark terminal)
3. **Static theme presets** — Themes selected at startup via config/flag, immutable during session (avoids runtime switching complexity)
4. **Config extension** — Add `theme=dracula` to existing ~/.wakatime.cfg (simple key=value format)
5. **No runtime switching** — Avoid theme mutation during session (causes performance lag, state complexity)

**Major components:**
1. **Theme Registry** (internal/tui/theme.go) — Theme struct definition, GetTheme() function, preset registry
2. **Individual Theme Definitions** (internal/tui/themes_*.go) — One file per theme (Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night)
3. **Stats Panel Components** — Reuse existing barchart.Model pattern for Categories, Editors, OS, Machines
4. **Summary Stats Computation** — SummaryStats struct computed in Update (not View), maintains pure render function
5. **Responsive Layout** — Grid-based layout adapts to terminal width: 3-column (≥120 cols), 2-column (80-119 cols), 1-column (<80 cols)

**Integration points:**
- `config.go` — Add Theme string field, parse `theme=` key from ~/.wakatime.cfg
- `main.go` — Add --theme flag, load theme via GetTheme(), pass to Model
- `model.go` — Add theme field, new chart instances, summaryStats, visibility toggles (extend 1-4 keys to 5-9)
- `styles.go` — DELETE (move to theme.go for centralization)

**Data flow:**
```
main.go → config.Load() → (APIURL, APIKey, Theme)
       → tui.GetTheme(cfg.Theme) → Theme struct
       → tui.NewModel(client, range, refresh, theme)
       → Model with theme-based styles
```

### Critical Pitfalls

Research identified **8 critical pitfalls** and **4 moderate pitfalls** with specific prevention strategies. The critical ones are all integration issues from adding features to existing code.

**Top 5 critical pitfalls to avoid:**

1. **Hardcoded Colors Break Theme System Integration** — Existing wakadash code has hardcoded ANSI colors (Color "62", "205", "241", "196", "214") in styles.go that will conflict with themes. **Prevention:** Audit all lipgloss.Color() calls with grep before implementing themes, create Theme struct first, migrate incrementally, use semantic names (theme.AccentColor not theme.Color62).

2. **AdaptiveColor Terminal Queries Cause Startup Hangs** — Using lipgloss.AdaptiveColor can cause 3-5 second hangs or indefinite freezes due to Bubble Tea and Lipgloss racing for stdout during terminal background detection. **Prevention:** Call `_ = lipgloss.HasDarkBackground()` in main() BEFORE program.Run(), ensure BubbleTea v0.27.1+ (bug fixed), provide --theme-mode=dark|light flag as fallback.

3. **Dynamic Width Styles Cause Rendering Corruption** — Using .Width() on dynamically changing content causes text to render faint/ghosted on subsequent frames, borders bleed between panels. **Prevention:** Use .MaxWidth() on containers (not .Width() on dynamic text), explicit truncation for changing values, test rapid rerenders over 5+ minutes.

4. **API Rate Limiting Triggers with Multiple Panel Data Fetches** — Naive implementation with 6 separate API requests per panel hits WakaTime's <10 req/sec limit within 5 minutes. **Prevention:** Single /stats API request that returns all data, distribute to panels. API already returns Categories, Editors, OS, Machines in one response.

5. **Border Calculations Break Multi-Panel Layouts** — Forgetting to subtract border characters causes panels to overflow. 6 panels × 2 chars each = 12 characters miscalculation. **Prevention:** Always calculate `contentHeight = total - 2` before rendering, create layout calculator function, test at 80×24 terminal size.

**Other critical pitfalls:**
- **Viewport Memory Leaks** — Pre-v0.21.1 Bubbles creates new 4KB parser objects per render. With 6+ panels updating every second, compounds rapidly. **Prevention:** Update to Bubbles v0.21.1+ with parser pooling.
- **TERM Variable Incompatibility** — Dashboard looks perfect in dev terminal but broken colors in users' terminals (xterm vs xterm-256color vs tmux). **Prevention:** Test across TERM values, use AdaptiveColor for UI, verify 256-color degradation looks acceptable.
- **Runtime Theme Switching Lag** — Changing themes requires rebuilding all styles, causes 100-500ms lag spikes. **Prevention:** Use restart-based themes (simpler, zero runtime cost) OR implement lazy/progressive redraw if runtime switching critical.

## Implications for Roadmap

Based on research, recommended **3-phase structure** with clear boundaries and dependencies.

### Phase 1: Theme Foundation

**Rationale:** Must establish theme system foundation before adding new panels. Migrating existing 4 panels to themes validates the pattern before scaling to 10+ panels. Critical pitfalls (hardcoded colors, AdaptiveColor hangs, TERM compatibility) must be addressed before expanding UI surface area.

**Delivers:**
- Theme struct with semantic color fields
- 6 theme presets (Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night)
- Config extension (theme= key in ~/.wakatime.cfg)
- CLI flag override (--theme dracula)
- Migration of existing panels to theme-based styling
- TERM compatibility verified (test on 6+ terminal types)

**Addresses features:**
- Named theme presets (differentiator)
- Theme flag + config (differentiator)
- Automatic theme persistence (differentiator)

**Avoids pitfalls:**
- Hardcoded colors breaking theme system (audit with grep, incremental migration)
- AdaptiveColor startup hangs (call HasDarkBackground() in main() first)
- TERM incompatibility (test xterm, xterm-256color, tmux, screen, alacritty, kitty)
- Style creation in View() (create once in Init/Update, not every render)

**Technical debt paid:**
- Delete styles.go (5 global variables with hardcoded colors)
- Centralize all styling in theme.go
- Establish semantic color naming convention

### Phase 2: Stats Panels + Summary

**Rationale:** With theme foundation established, add new panels using proven patterns. API data already fetched (Categories, Editors, OS, Machines in StatsData), just need rendering. Summary panel requires all stats panels to exist first (aggregates top items from each).

**Delivers:**
- 4 new stat panels (Categories, Editors, Operating Systems, Machines) using existing barchart.Model pattern
- Stats summary panel (Last 30d, Daily avg, Top items, counts)
- Responsive 2-column layout (side-by-side on ≥80 cols, stack on <80)
- Panel toggles extended (keys 1-4 to 5-9)
- Time labels on all bars ("5h 23m" formatting)
- Top 10 lists for all panels

**Addresses features:**
- Multiple stat panels (table stakes)
- Top 10 lists (table stakes)
- Summary metrics panel (table stakes)
- Time labels on bars (table stakes)
- Responsive layout (table stakes)
- Panel toggles (table stakes)
- Visual stats summary (differentiator)

**Uses stack:**
- Existing ntcharts barchart.Model (no new chart libraries)
- Existing lipgloss layout (no table libraries)
- Existing types.StatsData (API already fetches all data)

**Implements architecture:**
- SummaryStats struct computed in Update (pure View functions)
- Responsive layout helpers (calculatePanelWidth() based on terminal size)
- Unified panel renderer (renderStatsPanel() pattern)

**Avoids pitfalls:**
- API rate limiting (single /stats request, distribute data to panels)
- Border calculation errors (contentHeight = total - 2, test at 80×24)
- Dynamic width corruption (use .MaxWidth() on containers, explicit truncation)
- Viewport memory leaks (update to Bubbles v0.21.1+, limit viewport usage)
- All panels updating simultaneously (consider staggered updates 5s apart)

### Phase 3: Polish + Edge Cases

**Rationale:** With core functionality complete, address edge cases and polish. Minimum terminal size check prevents confusing broken layouts. Theme validation ensures graceful degradation on invalid theme names.

**Delivers:**
- Minimum terminal size enforcement (100×30 recommended based on panel layout)
- Theme validation with fallback (invalid theme → default to Dracula)
- Error message polish (rate limiting, API failures, terminal too small)
- 3-column layout for ultra-wide terminals (optional, if ≥120 cols common)
- Panel size optimization based on actual usage patterns

**Addresses features:**
- Smart 2-column fallback (differentiator) — extends to 3-column
- Activity heatmap integration (differentiator) — apply theme colors to existing heatmap

**Avoids pitfalls:**
- No minimum terminal size check (friendly error at 80×24)
- Theme config and API key in same file (acceptable for v2.1, separate in v2.2+)
- All panels update simultaneously (implement staggered updates)

### Phase Ordering Rationale

**Why this order:**
1. **Theme foundation must come first** — Can't add 6+ new panels with hardcoded colors, migration becomes impossible. Validating theme system with existing 4 panels (Languages, Projects, Sparkline, Heatmap) before scaling to 10+ panels reduces risk.

2. **Stats panels build on proven patterns** — Theme system established, ntcharts barchart.Model pattern proven with Languages/Projects, API data already fetched. Just rendering, no new dependencies or architecture patterns.

3. **Polish after validation** — Can't determine minimum terminal size until panel layout finalized. Edge case handling (invalid themes, terminal too small) makes sense after core functionality working.

**Grouping based on dependencies:**
- Phase 1: Zero dependencies on new panels (works with existing 4 panels)
- Phase 2: Depends on theme system existing (new panels use theme colors)
- Phase 3: Depends on panel layout finalized (need actual dimensions to set minimums)

**How this avoids pitfalls:**
- Hardcoded color audit happens in Phase 1 before scaling UI surface area
- TERM compatibility testing in Phase 1 with smaller UI validates before expansion
- API rate limiting addressed in Phase 2 when adding data-dependent panels
- Memory profiling in Phase 2 after implementing 2-3 panels, before building all 6+
- Border calculation patterns established in Phase 2, applied consistently

### Research Flags

**Phases likely needing deeper research during planning:**
- **Phase 2 (Stats Panels):** Responsive layout calculations for 2-column → 3-column grid. Research indicates patterns exist (lipgloss.JoinHorizontal, calculatePanelWidth()) but may need experimentation for optimal panel sizing at different terminal widths.

**Phases with standard patterns (skip research-phase):**
- **Phase 1 (Theme Foundation):** Well-documented theme struct pattern from charmbracelet/huh theme.go and purpleclay/lipgloss-theme. Config extension follows existing ~/.wakatime.cfg parser pattern.
- **Phase 3 (Polish):** Standard error handling and validation patterns.

**Additional research recommendations:**
- **Before Phase 1:** Audit current codebase for all lipgloss.Color() usage (grep -r "lipgloss.Color" .) to understand hardcoded color scope
- **During Phase 2:** Profile memory usage after implementing 2-3 panels before building all 6+ (verify <10 MB baseline, stable over 30 minutes)
- **After Phase 2:** Visual regression testing with all 6 themes across 6 terminal types (xterm, xterm-256color, tmux, screen, alacritty, kitty)

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Single new dependency (gogh-themes/lipgloss v1.2.0) with official pkg.go.dev docs, Oct 2025 release. Existing dependencies stable (v1.x releases), v2 betas explicitly avoided. Config via standard library flag package. |
| Features | HIGH | Feature research cross-referenced WTF, Sampler, Grafana, WakaTime web dashboard patterns. Table stakes vs differentiators vs anti-features clearly categorized. MVP definition (v2.1 scope) extracted from existing FEATURES.md milestone. |
| Architecture | HIGH | Theme struct pattern verified from charmbracelet/huh theme.go and purpleclay/lipgloss-theme official implementations. Bubble Tea Model-View-Update pattern well-documented. Integration points clearly defined with specific file changes needed. |
| Pitfalls | HIGH | Critical pitfalls sourced from official GitHub issues with maintainer responses (BubbleTea #1036 AdaptiveColor hang, BubbleTea #1225 Width() corruption, Bubbles #829 viewport memory). Terminal compatibility from comprehensive TERM variable analysis. WakaTime rate limiting from official API docs. |

**Overall confidence:** HIGH

All four research areas backed by primary sources (official documentation, verified GitHub issues with maintainer confirmations, established codebase patterns from wakadash v2.0). Architecture approach validated by multiple BubbleTea applications (charmbracelet/huh, purpleclay/lipgloss-theme, OpenCode TUI). Pitfalls extracted from actual bug reports with confirmed fixes in specific versions.

### Gaps to Address

Minor gaps that need validation during implementation:

**Gap 1: Optimal panel sizing for responsive layout**
- **Issue:** Research indicates patterns (2-column at ≥80 cols, 3-column at ≥120 cols) but actual optimal widths depend on content. Top 10 lists with time labels need minimum width for readability.
- **Handling:** Implement initial layout with 2-column focus (80-119 cols), measure actual content widths with longest language/project/category names. 3-column layout (≥120 cols) can be added in Phase 3 if testing shows benefit.
- **Validation:** Test with real WakaTime data (not mock data) at 80, 100, 120, 160 cols to verify panel widths work.

**Gap 2: GitHub Linguist color degradation in 256-color mode**
- **Issue:** Existing wakadash uses GitHub Linguist hex codes for language colors. Research indicates these should degrade gracefully in 256-color terminals, but actual degradation quality unknown.
- **Handling:** Test language panel rendering in terminals with TERM=xterm-256color (not true color). If colors indistinguishable or jarring, implement 256-color palette mapping for common languages.
- **Validation:** Visual test of top 10 languages in 256-color mode. Colors must be distinguishable.

**Gap 3: Summary stats "Last 30 Days" calculation**
- **Issue:** Architecture research recommends using existing stats.HumanReadableTotal if range is last_30_days, but unclear if WakaTime API always provides 30-day summary or if separate endpoint fetch needed.
- **Handling:** If current range matches, use TotalTime directly. Otherwise, show "Range: [actual range]" instead of "Last 30 Days" label. Don't fetch separate summaries endpoint (avoids rate limiting complexity).
- **Validation:** Test with --range last_30_days vs --range last_7_days to verify label accuracy.

**Gap 4: Heatmap theme color integration**
- **Issue:** Features research suggests theme-specific heatmap gradients (Dracula = purple gradient, Nord = blue gradient) but unclear if users prefer consistent GitHub-style green gradient across themes.
- **Handling:** Phase 1 implements universal GitHub-style green gradient for all themes (simpler, familiar). Phase 3 can add theme-specific gradients if user feedback requests it.
- **Validation:** User testing with 2-3 beta testers — ask "Would you prefer heatmap colors match theme, or stay GitHub green?"

**Non-gaps (research conclusive):**
- Theme struct pattern — Multiple verified implementations (charmbracelet/huh, purpleclay/lipgloss-theme)
- API rate limiting prevention — Official WakaTime docs confirm single /stats endpoint returns all data
- AdaptiveColor hang fix — BubbleTea v0.27.1+ confirmed fix with maintainer response
- Viewport memory leak fix — Bubbles v0.21.1+ confirmed fix with parser pooling
- Border calculation pattern — Golden rule documented in BubbleTea best practices

## Sources

### Primary (HIGH confidence)

**Official Documentation:**
- [gogh-themes/lipgloss pkg.go.dev](https://pkg.go.dev/github.com/willyv3/gogh-themes/lipgloss) — v1.2.0 API, Oct 2025 release, 361 themes verified
- [WakaTime API Docs](https://wakatime.com/developers) — Rate limiting, /stats endpoint structure
- [charmbracelet/lipgloss GitHub](https://github.com/charmbracelet/lipgloss) — v1.1.0, AdaptiveColor usage
- [charmbracelet/bubbletea GitHub](https://github.com/charmbracelet/bubbletea) — v1.3.10, Model-View-Update pattern
- [NimbleMarkets/ntcharts](https://github.com/NimbleMarkets/ntcharts) — v0.4.0, barchart API

**GitHub Issues (Confirmed Bugs/Fixes):**
- [BubbleTea #1036](https://github.com/charmbracelet/bubbletea/issues/1036) — AdaptiveColor startup hang, FIXED v0.27.1, maintainer response
- [BubbleTea #1225](https://github.com/charmbracelet/bubbletea/issues/1225) — Width() rendering corruption, detailed reproduction steps
- [Bubbles #829](https://github.com/charmbracelet/bubbles/issues/829) — Viewport memory leak, FIXED v0.21.1, pprof analysis

**Reference Implementations:**
- [charmbracelet/huh theme.go](https://github.com/charmbracelet/huh/blob/main/theme.go) — Theme struct pattern, adaptive colors
- [purpleclay/lipgloss-theme](https://pkg.go.dev/github.com/purpleclay/lipgloss-theme) — Theme palette example

### Secondary (MEDIUM confidence)

**Theme System Patterns:**
- [Design Tokens & Theming Guide](https://materialui.co/blog/design-tokens-and-theming-scalable-ui-2025) — Migration from hardcoded values
- [Material-UI #25018](https://github.com/mui/material-ui/issues/25018) — Theme switching performance issues (web framework, analogous to TUI)

**Terminal Compatibility:**
- [Terminal Colours Are Tricky](https://jvns.ca/blog/2024/10/01/terminal-colours/) — Comprehensive TERM variable explanation
- [Ratatui Color Discussion](https://github.com/ratatui/ratatui/discussions/877) — Color palette best practices

**TUI Best Practices:**
- [Tips for Building BubbleTea Programs](https://leg100.github.io/en/posts/building-bubbletea-programs/) — Layout calculations, border handling, golden rules
- [Shifoo: Multi View Interfaces](https://shi.foo/weblog/multi-view-interfaces-in-bubble-tea) — Component composition patterns

**Dashboard Design:**
- [9 Dashboard Design Principles (2026)](https://www.designrush.com/agency/ui-ux-design/dashboard/trends/dashboard-design-principles) — Visual hierarchy, S.M.A.R.T framework
- [Grafana Dashboard Best Practices](https://grafana.com/docs/grafana/latest/visualizations/dashboards/build-dashboards/best-practices/) — Multi-panel complexity management

**Theme Color Specifications:**
- [Dracula Theme Official](https://draculatheme.com/spec) — Official Dracula color specification
- [Nord Theme Colors and Palettes](https://www.nordtheme.com/docs/colors-and-palettes/) — Nord color system documentation
- [Gruvbox GitHub](https://github.com/morhetz/gruvbox) — Gruvbox color scheme
- [Tokyo Night GitHub](https://github.com/folke/tokyonight.nvim) — Tokyo Night color definitions

### Tertiary (Project-Specific Context)

**Existing Codebase:**
- `/workspace/wakadash/internal/tui/styles.go` — Current hardcoded colors (Color "62", "205", "241", "196", "214")
- `/workspace/wakadash/internal/tui/colors.go` — Language colors using GitHub Linguist hex codes
- `/workspace/wakadash/internal/types/types.go` — StatsData struct confirming Categories, Editors, OS, Machines available
- `/workspace/wakadash/internal/config/config.go` — Existing ~/.wakatime.cfg parser (api_url, api_key)

---

*Research completed: 2026-02-19*
*Ready for roadmap: yes*
