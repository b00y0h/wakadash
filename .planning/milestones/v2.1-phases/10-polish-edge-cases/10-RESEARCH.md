# Phase 10: Polish + Edge Cases - Research

**Researched:** 2026-02-20
**Domain:** Error handling, graceful degradation, and edge case management in Go TUI applications
**Confidence:** HIGH

## Summary

Phase 10 focuses on hardening the wakadash dashboard against edge cases and unexpected inputs. The three success criteria target distinct failure modes: terminal size constraints, invalid configuration values, and missing API data. BubbleTea applications must handle these gracefully without crashing, as panics can leave the terminal in a broken state requiring `stty sane` to recover.

Go's explicit error handling philosophy aligns well with TUI edge cases. Rather than exceptions, Go code returns errors as values, enabling sentinel error patterns and graceful fallbacks. For TUIs specifically, the View() function is the primary defense: it must check preconditions (terminal size, data availability) before rendering and provide informative fallback messages.

Current wakadash code already implements several defensive patterns: minimum terminal size checking (40x10), nil checks for stats data before rendering, and theme fallback to Dracula on invalid names. However, these patterns are incomplete. The terminal size check displays a generic message that doesn't guide users toward resolution. Missing data categories (when API returns empty arrays) render "No data" but don't explain why. Theme fallback is silent, potentially confusing users who typoed a theme name.

**Primary recommendation:** Strengthen error messages with actionable guidance. For terminal size errors, display current dimensions and minimum requirements. For invalid theme names, log a warning with available options. For missing data, distinguish between "no activity yet" vs "API returned empty response" vs "network error." Follow BubbleTea's error message pattern of including recovery hints in the message itself.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/charmbracelet/bubbletea | v1.3.10+ | TUI framework with panic recovery | Already in use; handles graceful shutdown and panic recovery |
| github.com/charmbracelet/lipgloss | v1.1.0+ | Styled error message rendering | Already in use; provides terminal-safe text rendering |
| errors | stdlib | Sentinel errors and error wrapping | Go standard library; idiomatic error handling |
| fmt | stdlib | Error formatting with context | Go standard library; used throughout codebase |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| log | stdlib | Development-time warnings for config issues | For non-critical issues like invalid theme names |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom error types | Sentinel errors | Sentinel errors simpler for this scale; custom types overkill |
| Panic on invalid input | Graceful fallback | Panics leave terminal broken; fallbacks keep app running |
| Silent failures | Logged warnings | Silent failures confuse users; warnings provide debugging hints |

**Installation:**
No new dependencies required. All error handling uses stdlib and existing BubbleTea/lipgloss.

## Architecture Patterns

### Recommended Error Handling Structure
```
wakadash/internal/
├── tui/
│   ├── model.go         # View() with precondition checks
│   ├── errors.go        # NEW: Helper functions for error messages
│   └── ...
├── theme/
│   └── config.go        # Add fallback logging for invalid themes
└── api/
    └── client.go        # Already has comprehensive HTTP error handling
```

### Pattern 1: Defensive View Rendering with Informative Messages

**What:** Check all preconditions in View() before rendering and return actionable error messages.

**When to use:** Every View() function that depends on terminal size, data availability, or configuration.

**Example:**
```go
// Source: BubbleTea best practices and wakadash existing pattern

func (m Model) View() string {
    if m.quitting {
        return ""
    }

    // Check minimum terminal size with helpful message
    if m.width < 40 || m.height < 10 {
        return renderTerminalTooSmallError(m.width, m.height, 40, 10)
    }

    // Check for errors with recovery hint
    if m.err != nil {
        return ErrorStyle(m.theme).Render(
            fmt.Sprintf("\n  Error: %v\n\n  Press r to retry, q to quit.", m.err),
        )
    }

    // Check for missing data (distinguish from loading state)
    if m.stats == nil && !m.loading {
        return WarningStyle(m.theme).Render(
            "\n  No data available.\n\n  Press r to refresh, q to quit.",
        )
    }

    return m.renderDashboard()
}

// Helper function for terminal size errors
func renderTerminalTooSmallError(width, height, minWidth, minHeight int) string {
    return fmt.Sprintf(
        "Terminal too small\n\n"+
            "Current:  %dx%d\n"+
            "Required: %dx%d (width x height)\n\n"+
            "Please resize your terminal and the dashboard will adjust automatically.",
        width, height, minWidth, minHeight,
    )
}
```

### Pattern 2: Theme Fallback with Warning

**What:** When loading an invalid theme name from config, fall back to default and log a warning.

**When to use:** Configuration loading where user-edited values might be invalid.

**Example:**
```go
// Source: Go error handling best practices

package theme

import (
    "log"
    "strings"
)

// GetTheme returns a Theme by name.
// Returns the Dracula theme if the name is not recognized.
func GetTheme(name string) Theme {
    normalizedName := strings.ToLower(strings.TrimSpace(name))

    switch normalizedName {
    case "dracula":
        return Dracula
    case "nord":
        return Nord
    case "gruvbox":
        return Gruvbox
    case "monokai":
        return Monokai
    case "solarized":
        return Solarized
    case "tokyonight":
        return TokyoNight
    default:
        if normalizedName != "" {
            // Log warning for invalid theme (but don't crash)
            log.Printf(
                "Warning: unknown theme %q, falling back to 'dracula'. "+
                    "Available themes: %v",
                name,
                AllThemes(),
            )
        }
        return Dracula
    }
}
```

### Pattern 3: Missing Data Category Handling

**What:** Distinguish between empty data (no activity) and missing API fields (data category not returned).

**When to use:** Rendering panels that depend on API data arrays (Categories, Editors, OS, Machines).

**Example:**
```go
// Source: Existing wakadash stats_panels.go pattern

func (m Model) renderCategoriesPanel() string {
    var sb strings.Builder
    sb.WriteString(TitleStyle(m.theme).Render("Categories") + "\n")

    // Check if stats loaded
    if m.stats == nil {
        return sb.String()  // Empty panel during loading
    }

    // Check if data category exists
    if len(m.stats.Data.Categories) == 0 {
        sb.WriteString(DimStyle(m.theme).Render("  No categories tracked yet"))
    } else {
        sb.WriteString(m.categoriesChart.View())
    }

    return sb.String()
}

// Similar pattern for Editors, OS, Machines panels
```

### Pattern 4: Graceful Degradation for Partial Data

**What:** When some data is available but other fields are missing, render what you have and indicate missing pieces.

**When to use:** When API returns partial responses (e.g., stats but no durations).

**Example:**
```go
// Source: BubbleTea graceful degradation pattern

func (m Model) renderDashboard() string {
    // ... existing code

    // Render available panels, skip unavailable ones
    var panels []string

    if m.stats != nil {
        if m.showLanguages && len(m.stats.Data.Languages) > 0 {
            panels = append(panels, m.renderLanguagesPanel())
        }
        if m.showProjects && len(m.stats.Data.Projects) > 0 {
            panels = append(panels, m.renderProjectsPanel())
        }
    }

    // Heatmap depends on summaryData
    if m.showHeatmap {
        if m.summaryData != nil && len(m.summaryData.Data) > 0 {
            panels = append(panels, m.renderHeatmapPanel())
        } else {
            // Show placeholder if heatmap enabled but data missing
            panels = append(panels, m.renderHeatmapPlaceholder())
        }
    }

    return lipgloss.JoinVertical(lipgloss.Left, panels...)
}

func (m Model) renderHeatmapPlaceholder() string {
    return DimStyle(m.theme).Render("Heatmap: No activity data available")
}
```

### Pattern 5: Input Validation with Sentinel Errors

**What:** Define sentinel errors for common validation failures and check with errors.Is().

**When to use:** When validating user input or configuration values that might be invalid.

**Example:**
```go
// Source: Go error handling patterns 2026

package theme

import "errors"

var (
    ErrThemeNotFound = errors.New("theme not found")
    ErrInvalidConfig = errors.New("invalid theme configuration")
)

// ValidateThemeName checks if a theme name is valid
func ValidateThemeName(name string) error {
    normalizedName := strings.ToLower(strings.TrimSpace(name))

    if normalizedName == "" {
        return ErrInvalidConfig
    }

    for _, validName := range AllThemes() {
        if normalizedName == validName {
            return nil
        }
    }

    return ErrThemeNotFound
}

// Usage in config loading:
themeName, _ := LoadThemeFromConfig()
if err := ValidateThemeName(themeName); err != nil {
    if errors.Is(err, ErrThemeNotFound) {
        log.Printf("Unknown theme %q, using default", themeName)
        themeName = DefaultTheme
    }
}
```

### Anti-Patterns to Avoid

- **Panic on invalid input:** Panics leave terminal in broken state; always return errors or fall back to defaults
- **Silent failures without logging:** Users need hints about what went wrong; log warnings for config issues
- **Generic error messages:** "Error occurred" doesn't help; include specific issue and recovery steps
- **Ignoring nil checks:** Always check `if m.stats == nil` before accessing nested fields
- **Hardcoded error strings:** Use fmt.Sprintf or errors.New for consistency and testability

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Terminal size detection | Custom ANSI escape parsing | BubbleTea's WindowSizeMsg | BubbleTea handles resize events and provides width/height in message |
| Panic recovery | Custom recover() wrapper | BubbleTea's built-in panic recovery | BubbleTea already recovers from panics in Update loop and returns ErrProgramPanic |
| Error type hierarchies | Complex error type trees | Sentinel errors with errors.Is() | Simple sentinel errors sufficient for this scale; no need for type assertions |
| Logging framework | Full logger with levels/rotation | stdlib log package | Simple warnings for config issues don't justify logger dependency |

**Key insight:** BubbleTea handles terminal state management and panic recovery automatically. Your job is to validate inputs, check preconditions, and provide clear error messages. Don't reinvent terminal handling or panic recovery.

## Common Pitfalls

### Pitfall 1: Terminal Left in Broken State After Panic

**What goes wrong:** When application panics without BubbleTea's recovery, the terminal is left in raw mode with no cursor, requiring `stty sane` to fix.

**Why it happens:** BubbleTea's panic recovery only works for panics in Update() loop. Panics in Init() or main() before program starts are unrecovered.

**How to avoid:**
1. Don't panic in main() or Init() - return errors instead
2. Validate all inputs before starting BubbleTea program
3. Use `defer` in main() to restore terminal state if needed

**Warning signs:** Terminal shows no cursor after crash, typed commands not echoed, shell prompt misaligned.

**Source:** [BubbleTea Issue #1459](https://github.com/charmbracelet/bubbletea/issues/1459)

### Pitfall 2: Division by Zero in Percentage Calculations

**What goes wrong:** When calculating percentages for stats panels, dividing by zero total causes panic.

**Why it happens:** API might return empty categories array, or all values are zero.

**How to avoid:**
```go
// Calculate total for percentages
var total float64
for _, cat := range data.Categories {
    total += cat.TotalSeconds
}

// Protect against division by zero
if total == 0 {
    total = 1  // Avoid divide by zero; percentages will all be 0%
}

percent := (cat.TotalSeconds / total) * 100
```

**Warning signs:** Panic with "runtime error: floating point divide by zero" in stats panel rendering.

### Pitfall 3: Accessing Nil Slices Without Length Check

**What goes wrong:** Code assumes API always returns non-empty arrays, crashes when accessing slice elements.

**Why it happens:** API might return empty arrays for new users or certain time ranges.

**How to avoid:**
```go
// WRONG:
languages := m.stats.Data.Languages[:5]  // Panics if Languages has < 5 items

// RIGHT:
limit := 5
if len(data.Languages) < limit {
    limit = len(data.Languages)
}
languages := data.Languages[:limit]
```

**Warning signs:** Panic with "runtime error: slice bounds out of range" when rendering charts.

**Source:** Already implemented correctly in existing wakadash code (stats_panels.go)

### Pitfall 4: Case-Sensitive Theme Name Comparison

**What goes wrong:** User types "Dracula" or "DRACULA" in config but GetTheme() expects lowercase "dracula", falls back to default silently.

**Why it happens:** String comparison is case-sensitive; user-edited config might use different casing.

**How to avoid:** Normalize theme names with `strings.ToLower()` before comparison (Pattern 2 above).

**Warning signs:** User's theme selection ignored despite correct spelling; always shows Dracula theme.

### Pitfall 5: Missing WindowSizeMsg Before First Render

**What goes wrong:** View() is called before WindowSizeMsg arrives, causing width/height to be zero or default values.

**Why it happens:** BubbleTea calls View() immediately, but WindowSizeMsg arrives asynchronously.

**How to avoid:** Initialize Model with safe defaults (80x24) and update on WindowSizeMsg. Never assume initial size is accurate.

**Warning signs:** Layout broken on first render, corrects itself after resize.

**Source:** Already handled correctly in wakadash Model initialization (80x24 defaults)

## Code Examples

Verified patterns for Phase 10 implementation:

### Example 1: Enhanced Terminal Size Error Message

```go
// Source: Improved version of existing wakadash model.go pattern

func (m Model) View() string {
    if m.quitting {
        return ""
    }

    // Enhanced terminal size check with informative message
    const minWidth = 40
    const minHeight = 10

    if m.width < minWidth || m.height < minHeight {
        // Use theme-aware styling even for error messages
        errorStyle := lipgloss.NewStyle().
            Foreground(m.theme.Error).
            Bold(true)

        dimStyle := lipgloss.NewStyle().
            Foreground(m.theme.Dim)

        var sb strings.Builder
        sb.WriteString(errorStyle.Render("Terminal Too Small") + "\n\n")
        sb.WriteString(dimStyle.Render(fmt.Sprintf(
            "Current size:  %d cols × %d rows\n"+
            "Required:      %d cols × %d rows\n\n"+
            "Please resize your terminal window.\n"+
            "The dashboard will adjust automatically.",
            m.width, m.height, minWidth, minHeight,
        )))

        return "\n" + sb.String() + "\n"
    }

    // ... rest of View logic
}
```

### Example 2: Theme Fallback with Informative Warning

```go
// Source: Enhanced version of existing theme/theme.go GetTheme()

package theme

import (
    "log"
    "strings"
)

// GetTheme returns a Theme by name.
// Invalid names fall back to Dracula with a warning.
func GetTheme(name string) Theme {
    normalizedName := strings.ToLower(strings.TrimSpace(name))

    switch normalizedName {
    case "dracula":
        return Dracula
    case "nord":
        return Nord
    case "gruvbox":
        return Gruvbox
    case "monokai":
        return Monokai
    case "solarized":
        return Solarized
    case "tokyonight":
        return TokyoNight
    default:
        // Only log warning for non-empty invalid names
        // Empty string is expected on first run
        if normalizedName != "" {
            log.Printf(
                "Warning: unknown theme %q in config, using 'dracula' instead.\n"+
                "Available themes: %v\n"+
                "To fix: edit ~/.wakatime.cfg and set theme to one of the above.",
                name,
                strings.Join(AllThemes(), ", "),
            )
        }
        return Dracula
    }
}
```

### Example 3: Robust Data Category Rendering

```go
// Source: Enhanced version of existing tui/stats_panels.go pattern

func (m Model) renderCategoriesPanel() string {
    var sb strings.Builder
    sb.WriteString(TitleStyle(m.theme).Render("Categories") + "\n")

    // Handle nil stats (shouldn't happen, but defensive)
    if m.stats == nil {
        return sb.String()
    }

    // Handle empty categories array
    if len(m.stats.Data.Categories) == 0 {
        noDataMsg := DimStyle(m.theme).Render("  No categories tracked in this time range")
        sb.WriteString(noDataMsg)
        return sb.String()
    }

    // Render chart normally
    sb.WriteString(m.categoriesChart.View())
    return sb.String()
}

// Apply same pattern to Editors, OS, Machines panels
func (m Model) renderEditorsPanel() string {
    var sb strings.Builder
    sb.WriteString(TitleStyle(m.theme).Render("Editors") + "\n")

    if m.stats == nil {
        return sb.String()
    }

    if len(m.stats.Data.Editors) == 0 {
        sb.WriteString(DimStyle(m.theme).Render("  No editors tracked in this time range"))
        return sb.String()
    }

    sb.WriteString(m.editorsChart.View())
    return sb.String()
}

func (m Model) renderOSPanel() string {
    var sb strings.Builder
    sb.WriteString(TitleStyle(m.theme).Render("Operating Systems") + "\n")

    if m.stats == nil {
        return sb.String()
    }

    if len(m.stats.Data.OperatingSystems) == 0 {
        sb.WriteString(DimStyle(m.theme).Render("  No operating systems tracked in this time range"))
        return sb.String()
    }

    sb.WriteString(m.osChart.View())
    return sb.String()
}

func (m Model) renderMachinesPanel() string {
    var sb strings.Builder
    sb.WriteString(TitleStyle(m.theme).Render("Machines") + "\n")

    if m.stats == nil {
        return sb.String()
    }

    if len(m.stats.Data.Machines) == 0 {
        sb.WriteString(DimStyle(m.theme).Render("  No machines tracked in this time range"))
        return sb.String()
    }

    sb.WriteString(m.machinesChart.View())
    return sb.String()
}
```

### Example 4: Graceful Handling of Missing Heatmap Data

```go
// Source: Enhanced version of model.go heatmap rendering

func (m Model) renderHeatmapPanel() string {
    heatmapTitle := TitleStyle(m.theme).Render("Activity (Last 7 Days)")

    // Check if summary data exists
    if m.summaryData == nil || len(m.summaryData.Data) == 0 {
        noDataMsg := DimStyle(m.theme).Render("\nNo activity data available for heatmap")
        return lipgloss.JoinVertical(lipgloss.Left, heatmapTitle, noDataMsg)
    }

    heatmapContent := m.renderHeatmap()
    return lipgloss.JoinVertical(lipgloss.Left, heatmapTitle, heatmapContent)
}

func (m Model) renderHeatmap() string {
    // Already checked summaryData in caller, but defensive check anyway
    if m.summaryData == nil || len(m.summaryData.Data) == 0 {
        return DimStyle(m.theme).Render("No activity data")
    }

    var blocks []string
    for _, day := range m.summaryData.Data {
        hours := day.GrandTotal.TotalSeconds / 3600.0
        color := getThemedActivityColor(hours, m.theme)
        label := day.Range.Date[5:] // MM-DD format
        block := lipgloss.NewStyle().
            Background(color).
            Foreground(m.theme.Foreground).
            Padding(0, 1).
            Render(label)
        blocks = append(blocks, block)
    }

    return lipgloss.JoinHorizontal(lipgloss.Top, blocks...)
}
```

### Example 5: Division by Zero Protection in Percentage Calculations

```go
// Source: Enhanced version of tui/stats_panels.go

func (m *Model) updateCategoriesChart() {
    if m.stats == nil {
        return
    }

    m.categoriesChart.Clear()
    data := m.stats.Data

    // Limit to top 10 categories
    limit := 10
    if len(data.Categories) < limit {
        limit = len(data.Categories)
    }

    // Calculate total for percentages
    var total float64
    for _, cat := range data.Categories {
        total += cat.TotalSeconds
    }

    // Protect against division by zero
    if total == 0 {
        // No data or all zeros - don't render chart
        return
    }

    // Add top 10 categories with safe percentage calculation
    var otherSeconds float64
    for i, cat := range data.Categories {
        if i < limit {
            hours := cat.TotalSeconds / 3600.0
            percent := (cat.TotalSeconds / total) * 100  // Safe: total > 0
            label := fmt.Sprintf("%s: %s", cat.Name, formatTimeWithPercent(cat.TotalSeconds, percent))
            barStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
            m.categoriesChart.Push(barchart.BarData{
                Label: label,
                Values: []barchart.BarValue{
                    {
                        Name:  "",
                        Value: hours,
                        Style: barStyle,
                    },
                },
            })
        } else {
            otherSeconds += cat.TotalSeconds
        }
    }

    // Add "Other" category if there are remaining items
    if len(data.Categories) > limit && otherSeconds > 0 {
        hours := otherSeconds / 3600.0
        percent := (otherSeconds / total) * 100
        label := fmt.Sprintf("Other: %s", formatTimeWithPercent(otherSeconds, percent))
        barStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
        m.categoriesChart.Push(barchart.BarData{
            Label: label,
            Values: []barchart.BarValue{
                {
                    Name:  "",
                    Value: hours,
                    Style: barStyle,
                },
            },
        })
    }

    m.categoriesChart.Draw()
}

// Apply same pattern to updateEditorsChart, updateOSChart, updateMachinesChart
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Generic error messages | Actionable error messages with hints | Go 1.13+ error wrapping (2019) | Users can self-recover without documentation |
| Panic on errors | Error values with graceful fallback | Go idiom since 1.0 (2012) | Applications stay running, terminals stay functional |
| Silent config failures | Logged warnings with suggestions | Modern logging practices (2020+) | Easier debugging, clearer user feedback |
| Terminal state loss on panic | BubbleTea panic recovery | BubbleTea v0.27+ (2024) | Terminal remains usable even after crashes |

**Deprecated/outdated:**
- **Panic-based error handling:** Never appropriate in Go; use error returns
- **Generic "something went wrong" messages:** Modern practice emphasizes specificity
- **Ignoring user config errors silently:** Users expect warnings for typos/mistakes

## Open Questions

1. **Should we add a --validate flag to check config without starting UI?**
   - What we know: Some CLI tools offer dry-run/validate modes
   - What's unclear: Whether wakadash users would benefit from pre-flight checks
   - Recommendation: Defer to future enhancement; current fallback behavior is sufficient

2. **How to handle extremely large terminal sizes (>200 cols)?**
   - What we know: Current responsive layout works for 40-120 cols
   - What's unclear: Whether ultra-wide terminals need special handling
   - Recommendation: Current 2-column layout scales fine; no changes needed

3. **Should missing data show "no data yet" vs "no data in range"?**
   - What we know: Empty arrays could mean new user OR empty time range
   - What's unclear: Whether distinction adds value vs confuses users
   - Recommendation: Use simple "No [category] tracked in this time range" for consistency

4. **Log to file vs stderr for config warnings?**
   - What we know: Current code uses stdlib log (goes to stderr)
   - What's unclear: Whether users want persistent logs for debugging
   - Recommendation: Keep stderr logging; add file logging only if users request it

## Sources

### Primary (HIGH confidence)
- [BubbleTea package - github.com/charmbracelet/bubbletea](https://pkg.go.dev/github.com/charmbracelet/bubbletea) - TUI framework error handling patterns
- [Go errors package](https://pkg.go.dev/errors) - Sentinel errors and error wrapping
- [Handling failures in Init · BubbleTea Discussion #623](https://github.com/charmbracelet/bubbletea/discussions/623) - Error handling in BubbleTea programs
- [Error handling in Go (Golang)](https://medium.com/@virtualik/error-handling-patterns-in-go-every-developer-should-know-8962777c935b) - Modern Go error patterns
- [wakadash existing codebase](/workspace/wakadash) - Current defensive patterns (terminal size, nil checks, theme fallback)

### Secondary (MEDIUM confidence)
- [Terminal settings not restored after panics · BubbleTea Issue #1459](https://github.com/charmbracelet/bubbletea/issues/1459) - Terminal state recovery issues
- [Tips for building Bubble Tea programs](https://leg100.github.io/en/posts/building-bubbletea-programs/) - Best practices for graceful degradation
- [Go gRPC error handling](https://oneuptime.com/blog/post/2026-01-07-go-grpc-error-handling/view) - Error code patterns applicable to TUIs
- [Fault-tolerant services with graceful degradation](https://oneuptime.com/blog/post/2026-01-25-fault-tolerant-graceful-degradation-go/view) - Fallback strategies in Go

### Tertiary (LOW confidence)
- Web search results on general error handling (used for pattern verification only)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All patterns use stdlib + existing dependencies
- Architecture: HIGH - BubbleTea docs provide clear guidance, existing code demonstrates patterns
- Pitfalls: HIGH - BubbleTea issues document terminal state recovery, existing code shows nil checks
- Error messages: MEDIUM - Best practices evolving; recommendations based on user feedback patterns

**Research date:** 2026-02-20
**Valid until:** 60 days (error handling patterns stable; BubbleTea API mature)
