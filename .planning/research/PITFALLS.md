# Pitfalls Research: v2.1 Visual Overhaul + Themes

**Domain:** Adding theme system and 6+ stats panels to existing wakadash TUI
**Researched:** 2026-02-19
**Confidence:** HIGH for Lipgloss integration issues (GitHub issues, maintainer responses), MEDIUM for terminal compatibility (community sources), HIGH for WakaTime API (official docs)

**Focus:** Integration pitfalls when adding features to existing system, not building from scratch.

---

## Critical Pitfalls

### Pitfall 1: Hardcoded Colors Break Theme System Integration

**What goes wrong:**
Adding a theme system to code with hardcoded colors (like `lipgloss.Color("62")`, `lipgloss.Color("205")`) creates a maintenance nightmare. Every theme requires manually finding and updating dozens of scattered color values. Missed hardcoded colors cause visual inconsistencies where some elements theme correctly while others remain stuck in the old color scheme.

**Why it happens:**
Developers start with simple hardcoded values during prototyping, intending to refactor later. When adding themes becomes a requirement, they underestimate the scope of refactoring needed. The existing wakadash codebase has hardcoded ANSI color codes in `internal/tui/styles.go` (Color "62", "205", "241", "196", "214") that will conflict with any theme system.

**Consequences:**
- Every theme preset requires manual find/replace of color codes across multiple files
- Visual inconsistencies when 80% of UI themes correctly but 20% uses old hardcoded colors
- Can't add new themes without full codebase audit each time
- Regression where adding new UI element with hardcoded color breaks existing theme

**How to avoid:**
1. **Audit before implementing themes** - Use `grep -r "lipgloss.Color" .` to find all hardcoded colors
2. **Create theme struct first** - Define `type Theme struct { Primary, Secondary, Accent, Error, Warning lipgloss.Color }` before migration
3. **Incremental migration strategy** - Convert one component at a time, verify visually after each step
4. **Use semantic names** - `theme.AccentColor` not `theme.Color62` - semantic names survive theme changes
5. **Add build-time verification** - Consider linting rules that flag new hardcoded color values in code review

**Warning signs:**
- Finding color codes during `git grep` that aren't in your theme definition
- Some UI elements don't change when switching themes
- Needing to update multiple files to change a single semantic color
- Code review shows `lipgloss.Color("...")` with literal values

**Phase to address:**
Phase 1 (Theme Foundation) - Must complete hardcoded color audit and migration before adding new panels

**References:**
- [Design Tokens & Theming Guide](https://materialui.co/blog/design-tokens-and-theming-scalable-ui-2025) - Migration from hardcoded values
- wakadash `/workspace/wakadash/internal/tui/styles.go` - Current hardcoded state

---

### Pitfall 2: AdaptiveColor Terminal Queries Cause Startup Hangs

**What goes wrong:**
Using `lipgloss.AdaptiveColor` for light/dark terminal detection can cause the application to hang for several seconds or indefinitely on startup. This happens because Lipgloss and BubbleTea both try to query the terminal simultaneously, creating a race condition when `termenv.HasDarkBackground()` attempts to detect terminal capabilities while the framework is initializing.

**Why it happens:**
The detection mechanism requires querying the terminal's background color through escape sequences. When this happens during BubbleTea initialization (typically in the second `View()` call after `tea.WindowSizeMsg`), both libraries compete for `stdout` access. As explained by a BubbleTea maintainer: "Bubble Tea and Lip Gloss are both jumping on, and fighting over, stdout at the same time."

**Consequences:**
- Application hangs 3-5 seconds on startup in affected terminals
- Some terminal types hang indefinitely, requiring Ctrl+C
- Inconsistent behavior - works on developer's terminal, hangs for users
- Impossible to debug without understanding the stdout race condition

**How to avoid:**
1. **Force early detection** - Call `_ = lipgloss.HasDarkBackground()` in `main()` BEFORE `program.Run()`
2. **Use BubbleTea v0.27.1+** - The bug was fixed in v0.27.1, ensure dependencies are up-to-date
3. **Test on actual terminals** - Some terminals handle queries better than others; test on iTerm2, Alacritty, GNOME Terminal, tmux
4. **Provide manual override** - Add `--theme-mode=dark|light` flag as fallback when auto-detection fails
5. **Timeout detection** - Wrap detection in timeout logic to fail gracefully rather than hang indefinitely

**Warning signs:**
- Application hangs for 3-5 seconds on startup
- Works fine on some terminals but hangs on others
- Delay happens specifically when first rendering AdaptiveColor styles
- Memory profiling shows blocking on `HasDarkBackground()` call
- First `View()` renders fine, second `View()` hangs

**Phase to address:**
Phase 1 (Theme Foundation) - Must verify AdaptiveColor initialization before implementing theme presets

**References:**
- [BubbleTea #1036: AdaptiveColor Hanging Bug](https://github.com/charmbracelet/bubbletea/issues/1036) - Official issue with maintainer explanation and fix
- Fixed in BubbleTea v0.27.1

---

### Pitfall 3: Dynamic Width Styles Cause Rendering Corruption on Rerenders

**What goes wrong:**
When using `.Width()` with dynamically changing content across multiple styled substrings, text renders incorrectly on subsequent frames. The initial render appears correct, but rerenders show faint/ghosted text, bleeding borders, and escape sequences affecting unintended screen areas. Text that should appear normal renders with the "faint" style applied, and borders from panels appear affected.

**Why it happens:**
The renderer unconditionally erases lines when cursor reaches line end, causing escape sequences from styled segments to persist and bleed into subsequent renders. According to the fix commit: "When the cursor reaches the end of the line, any escape sequences that follow will only affect the last cell of the line." When combining `.Width()` constraints with dynamic text that changes between frames (like updating stats panels), line wrap calculations become incorrect and affect rendering outside the intended boundaries.

**Consequences:**
- Text appears normal on first render, then becomes faint/ghosted on subsequent renders
- Borders from one panel bleed into adjacent panels
- Problem compounds with each rerender (accumulating escape sequences)
- Difficult to debug - looks like a styling bug, not a rendering engine issue
- Only manifests with dynamic content that changes values, not static text

**How to avoid:**
1. **Separate static and dynamic content** - Apply `.Width()` only to containers, not individual changing text elements
2. **Use explicit truncation** - `lipgloss.NewStyle().MaxWidth(50)` with manual truncation instead of `.Width()` for dynamic text
3. **Test rapid rerenders** - If stats update every second, verify rendering doesn't degrade over 5+ minutes
4. **Prefer fixed-width layouts** - Dashboard panels should have consistent widths, not dynamically sized based on content
5. **Clear before redraw** - For panels with `.Width()`, explicitly clear the area before rendering new content
6. **Update dependencies** - Ensure using latest x/ansi and lipgloss versions with rendering fixes

**Warning signs:**
- Text appears normal on first render, then becomes faint or ghosted
- Borders from one panel bleed into adjacent panels
- Problem worsens with each rerender (accumulating escape sequences)
- Only happens with dynamic content, not static text
- Disappears when removing `.Width()` constraints
- Bug manifests after 10-20 seconds of live updates, not immediately

**Phase to address:**
Phase 2 (Stats Panels) - Critical to address before implementing 6+ dynamic panels with live updates

**References:**
- [BubbleTea #1225: Width() Rendering Issue](https://github.com/charmbracelet/bubbletea/issues/1225) - Detailed bug report with reproduction steps
- Fix involved changing line erasure behavior in renderer

---

### Pitfall 4: Viewport Memory Leaks with Multiple Panels

**What goes wrong:**
Using BubbleTea viewports for multiple scrollable panels causes excessive memory consumption - applications using 20-40 MB for simple content that should consume <1 MB. Memory profiling reveals allocations come from `viewport.SetContent()` and underlying Lipgloss rendering, not from actual data size. With 6+ panels each using viewports, this compounds to 60+ MB baseline memory usage.

**Why it happens:**
Pre-v0.21.1 Bubbles versions didn't pool ANSI parser instances. Every viewport render created new 4KB parser objects without reuse, causing linear memory growth with panel count and render frequency. With 6+ panels updating every second, this compounds rapidly: 6 panels × 4KB per render × 60 renders/minute = ~1.4 MB per minute memory leak.

**Consequences:**
- Memory usage grows 5-10 MB per panel added
- Baseline memory usage >20 MB for simple dashboard showing text
- Long-running sessions (hours) consume hundreds of MB
- GC thrashing causes periodic lag spikes
- Application slow to exit (GC struggling with accumulated allocations)

**How to avoid:**
1. **Update to Bubbles v0.21.1+** - Implements parser pooling that dramatically reduces allocations
2. **Limit viewport usage** - Not every panel needs scrolling; use static rendered lists where possible
3. **Share viewport instances** - For similar content types (e.g., multiple language lists), consider single viewport with tab switching
4. **Monitor memory growth** - Use `pprof` during development: `go tool pprof -alloc_space ./wakadash`
5. **Cap content size** - For stats panels showing "top 10", don't load full dataset into viewport
6. **Disable high-performance mode** - Standard rendering is more memory-efficient for dashboards not using alternate screen buffer full-time
7. **Profile before and after** - Measure memory baseline with 1 panel, then 3 panels, then 6 panels

**Warning signs:**
- Memory usage grows 5-10 MB per panel added
- Baseline memory usage >20 MB for simple dashboard
- `pprof` shows `x/ansi.GetParser` as top allocator
- Memory grows linearly with update frequency
- Application slow to exit (GC struggling with allocations)
- `htop` shows memory climbing over 30+ minutes

**Phase to address:**
Phase 2 (Stats Panels) - Must verify memory usage after implementing first 2-3 panels, before building all 6+

**References:**
- [Bubbles #829: Viewport Memory Issue](https://github.com/charmbracelet/bubbles/issues/829) - Detailed analysis with pprof output
- Fixed in Bubbles v0.21.1 with parser pooling

---

### Pitfall 5: TERM Variable Incompatibility Breaks Colors Across Terminals

**What goes wrong:**
Dashboard looks perfect in your development terminal but shows broken colors, missing styles, or no colors at all in other users' terminals. This happens because TERM variable values (`xterm`, `xterm-256color`, `tmux-256color`, `screen`) declare different color capabilities through the terminfo database. Users report "all the colors are too dark" or "borders don't appear" or "no colors at all."

**Why it happens:**
TUI applications rely on the terminfo database to look up terminal capabilities based on `$TERM`. If the system lacks a terminfo entry for the user's `$TERM` value, or if `$TERM` is set incorrectly, the application can't determine supported colors. According to terminal color research: "By setting TERM to xterm-256color, you're effectively telling whatever is running that the terminal supports all the features of xterm. This does get you pretty colors usually but you will also sacrifice some modern features."

Colors also fail when SSHing if `TERM` isn't forwarded or when inside tmux without proper configuration. The wakadash codebase currently uses GitHub Linguist colors (hex codes like "#00ADD8" for Go) which require true color support - not available in all terminals.

**Consequences:**
- "Works on my machine" but users report broken rendering
- Colors work locally but fail over SSH
- Some terminals show colors, others don't
- Color issues only in tmux/screen sessions without proper config
- True color hex codes fallback to wrong colors instead of gracefully degrading

**How to avoid:**
1. **Test across common TERM values** - Verify on `xterm`, `xterm-256color`, `tmux-256color`, `screen-256color`, `alacritty`, `kitty`
2. **Use AdaptiveColor for core UI** - Falls back gracefully when true color unavailable
3. **Provide monochrome fallback** - Add `--no-color` flag that disables all color styling
4. **Detect color support explicitly** - Check `termenv.ColorProfile()` and adjust rendering accordingly
5. **Document TERM requirements** - Specify minimum recommended TERM values in README
6. **Test in SSH sessions** - Many users will run over SSH where TERM forwarding may fail
7. **Never rely on true color exclusively** - Ensure 256-color degradation looks acceptable
8. **Verify language colors degrade** - Test that GitHub Linguist hex codes look reasonable in 256-color mode

**Warning signs:**
- "Works on my machine" but users report broken rendering
- Colors work locally but fail over SSH
- Some terminals show colors, others don't
- Color issues only in tmux/screen sessions
- User bug reports mentioning TERM value or "colors don't work"
- Language bar colors look completely different in some terminals

**Phase to address:**
Phase 1 (Theme Foundation) - Must establish terminal compatibility baseline before choosing color approaches

**References:**
- [Terminal Colours Are Tricky](https://jvns.ca/blog/2024/10/01/terminal-colours/) - Comprehensive TERM variable explanation
- [Why Terminal Emacs Requires TERM=xterm-256color](https://www.w3tutorials.net/blog/terminal-emacs-colors-only-work-with-term-xterm-256color/) - TERM compatibility details
- wakadash `/workspace/wakadash/internal/tui/colors.go` - Current language color implementation using hex codes

---

### Pitfall 6: Runtime Theme Switching Triggers Full Model Rerender

**What goes wrong:**
Implementing runtime theme switching (e.g., keyboard shortcut to toggle themes) requires rebuilding all styled components, causing noticeable UI lag or flicker. Every panel must recreate its styles from the new theme, recompute layouts, and redraw completely. For dashboards with 6+ panels and live-updating data, this creates 100-500ms lag spikes. Users see the entire screen flicker or freeze briefly.

**Why it happens:**
Lipgloss styles are immutable value types, not references. Changing theme requires creating new style objects and re-rendering all components. The BubbleTea model must propagate theme changes through the entire component tree, triggering every component's `View()` method. With complex layouts, this cascades into expensive recalculations. Material-UI (web framework) has similar issues: "Switching the theme causes that virtually all component must be recomputed, it's really slow in dev mode, and noticeable in prod mode."

**Consequences:**
- Visible lag/flicker when switching themes (100-500ms freeze)
- CPU spike during theme change visible in profilers
- Multiple theme switches cause compounding slowdown
- Panel updates pause during theme switch
- Poor UX - users avoid using theme switching because it's jarring
- In live-updating dashboard, data updates conflict with theme updates causing corruption

**How to avoid:**
1. **Consider restart-based themes** - Select theme via flag/config, require restart to change (simpler, zero runtime cost)
2. **Lazy theme propagation** - Don't rebuild all panels immediately; rebuild each panel on next update
3. **Cache calculated layouts** - Store panel dimensions/positions; only recalculate styles, not layouts
4. **Debounce theme changes** - Prevent rapid theme toggling (e.g., 500ms cooldown)
5. **Memoize style creation** - Create theme styles once, store in model, reference rather than recreate
6. **Progressive redraw** - Mark panels dirty, redraw over 3-4 frames instead of synchronously
7. **Measure performance** - Benchmark `View()` execution time; target <50ms for theme switch
8. **Document restart requirement** - If choosing restart-based approach, show "Theme changed, restart to apply" message

**Warning signs:**
- Visible lag/flicker when switching themes
- CPU spike during theme change
- `View()` methods recreating styles instead of using theme references
- Multiple theme switches cause compounding slowdown
- Panel updates pause during theme switch
- Users reporting "laggy" theme switching

**Phase to address:**
Phase 3 (Theme Switching) - Only if implementing runtime switching; Phase 1 if restart-based

**References:**
- [Material-UI #25018: Theme Switching Performance](https://github.com/mui/material-ui/issues/25018) - Similar performance issues in web UI framework
- [OpenTUI #3731: System Theme Support](https://github.com/anomalyco/opencode/issues/3731) - Implementation discussion for theme systems

---

### Pitfall 7: API Rate Limiting Triggers with Multiple Panel Data Fetches

**What goes wrong:**
Loading data for 6+ stats panels (Languages, Projects, Categories, Editors, OS, Machines) simultaneously hits WakaTime's rate limit of "fewer than 10 requests per second averaged over 5 minutes." App shows incomplete dashboards or rate limit errors because each panel independently fetches its dataset. Dashboard loads with some panels showing data and others showing "Loading..." indefinitely.

**Why it happens:**
Naive implementation treats each panel as independent, issuing separate API requests. Six panels requesting data on startup = 6 requests in <1 second, already at 60% of rate limit. Add periodic refresh every 30 seconds, and you'll average >10 req/sec within 5 minutes: (6 requests × 10 refreshes) / 300 seconds = 12 req/5min average. WakaTime returns HTTP 429 when rate limited.

**Consequences:**
- HTTP 429 errors after 3-5 dashboard loads within 5 minutes
- Dashboard shows "Loading..." indefinitely for some panels
- Works on first run, fails after several refreshes
- Some panels load, others timeout
- Error messages mentioning rate limits
- Auto-refresh makes problem worse (continued requests while rate limited)

**How to avoid:**
1. **Single API request, fan-out data** - WakaTime's `/users/current/stats` returns all stats in one response; parse and distribute to panels
2. **Batch fetches with delays** - If multiple endpoints needed, space requests 200ms apart to stay under 10/sec
3. **Cache aggressively** - Don't refetch data that updates infrequently (e.g., all-time stats)
4. **Implement exponential backoff** - On 429, wait 60s before retry, doubling on each subsequent 429
5. **Show rate limit warnings** - Display "Rate limited, retrying in 60s" with countdown instead of silent failure
6. **Respect Retry-After header** - WakaTime may include this; honor it
7. **Local caching** - Store last successful response, continue showing while waiting for rate limit to clear
8. **Test with rapid refreshes** - Run dashboard with 30s refresh interval for 10 minutes, verify no 429s

**Warning signs:**
- HTTP 429 errors in logs
- Dashboard shows "Loading..." indefinitely for some panels
- Works on first run, fails after several refreshes
- Error messages mentioning rate limits
- Some panels load, others time out
- Problem gets worse with faster refresh intervals

**Phase to address:**
Phase 2 (Stats Panels) - Critical during API integration when adding multiple data fetches

**References:**
- [WakaTime API Docs: Rate Limiting](https://wakatime.com/developers) - "fewer than 10 requests per second on average over any 5 minute period"
- [WakaTime FAQ](https://wakatime.com/faq) - Rate limit details and backoff strategies

---

### Pitfall 8: Border Calculations Break Multi-Panel Layouts

**What goes wrong:**
When calculating available space for panel content, forgetting to account for borders causes content to overflow panel boundaries, misaligned layouts, or panels that don't fit within terminal window. Classic mistake: using `m.height` directly instead of `m.height - 2` (top + bottom border). With 6 panels in a grid, these 2-character offsets compound, causing 12+ characters of miscalculation in layout calculations.

**Why it happens:**
Developers think in terms of "available space" but borders consume that space. A panel with `lipgloss.RoundedBorder()` needs 2 characters vertical (top/bottom) and 2 horizontal (left/right). When laying out 6 panels in a 2×3 grid, forgetting borders means 6 panels × 2 chars each = 12 characters of overflow. At terminal height 24, this means panels try to occupy 36 lines - doesn't fit.

**Consequences:**
- Content overflows panel borders (text visible outside borders)
- Panels don't fit in terminal window (bottom panels cut off)
- Bottom panels cut off at small terminal sizes
- Layout breaks when adding 6th panel but works with 5
- Horizontal scrolling appears unexpectedly
- Panels overlap each other in grid layouts

**How to avoid:**
1. **Always subtract borders FIRST** - Calculate `contentHeight = m.height - 2` before any panel rendering
2. **Create layout calculator function** - Centralize logic:
   ```go
   func calculatePanelLayout(totalHeight, borderChars, numPanels int) int {
       return (totalHeight - (borderChars * numPanels)) / numPanels
   }
   ```
3. **Document border constants** - `const BorderHeightCost = 2` with comments explaining usage
4. **Test at minimum terminal size** - Verify 80×24 terminal doesn't cause overflow
5. **Explicit truncation over wrapping** - In bordered panels, always truncate text rather than auto-wrap
6. **Account for titles** - If panels have title bars, subtract those too: `contentHeight - 2 (borders) - 1 (title)`
7. **Use golden rule** - "Always Account for Borders - Subtract 2 from height calculations BEFORE rendering panels"

**Warning signs:**
- Content overflows panel borders
- Panels don't fit in terminal window
- Bottom panels cut off at small terminal sizes
- Layout breaks when adding 6th panel but works with 5
- Horizontal scrolling appears unexpectedly
- Math doesn't add up: terminal height 24, but panels try to use 30 lines

**Phase to address:**
Phase 2 (Stats Panels) - Must establish correct layout calculations before implementing multi-panel grid

**References:**
- [Tips for Building BubbleTea Programs: Layout Rules](https://leg100.github.io/en/posts/building-bubbletea-programs/) - "Always Account for Borders - Subtract 2 from height calculations BEFORE rendering panels"

---

## Moderate Pitfalls

### Pitfall 9: Theme Config and API Key in Same File

**What goes wrong:**
Storing theme selection in the same config file as the WakaTime API key (`~/.wakatime.cfg`) causes users to accidentally share API keys when sharing theme configurations in dotfiles repos, screenshots, or screen shares. Theme configs are intended to be shareable, but API keys must remain secret.

**Why it happens:**
Single config file pattern is simpler to implement. Developers add new settings to existing config without considering security boundaries. Users copy entire config files when sharing customizations.

**How to avoid:**
1. **Separate config files** - API key in `~/.wakatime.cfg`, theme in `~/.config/wakadash/theme.toml`
2. **Document separation** - README explains why configs are separate
3. **Default to safe sharing** - Theme config should be safe to commit to dotfiles repo

**Phase to address:**
Phase 1 (Theme Foundation) - Establish config file structure before implementation

---

### Pitfall 10: Creating New Styles in View() on Every Render

**What goes wrong:**
Creating Lipgloss style objects inside `View()` method on every render causes CPU spikes and 50-200ms render times. With 6+ panels each recreating 5-10 styles per render, this compounds to thousands of allocations per second. Users experience UI lag on updates.

**Why it happens:**
Straightforward implementation creates styles where they're used. Developers don't realize Lipgloss style creation has measurable performance cost when done thousands of times per second.

**How to avoid:**
1. **Create styles once in Init()** - Store in model struct, reference in View()
2. **Recreate only on theme change** - In Update() when theme changes, rebuild styles once
3. **Memoize by theme** - Cache styles per theme, lookup instead of recreate
4. **Profile before optimizing** - Measure View() execution time; target <50ms per frame

**Phase to address:**
Phase 1 (Theme Foundation) - Establish style creation patterns before building panels

---

### Pitfall 11: No Minimum Terminal Size Check

**What goes wrong:**
Dashboard attempts to render in 80×24 terminal, panels overlap, text truncated mid-word, borders broken. Users see corrupted display and don't understand why.

**How to avoid:**
1. **Set minimum size** - Require 100×30 or similar based on panel layout
2. **Show friendly message** - "Terminal too small (80×24), need at least 100×30" with current size
3. **Test at 80×24** - Verify graceful degradation or clear error message

**Phase to address:**
Phase 2 (Stats Panels) - Test with final layout to determine minimum size

---

### Pitfall 12: All Panels Update Simultaneously

**What goes wrong:**
All 6 panels redraw simultaneously every refresh interval, causing CPU spike, brief freeze, and distracting full-screen flash. Users see periodic lag spikes every 30-60 seconds.

**How to avoid:**
1. **Stagger panel updates** - Update panels 5 seconds apart, or
2. **Update only focused panel** - Full refresh only on user request
3. **Differential rendering** - Only redraw panels with changed data

**Phase to address:**
Phase 2 (Stats Panels) - Implement during refresh logic

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Hardcode colors directly in components | Fast prototyping, no abstraction needed | Every theme change requires finding/updating dozens of color codes | Only for throwaway proofs-of-concept, never for production |
| Skip theme system, use flag for preset selection only | Simpler implementation, no runtime switching complexity | User must restart to change theme, limits flexibility | Acceptable for v1.0 if runtime switching not critical UX requirement |
| Fetch all panel data simultaneously without batching | Simpler code, all panels populate at once | Rate limiting on every dashboard load, poor UX at scale | Never acceptable with known rate limits - design with batching from start |
| Use High Performance Rendering for all viewports | Potentially smoother rendering in some terminals | Higher memory usage, deprecated feature, may break in future versions | Only when profiling proves standard rendering insufficient for use case |
| Manual style creation in each View() call | Straightforward, no state management | Performance penalty on every render, noticeable lag with 6+ panels | Never - always create styles once in Init() or on theme change |
| Set .Width() on all dynamic text elements | Consistent visual alignment | Rendering corruption bugs on rerenders, hard to debug | Only for static text; use containers with MaxWidth() for dynamic content |
| Store theme config with API key | Single file, simpler to manage | API key leaks when sharing theme configs | Never - always separate security-sensitive from shareable configs |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| WakaTime API | Separate request per panel (6+ simultaneous requests) | Single `/stats` request, distribute data to panels |
| Lipgloss AdaptiveColor | Using in View() without pre-initialization | Call `lipgloss.HasDarkBackground()` in main() before BubbleTea starts |
| BubbleTea viewports | Creating viewport per panel without pooling | Update to Bubbles v0.21.1+ with parser pooling, limit viewport usage |
| Terminal detection | Assuming TERM=xterm-256color always available | Test across TERM values, provide 256-color degradation, add --no-color flag |
| Theme switching | Rebuilding all styles synchronously on theme change | Cache theme styles in model, use lazy/progressive redraw, or restart-based |
| Language colors | Using GitHub Linguist hex codes without fallback | Verify 256-color degradation looks acceptable, test in degraded terminals |
| Border calculations | Using total height directly for panel content | Always subtract border chars first: `contentHeight = total - 2` |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Creating new styles in View() on every render | CPU spikes, 50-200ms render times, UI lag on updates | Create styles once in Init() or Update() on theme change, store in model | Becomes noticeable with 3+ panels, unacceptable at 6+ panels |
| Viewport without content limits | Memory grows from 5 MB to 40+ MB over time | Cap content size (e.g., "top 10" not full dataset), update x/ansi dependency to v0.21.1+ | Breaks when multiple viewports × frequent updates × long-running session |
| Synchronous theme rebuild | 100-500ms lag spike, visible flicker, frozen UI | Lazy rebuild (mark dirty, rebuild on next update), or restart-based themes | Noticeable with 4+ panels, unacceptable with 6+ panels updating live |
| Auto-wrap in bordered panels | Layout corruption, text bleeding outside borders, misaligned panels | Explicit truncation with lipgloss.NewStyle().MaxWidth(), never rely on auto-wrap | Breaks immediately when content exceeds panel width |
| Simultaneous API requests | HTTP 429 after 3-5 dashboard loads within 5 minutes | Batch requests with 200ms delays, single aggregate API call preferred | Hits rate limit after <5 minutes of normal usage with 30s refresh |
| Dynamic .Width() styles | Rendering corruption, faint text, bleeding borders after 10-20s of updates | Use .MaxWidth() on containers, explicit truncation for dynamic text | Manifests after 10-20 seconds of live updates, compounds over time |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Storing API key in theme config | API key leaked in dotfiles repos, screenshots, screen shares | Separate config files: api-key in ~/.wakatime.cfg, theme in ~/.config/wakadash/theme.toml |
| Exposing full API errors in UI | Leaks API endpoint structure, rate limit details useful for abuse | Sanitize error messages: show "API request failed" not full response body |
| No timeout on background color detection | User holds Ctrl+C, terminal query hangs, process unkillable | 5-second timeout on all terminal queries with graceful fallback |
| Logging sensitive data during API calls | API keys in log files, stats data in crash reports | Strip Authorization headers before logging, redact numeric stats in debug mode |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Silent rate limit failures | Dashboard shows "Loading..." forever, no explanation | Display "Rate limited, retrying in 60s" with countdown |
| Theme change requires restart without explanation | User toggles theme shortcut, nothing happens, confusion | If restart-required: show "Theme changed, restart to apply" message |
| No visual feedback during API fetch | User sees blank panels, thinks app frozen | Show skeleton loaders or "Fetching data..." in each panel |
| Broken rendering on unknown terminals | Colors missing, borders wrong, unusable on user's terminal | Detect terminal capabilities, fall back to monochrome gracefully, test on 5+ TERM values |
| Panels overflow small terminals | Dashboard unusable on 80×24 terminals, horizontal scroll needed | Set minimum terminal size (100×30), show "Terminal too small" message with dimensions below threshold |
| All panels update simultaneously | CPU spike every 30s, brief freeze, distracting refresh | Stagger panel updates 5s apart, or refresh only focused panel |
| Language colors don't degrade well | In 256-color mode, colors indistinguishable or jarring | Test GitHub Linguist hex codes in 256-color mode, adjust if needed |

## "Looks Done But Isn't" Checklist

- [ ] **Theme system:** Verified NO hardcoded lipgloss.Color() values remain in codebase (grep audit passed)
- [ ] **Terminal compatibility:** Tested on xterm, xterm-256color, tmux, screen, alacritty, kitty (not just development terminal)
- [ ] **Rate limiting:** Confirmed single API request or batched with delays (not 6+ simultaneous requests)
- [ ] **Memory profiling:** Verified <10 MB baseline memory, stable over 30+ minutes of updates
- [ ] **Border calculations:** All panel heights subtract border characters (contentHeight = total - 2)
- [ ] **AdaptiveColor initialization:** Called HasDarkBackground() in main() before program.Run()
- [ ] **Dynamic width styles:** No .Width() on dynamic text, only on containers or static text
- [ ] **Error handling:** API failures show user-friendly messages, not raw error dumps
- [ ] **Minimum terminal size:** Tested at 80×24, 100×30 - shows graceful message if too small
- [ ] **Theme persistence:** Selected theme saved to config, loaded correctly on restart
- [ ] **Color fallback:** Verified --no-color flag disables all styling, still usable
- [ ] **Render performance:** Measured View() execution <50ms per frame (100ms at absolute max)
- [ ] **Language color degradation:** Tested GitHub Linguist colors in 256-color mode, look acceptable
- [ ] **Bubbles version:** Updated to v0.21.1+ for viewport memory fix
- [ ] **BubbleTea version:** Updated to v0.27.1+ for AdaptiveColor hang fix
- [ ] **Style creation:** Styles created once in Init/Update, not in View() on every render

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Hardcoded colors discovered after implementing 3 themes | MEDIUM | 1. Grep all lipgloss.Color() calls, 2. Create theme.Colors struct, 3. Mass find/replace with theme references, 4. Visual regression test each theme |
| Memory leak from viewports | LOW | 1. Update x/ansi dependency to latest, 2. Verify Bubbles v0.21.1+, 3. Run pprof to confirm fix |
| AdaptiveColor startup hang | LOW | 1. Add `_ = lipgloss.HasDarkBackground()` before program.Run(), 2. Update BubbleTea to v0.27.1+ |
| Hit WakaTime rate limit | MEDIUM | 1. Implement exponential backoff (60s, 120s, 240s), 2. Add caching layer, 3. Refactor to single API call if possible |
| Layout breaks with 6 panels | MEDIUM | 1. Create layout calculator utility, 2. Audit all height calculations, 3. Add border constants, 4. Test at 80×24 |
| Width() rendering corruption | HIGH | 1. Identify all .Width() usage on dynamic text, 2. Replace with .MaxWidth() on containers, 3. Add explicit truncation, 4. Regression test rapid rerenders |
| Theme switch performance lag | MEDIUM-HIGH | 1. Profile View() execution times, 2. Move style creation to Init()/Update(), 3. Implement lazy rebuild or switch to restart-based themes |
| Terminal incompatibility | LOW-MEDIUM | 1. Add termenv.ColorProfile() detection, 2. Implement 256-color fallback, 3. Add --no-color flag, 4. Test on 5+ terminal types |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Hardcoded colors | Phase 1: Theme Foundation | Grep returns zero lipgloss.Color() with hex/ANSI literals |
| AdaptiveColor hangs | Phase 1: Theme Foundation | App starts <1s on all test terminals, no hanging |
| Dynamic width corruption | Phase 2: Stats Panels | Run dashboard for 5 minutes, no rendering degradation |
| Viewport memory leaks | Phase 2: Stats Panels | Memory stable after 30 minutes, <10 MB baseline |
| TERM incompatibility | Phase 1: Theme Foundation | Test matrix passes: xterm, xterm-256color, tmux, screen, kitty, alacritty |
| Runtime theme switching lag | Phase 3: Theme Switching | Theme switch completes <100ms, no visible flicker (if runtime switching implemented) |
| API rate limiting | Phase 2: Stats Panels | Run dashboard for 10 minutes with 30s refresh, zero HTTP 429 errors |
| Border calculation errors | Phase 2: Stats Panels | All panels fit at 100×30 terminal size, no overflow |
| Style creation in View() | Phase 1: Theme Foundation | Profile shows View() execution <50ms |
| Language color degradation | Phase 1: Theme Foundation | Visual test in 256-color terminal, colors distinguishable |
| Config file separation | Phase 1: Theme Foundation | Theme config separate from API key config |
| Minimum terminal size | Phase 2: Stats Panels | Clear error message at 80×24, works at 100×30 |

## Sources

### Official Documentation
- [WakaTime API Docs](https://wakatime.com/developers) - Rate limiting: <10 req/sec averaged over 5 minutes
- [Lipgloss GitHub](https://github.com/charmbracelet/lipgloss) - Style definitions and adaptive colors
- [BubbleTea GitHub](https://github.com/charmbracelet/bubbletea) - Elm Architecture TUI framework
- [Bubbles GitHub](https://github.com/charmbracelet/bubbles) - TUI components including viewport

### Critical GitHub Issues (PRIMARY SOURCES)
- [BubbleTea #1036](https://github.com/charmbracelet/bubbletea/issues/1036) - AdaptiveColor hanging on startup - FIXED v0.27.1
- [BubbleTea #1225](https://github.com/charmbracelet/bubbletea/issues/1225) - Width() rendering corruption on rerenders
- [Bubbles #829](https://github.com/charmbracelet/bubbles/issues/829) - Viewport memory usage - FIXED v0.21.1

### Theme System References
- [Lipgloss Theme Package](https://pkg.go.dev/github.com/purpleclay/lipgloss-theme) - Adaptive color palette example
- [Glitter UI Library](https://github.com/brittonhayes/glitter) - Theme components for Lipgloss/BubbleTea
- [Design Tokens & Theming](https://materialui.co/blog/design-tokens-and-theming-scalable-ui-2025) - Migration from hardcoded values
- [Material-UI #25018](https://github.com/mui/material-ui/issues/25018) - Theme switching performance issues
- [OpenTUI #3731](https://github.com/anomalyco/opencode/issues/3731) - System theme support implementation

### Terminal Compatibility
- [Terminal Colours Are Tricky](https://jvns.ca/blog/2024/10/01/terminal-colours/) - Comprehensive TERM variable explanation
- [So You Want to Make a TUI](https://p.janouch.name/article-tui.html) - True color vs 256-color, terminal compatibility
- [Why Terminal Emacs Requires TERM=xterm-256color](https://www.w3tutorials.net/blog/terminal-emacs-colors-only-work-with-term-xterm-256color/) - TERM compatibility details
- [Ratatui Color Discussion](https://github.com/ratatui/ratatui/discussions/877) - Choosing colors for terminal emulators

### Best Practices & Performance
- [Tips for Building BubbleTea Programs](https://leg100.github.io/en/posts/building-bubbletea-programs/) - Layout calculations, receiver types, golden rules
- [Phoenix TUI Framework](https://github.com/phoenix-tui/phoenix) - Differential rendering, emoji width issues in Lipgloss
- [BubbleLayout Package](https://pkg.go.dev/github.com/winder/bubblelayout) - Declarative layout management
- [Dashboard Design Best Practices](https://www.sisense.com/blog/4-design-principles-creating-better-dashboards/) - Layout complexity management
- [Dashboard UX Patterns](https://www.pencilandpaper.io/articles/ux-pattern-analysis-data-dashboards) - Multi-panel complexity

### Project Code Analysis
- `/workspace/wakadash/internal/tui/styles.go` - Current hardcoded colors (Color "62", "205", "241", "196", "214")
- `/workspace/wakadash/internal/tui/colors.go` - Language colors using GitHub Linguist hex codes
- `/workspace/wakafetch/ui/colors.go` - Legacy ANSI color implementation for reference

---

*Pitfalls research for: wakadash v2.1 Visual Overhaul + Themes*
*Researched: 2026-02-19*
*Primary focus: Integration pitfalls when adding theme system and multi-panel layouts to existing TUI application*
*Confidence: HIGH for Lipgloss/BubbleTea issues (official GitHub issues with maintainer responses), MEDIUM for terminal compatibility (community sources), HIGH for WakaTime API (official documentation)*
