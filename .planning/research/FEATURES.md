# Feature Research

**Domain:** Terminal dashboard TUI with theming and comprehensive stats panels
**Researched:** 2026-02-19 (Updated for v2.1 Visual Overhaul + Themes milestone)
**Confidence:** HIGH

## Research Scope: v2.1 Features Only

This research focuses on NEW features for v2.1 Visual Overhaul + Themes milestone:
- Additional stats panels (Categories, Editors, OS, Machines)
- Stats summary panel
- Theme system (6 popular presets)
- Theme selection mechanism

**Existing features in wakadash v2.0:** Languages panel, Projects panel, Sparkline, Heatmap, Panel toggles (1-4 keys), Auto-refresh, Help overlay

---

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Multiple stat panels** | All competitors (WTF, Sampler) show multiple data categories | LOW | Already have Languages/Projects, need Categories/Editors/OS/Machines |
| **Top 10 lists** | Industry standard for stats dashboards, WakaTime web dashboard shows this | LOW | Currently showing top 5, expand to top 10 with horizontal bars |
| **Summary metrics panel** | Dashboard users expect at-a-glance totals (Grafana, WakaTime patterns) | LOW | Total time, Daily avg, Top items, counts — already have some data |
| **Horizontal bar charts** | Universal pattern for ranked data visualization | LOW | Already using ntcharts barchart, just need more instances |
| **Time labels on bars** | Users need to see actual durations, not just relative lengths | LOW | Format as "5h 23m" next to each bar |
| **Color coding** | Terminal users expect visual distinction between categories | LOW | Already doing for languages, extend to other panels |
| **Responsive layout** | TUI must adapt to terminal size changes | MEDIUM | Already handling WindowSizeMsg, need 2-column + fallback layouts |
| **Panel toggles** | Users need to hide panels in small terminals or for focus | LOW | Already have 1-4 keys for toggle, extend to new panels (5-9) |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Named theme presets** | Familiar themes (Dracula, Nord, Gruvbox, etc.) reduce cognitive load vs custom colors | MEDIUM | Popular themes have cult followings, instant recognition |
| **Theme flag + config** | Flexibility — quick CLI flag OR persistent config file | LOW | Standard pattern: `--theme dracula` or config.yaml entry |
| **Automatic theme persistence** | Selected theme remembered across sessions | LOW | Save to config file on selection, UX win |
| **Smart 2-column fallback** | Panels arrange intelligently based on terminal width | MEDIUM | <80 cols = stack vertical, ≥80 = 2 columns, <40 = minimal |
| **Visual stats summary** | At-a-glance panel showing Last 30d, Totals, Averages, Top X | LOW | Unique value: everything important in one glance |
| **Activity heatmap integration** | GitHub-style heatmap already implemented, extend theming to it | LOW | Already built, just apply theme colors |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **Custom theme editor in TUI** | "Let me customize colors live!" | Complex UI, error-prone, bloats codebase | Config file editing + hot reload is simpler |
| **Per-panel color customization** | "Different colors for each stat type" | Overwhelming customization, decision paralysis | Named theme presets handle this better |
| **Animated theme transitions** | "Smooth color fades look cool" | Terminal rendering limitations, flicker, poor UX | Instant theme switch is cleaner |
| **Real-time theme sync across instances** | "Match all my terminals" | Requires file watching, race conditions, complexity | Manual sync via shared config file |
| **Auto-theme by time of day** | "Dark theme at night, light during day" | Assumes user preferences, hard to debug, annoying | Explicit user control via flag/config |
| **100+ theme pack** | "More themes = better product" | Maintenance burden, decision paralysis, unused | 6 curated popular themes is optimal |
| **Gradient/RGB terminal colors** | "Use full RGB color space" | Terminal compatibility issues, poor accessibility | Stick to 256-color palette for reliability |
| **Infinite scrolling panels** | "Show all 100 projects" | Breaks dashboard at-a-glance value, UX clutter | Top 10 with clear cutoff |

---

## Feature Dependencies

```
[Stats Panels] ──requires──> [API Data Structures]
                                  └──existing in types.StatsResponse

[Theme System] ──requires──> [Lipgloss Style Definitions]
                                  └──existing in styles.go

[Theme Selection] ──enhances──> [Config System]
                                      └──existing in config/config.go

[2-Column Layout] ──requires──> [WindowSizeMsg Handling]
                                      └──existing in model.go Update()

[Summary Panel] ──requires──> [Stats Panels Data]
                                   └──must fetch first

[Theme Presets] ──conflicts──> [Custom Theme Editor]
                                    └──choose one approach (we choose presets)

[Top 10 Lists] ──depends on──> [Bar Chart Rendering]
                                     └──existing ntcharts barchart
```

### Dependency Notes

- **Stats Panels require API Data**: Categories, Editors, OS, Machines already in WakaTime API response (types.StatsResponse), just need to render
- **Theme System requires Lipgloss**: Already using lipgloss for styles, just need to create theme structs that group colors
- **Theme Selection enhances Config**: Config system already exists, add `theme: "dracula"` field and load on startup
- **2-Column Layout requires WindowSizeMsg**: Already handling this in Update(), just need to calculate panel widths dynamically
- **Summary Panel requires Stats Data**: Must wait for API fetch, same pattern as existing panels
- **Theme Presets conflict with Custom Editor**: If we add custom theme editing, presets become less valuable. We choose presets.
- **Top 10 Lists depend on Bar Chart**: Already using ntcharts barchart for Languages/Projects, create more instances

---

## MVP Definition (v2.1 Scope)

### Launch With (v2.1 Visual Overhaul + Themes)

Minimum viable milestone — what's needed to achieve the stated goal.

- [x] **Categories panel** — Table stakes for comprehensive stats (Top 10, horizontal bars)
- [x] **Editors panel** — Table stakes (Top 10, horizontal bars)
- [x] **Operating Systems panel** — Table stakes (Top 10, horizontal bars)
- [x] **Machines panel** — Table stakes (Top 10, horizontal bars)
- [x] **Stats summary panel** — Differentiator (Last 30d total, Daily avg, Top project/editor/category/OS, counts)
- [x] **6 named theme presets** — Differentiator (Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night)
- [x] **Theme struct system** — Foundation (Group all lipgloss colors into Theme type)
- [x] **Theme flag** — Table stakes (`--theme dracula` CLI flag)
- [x] **Theme config persistence** — Differentiator (Save selected theme to config.yaml, load on startup)
- [x] **2-column responsive layout** — Table stakes (Side-by-side panels on wide terminals, stack on narrow)
- [x] **Panel toggles for new panels** — Table stakes (Extend existing 1-4 keys to 5-9 for new panels)
- [x] **Time labels on all bars** — Table stakes (Show "5h 23m" next to each bar)

### Add After Validation (v2.2+)

Features to add once core theming + panels are working and validated.

- [ ] **Theme list command** — `wakadash themes` shows available themes with color preview
- [ ] **Theme hot reload** — Edit config file, theme updates without restart (file watcher)
- [ ] **Custom theme in config** — Define custom color palette in config.yaml
- [ ] **Export current theme** — Save runtime theme to config format
- [ ] **3-column layout** — For ultra-wide terminals (>120 cols)
- [ ] **Panel size preferences** — User-configurable panel heights in config

### Future Consideration (v3+)

Features to defer until theming system is mature and user feedback gathered.

- [ ] **Community theme repository** — GitHub repo of user-submitted themes
- [ ] **Theme preview mode** — Cycle through themes with arrow keys before applying
- [ ] **Light theme variants** — Light versions of popular themes (Solarized Light, etc.)
- [ ] **Accessibility theme** — High-contrast WCAG-compliant theme for vision accessibility
- [ ] **Terminal background detection** — Auto-select theme based on terminal background (light/dark)

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Categories/Editors/OS/Machines panels | HIGH | LOW | P1 |
| Stats summary panel | HIGH | LOW | P1 |
| Theme presets (6 themes) | HIGH | MEDIUM | P1 |
| Theme flag | HIGH | LOW | P1 |
| Theme config persistence | MEDIUM | LOW | P1 |
| 2-column responsive layout | HIGH | MEDIUM | P1 |
| Time labels on bars | HIGH | LOW | P1 |
| Panel toggles (extend existing) | MEDIUM | LOW | P1 |
| Theme list command | MEDIUM | LOW | P2 |
| Theme hot reload | MEDIUM | MEDIUM | P2 |
| Custom theme in config | MEDIUM | MEDIUM | P2 |
| Export current theme | LOW | MEDIUM | P2 |
| 3-column layout | LOW | MEDIUM | P2 |
| Panel size preferences | LOW | MEDIUM | P3 |
| Community theme repository | LOW | HIGH | P3 |
| Theme preview mode | MEDIUM | HIGH | P3 |
| Light theme variants | MEDIUM | MEDIUM | P3 |
| Accessibility theme | HIGH | MEDIUM | P3 |
| Terminal background detection | LOW | HIGH | P3 |

**Priority key:**
- P1: Must have for v2.1 launch (Visual Overhaul + Themes milestone)
- P2: Should have for v2.2 (polish and flexibility)
- P3: Nice to have for v3+ (future enhancements)

---

## Implementation Patterns from Research

### Theme System Architecture (from OpenCode TUI, Process Compose)

**Pattern: Centralized theme struct**
```go
type Theme struct {
    Name            string
    BorderColor     lipgloss.Color
    TitleColor      lipgloss.Color
    DimColor        lipgloss.Color
    ErrorColor      lipgloss.Color
    WarningColor    lipgloss.Color
    AccentColor     lipgloss.Color
    // Activity heatmap colors (5 levels)
    HeatmapColors   [5]lipgloss.Color
    // Language-specific colors (fallback to accent)
    LanguageColors  map[string]lipgloss.Color
}
```

**Pattern: Theme registry**
```go
var themes = map[string]Theme{
    "dracula":      draculaTheme(),
    "nord":         nordTheme(),
    "gruvbox":      gruvboxTheme(),
    "monokai":      monokaiTheme(),
    "solarized":    solarizedTheme(),
    "tokyo-night":  tokyoNightTheme(),
}
```

**Pattern: Theme loading**
```go
func LoadTheme(name string) (Theme, error) {
    theme, ok := themes[name]
    if !ok {
        return themes["dracula"], fmt.Errorf("theme %q not found, using dracula", name)
    }
    return theme, nil
}
```

### Stats Panel Layout (from WTF, Grafana dashboards)

**Pattern: Summary panel structure**
```
┌─ Stats Summary (Last 30 Days) ──────────────┐
│ Total Time: 127h 34m                         │
│ Daily Avg:  4h 15m                           │
│                                              │
│ Top Project:  wakadash (45h 23m)             │
│ Top Editor:   VS Code (98h 12m)              │
│ Top Category: Coding (112h 45m)              │
│ Top OS:       macOS (127h 34m)               │
│                                              │
│ Projects: 12  Editors: 3  Languages: 8       │
└──────────────────────────────────────────────┘
```

**Pattern: 2-column layout with width detection**
```go
func (m Model) calculateLayout() (columns int, panelWidth int) {
    if m.width < 40 {
        return 1, m.width - 4  // Minimal, hide some panels
    } else if m.width < 80 {
        return 1, m.width - 4  // Stack all
    } else if m.width < 120 {
        return 2, (m.width / 2) - 6  // Side-by-side
    } else {
        return 3, (m.width / 3) - 6  // Optional: 3-column for ultra-wide
    }
}
```

### Panel Toggle Pattern (extend existing 1-4)

**Current: 1-4 keys for Languages/Projects/Sparkline/Heatmap**
**Extend: 5-8 keys for Categories/Editors/OS/Machines**
**New: 9 key for Summary panel**

Alternative: Use letter keys (c=categories, e=editors, o=os, m=machines, s=summary)

### Best Practices from BubbleTea Research

**Golden Rule #1: Always account for borders**
- Subtract 2 from height BEFORE rendering panels
- Already doing this in renderDashboard(): `panelHeight := m.height - lipgloss.Height(statusBar) - 4`

**Golden Rule #2: Never auto-wrap in bordered panels**
- Always truncate text explicitly
- Use lipgloss.Width() to measure, truncate strings that exceed panel width

**Project structure recommendation**
- ✅ Already following: model.go, commands.go, messages.go separation
- ✅ Already using: styles.go for lipgloss definitions
- ✅ Already using: types package for data structures
- New: Add `themes.go` for theme definitions and registry

---

## Theme Color Reference (from research)

### Dracula
- Background: #282a36
- Foreground: #f8f8f2
- Purple: #bd93f9
- Pink: #ff79c6
- Cyan: #8be9fd
- Green: #50fa7b
- Orange: #ffb86c
- Red: #ff5555
- Yellow: #f1fa8c

### Nord
- Background: #2e3440
- Foreground: #d8dee9
- Frost Blue: #81a1c1
- Frost Green: #88c0d0
- Aurora Green: #a3be8c
- Aurora Yellow: #ebcb8b
- Aurora Orange: #d08770
- Aurora Red: #bf616a
- Aurora Purple: #b48ead

### Gruvbox (Dark)
- Background: #282828
- Foreground: #ebdbb2
- Red: #cc241d / #fb4934
- Green: #98971a / #b8bb26
- Yellow: #d79921 / #fabd2f
- Blue: #458588 / #83a598
- Purple: #b16286 / #d3869b
- Aqua: #689d6a / #8ec07c
- Orange: #d65d0e / #fe8019

### Monokai
- Background: #272822
- Foreground: #f8f8f2
- Pink: #f92672
- Purple: #ae81ff
- Blue: #66d9ef
- Green: #a6e22e
- Orange: #fd971f
- Yellow: #e6db74
- Red: #f92672

### Solarized (Dark)
- Background: #002b36
- Foreground: #839496
- Base: #073642
- Yellow: #b58900
- Orange: #cb4b16
- Red: #dc322f
- Magenta: #d33682
- Violet: #6c71c4
- Blue: #268bd2
- Cyan: #2aa198
- Green: #859900

### Tokyo Night
- Background: #1a1b26
- Foreground: #c0caf5
- Blue: #7aa2f7
- Cyan: #7dcfff
- Green: #9ece6a
- Orange: #ff9e64
- Red: #f7768e
- Purple: #bb9af7
- Magenta: #c0caf5
- Yellow: #e0af68

---

## Sources

**Terminal Theme Systems:**
- [TUI Theming and Commands | OpenCode](https://deepwiki.com/sst/opencode/6.4-tui-theming-keybinds-and-commands) - Theme system architecture, JSON-based color schemes
- [Ratatui color scheme discussion](https://github.com/ratatui/ratatui/discussions/877) - Color palette best practices
- [Gogh Terminal Color Schemes](https://github.com/Gogh-Co/Gogh) - Collection of terminal themes
- [5 Stunning Tmux Themes](https://thefilibusterblog.com/5-stunning-tmux-themes-to-enhance-your-terminal-experience/) - Dracula, Nord, Kanagawa themes

**Dashboard Design Best Practices:**
- [9 Dashboard Design Principles (2026)](https://www.designrush.com/agency/ui-ux-design/dashboard/trends/dashboard-design-principles) - Visual hierarchy, S.M.A.R.T framework
- [WTF Terminal Dashboard](https://github.com/wtfutil/wtf) - Personal information dashboard patterns
- [Grafana Dashboard Best Practices](https://grafana.com/docs/grafana/latest/visualizations/dashboards/build-dashboards/best-practices/) - Documentation, reusability, avoiding clutter
- [Why Users Ignore Dashboards](https://www.eleken.co/blog-posts/why-users-ignore-dashboards) - Information overload anti-patterns

**TUI Framework Best Practices:**
- [Tips for Building BubbleTea Programs](https://leg100.github.io/en/posts/building-bubbletea-programs/) - Golden rules, project structure
- [BubbleTea vs Ratatui Comparison (2026)](https://www.glukhov.org/post/2026/02/tui-frameworks-bubbletea-go-vs-ratatui-rust/) - Framework comparison, when to use which
- [Building TUI with BubbleTea](https://themarkokovacevic.com/posts/terminal-ui-with-bubbletea/) - Architecture patterns

**Theme Configuration:**
- [Process Compose TUI](https://f1bonacc1.github.io/process-compose/tui/) - Theme selection with CTRL-T, settings.yaml config
- [Nx Terminal UI](https://nx.dev/docs/guides/tasks--caching/terminal-ui) - --tui flag pattern
- [Kubetui Configuration](https://github.com/sarub0b0/kubetui) - --config-file flag pattern

**WakaTime Dashboard Patterns:**
- [WakaTime Grafana Dashboard](https://grafana.com/grafana/dashboards/12790-wakatime-coding-stats/) - Metrics panel layout
- [WakaTime Official](https://wakatime.com/) - Dashboard metrics and quantified-self patterns

**Data Visualization:**
- [34 Top Chart Types (2026)](https://www.luzmo.com/blog/chart-types) - Horizontal bar chart best practices
- [Dashboard Anti-Patterns](https://medium.com/design-bootcamp/why-messy-dashboards-kill-user-experience-and-how-i-fixed-it-79275da8b2fe) - Clutter, overload, lack of hierarchy

**Popular Themes:**
- [Dracula Theme Official](https://draculatheme.com) - Official Dracula color scheme
- [Slant: Best VIM Color Schemes (2026)](https://www.slant.co/topics/480/~best-vim-color-schemes) - Gruvbox, Molokai, Dracula rankings
- [iTerm Color Schemes](https://iterm2colorschemes.com/) - 325+ terminal themes including Nord, Solarized, Tokyo Night

---
*Feature research for: wakadash v2.1 Visual Overhaul + Themes*
*Researched: 2026-02-19*
