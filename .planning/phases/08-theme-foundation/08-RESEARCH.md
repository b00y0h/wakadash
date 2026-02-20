# Phase 8: Theme Foundation - Research

**Researched:** 2026-02-20
**Domain:** Terminal UI theming with BubbleTea/lipgloss
**Confidence:** HIGH

## Summary

Theme systems in Go TUIs require carefully structured color palettes, runtime theme switching, and persistent configuration storage. Lipgloss provides the styling primitives, but applications must implement their own theme management layer. The wakadash codebase currently uses hardcoded colors scattered across styles.go, colors.go, and model.go that must be migrated to a centralized theme system.

The six target themes (Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night) are industry-standard color schemes with well-documented hex values. Each theme needs 8-12 semantic colors: background, foreground, accent colors for syntax/UI elements, and special colors for borders, selection, and comments.

For persistence, the existing ~/.wakatime.cfg INI file already exists and supports custom key-value pairs in the [settings] section. Adding a `theme` key is the natural extension point.

**Primary recommendation:** Create a centralized theme.go package with Theme structs containing semantic color fields (Background, Foreground, Primary, Secondary, Accent, Border, etc.). Inject the active theme into the Model at initialization. Persist theme selection to ~/.wakatime.cfg [settings] section. Build a full-screen theme picker using BubbleTea's standard view-switching pattern.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **Full-screen picker before dashboard loads (not an overlay)** - First-run experience shows picker before any dashboard content
- **Single theme preview at a time, arrow keys to browse** - One theme displayed with live preview, navigate with arrows
- **Preview shows mini dashboard with sample data** - Scaled-down version of actual panels with hardcoded sample data (no API calls)
- **Enter key confirms selection** - No number key shortcuts for theme selection
- **Press 't' from dashboard to open theme picker** - Runtime theme switching via keyboard shortcut
- **Same full-screen picker experience as first-run** - Consistent UX for both first-run and runtime switching
- **Dashboard resumes instantly with new theme applied** - No confirmation message after theme selection
- **'t - Change theme' shown in help overlay** - Document the shortcut in the '?' help screen
- **Full theming: borders, backgrounds, text all follow theme palette** - Comprehensive theme application
- **Heatmap uses theme's accent color gradient** - Not hardcoded GitHub green
- **Header/title bar fully themed** - Title, refresh indicator, status all use theme colors

### Claude's Discretion
- **Language bar chart colors** - Decide whether to use theme palette or keep GitHub Linguist colors based on visual balance
- **Exact mini dashboard layout in preview** - How to arrange/scale the preview panels
- **Theme palette structure** - How many colors per theme, naming conventions
- **Handling terminals with limited color support** - Fallback behavior for 8/16 color terminals

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| THEME-01 | User can choose from 6 theme presets (Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night) | Official color palettes documented below; Theme struct pattern from lipgloss-theme |
| THEME-02 | User sees visual theme preview on first run before selecting | BubbleTea view-switching pattern; Sample data rendering approach |
| THEME-03 | User's theme selection persists in ~/.wakatime.cfg | WakaTime config INI format supports custom [settings] keys |
| THEME-04 | All existing panels (Languages, Projects, Heatmap) use the selected theme colors | Theme injection via Model initialization; Replace hardcoded lipgloss.Color() calls |

</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/charmbracelet/lipgloss | v1.1.0+ | Terminal styling and colors | Already in use; industry standard for TUI styling |
| github.com/charmbracelet/bubbletea | v1.3.10+ | TUI framework | Already in use; handles view switching for theme picker |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| gopkg.in/ini.v1 | Latest | INI file parsing | Only if config.go's manual parsing becomes insufficient (currently not needed) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Manual INI parsing | gopkg.in/ini.v1 | More robust but adds dependency; current code is simple enough |
| Hardcoded themes | Dynamic theme loading from files | Adds complexity; 6 themes don't justify file-based config |
| lipgloss.AdaptiveColor | Manual terminal detection | AdaptiveColor auto-detects light/dark terminals; not needed for fixed dark themes |

**Installation:**
No new dependencies required. Lipgloss and BubbleTea already present in go.mod.

## Architecture Patterns

### Recommended Project Structure
```
wakadash/internal/
├── theme/              # NEW: Theme system
│   ├── theme.go        # Theme struct and GetTheme() function
│   ├── presets.go      # 6 theme definitions (Dracula, Nord, etc.)
│   └── config.go       # Read/write theme from ~/.wakatime.cfg
├── tui/
│   ├── model.go        # Model gains .theme field
│   ├── styles.go       # Styles become theme-aware functions
│   ├── colors.go       # Language colors (keep or theme-override)
│   ├── picker.go       # NEW: Full-screen theme picker
│   └── ...
└── config/
    └── config.go       # Extended to read/write theme setting
```

### Pattern 1: Centralized Theme Struct

**What:** Define a Theme type with semantic color fields that map to all UI elements.

**When to use:** When you need consistent colors across multiple UI components without hardcoding.

**Example:**
```go
// Source: Based on github.com/purpleclay/lipgloss-theme and github.com/willyv3/gogh-themes/lipgloss patterns

package theme

import "github.com/charmbracelet/lipgloss"

type Theme struct {
    Name string

    // Base colors
    Background   lipgloss.Color
    Foreground   lipgloss.Color

    // UI elements
    Border       lipgloss.Color
    Title        lipgloss.Color
    Dim          lipgloss.Color

    // Status colors
    Error        lipgloss.Color
    Warning      lipgloss.Color
    Success      lipgloss.Color

    // Accent colors (for charts, highlights)
    Primary      lipgloss.Color
    Secondary    lipgloss.Color
    Accent1      lipgloss.Color
    Accent2      lipgloss.Color
    Accent3      lipgloss.Color
    Accent4      lipgloss.Color
}

// GetTheme returns a theme by name
func GetTheme(name string) Theme {
    switch name {
    case "dracula":
        return draculaTheme
    case "nord":
        return nordTheme
    // ... etc
    default:
        return draculaTheme // sensible default
    }
}
```

### Pattern 2: Theme-Aware Style Functions

**What:** Replace global style variables with functions that accept a Theme parameter.

**When to use:** When styles need to change based on runtime theme selection.

**Example:**
```go
// Source: Inferred from lipgloss best practices

// OLD (hardcoded):
var titleStyle = lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("205"))

// NEW (theme-aware):
func TitleStyle(theme Theme) lipgloss.Style {
    return lipgloss.NewStyle().
        Bold(true).
        Foreground(theme.Title)
}

// Usage in View():
title := TitleStyle(m.theme).Render("WakaTime Stats")
```

### Pattern 3: Config Persistence with INI Format

**What:** Add custom settings to WakaTime's existing ~/.wakatime.cfg file in [settings] section.

**When to use:** When you need to persist user preferences alongside existing WakaTime config.

**Example:**
```go
// Source: Based on WakaTime config format from github.com/wakatime/wakatime-cli

// ~/.wakatime.cfg format:
// [settings]
// api_url = https://api.wakatime.com/api/v1
// api_key = waka_xxxx
// theme = dracula    # <-- NEW FIELD

// Read theme:
func LoadTheme() string {
    configPath := filepath.Join(os.UserHomeDir(), ".wakatime.cfg")
    file, _ := os.Open(configPath)
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if key, value, found := strings.Cut(line, "="); found {
            if strings.TrimSpace(key) == "theme" {
                return strings.TrimSpace(value)
            }
        }
    }
    return "" // No theme set (first run)
}

// Write theme (append or update in-place):
func SaveTheme(themeName string) error {
    // Read entire file, update theme= line or append, write back
    // Follow same pattern as existing config.Load() in config/config.go
}
```

### Pattern 4: Full-Screen Theme Picker with Live Preview

**What:** A BubbleTea Model that takes over the entire screen to show theme previews.

**When to use:** For first-run theme selection or runtime theme switching (user presses 't').

**Example:**
```go
// Source: BubbleTea view-switching pattern from official examples

type PickerModel struct {
    themes       []string  // ["dracula", "nord", "gruvbox", ...]
    selectedIdx  int
    width, height int
}

func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            m.selectedIdx = (m.selectedIdx - 1 + len(m.themes)) % len(m.themes)
        case "down", "j":
            m.selectedIdx = (m.selectedIdx + 1) % len(m.themes)
        case "enter":
            // Save theme and switch back to dashboard
            themeName := m.themes[m.selectedIdx]
            SaveTheme(themeName)
            return mainDashboard, nil  // Return to dashboard view
        }
    }
    return m, nil
}

func (m PickerModel) View() string {
    currentTheme := GetTheme(m.themes[m.selectedIdx])

    // Render a mini dashboard with sample data styled with currentTheme
    preview := renderMiniDashboard(currentTheme)

    // Add theme name label
    label := fmt.Sprintf("Theme: %s (↑/↓ to browse, Enter to select)", currentTheme.Name)

    return lipgloss.JoinVertical(lipgloss.Left, label, preview)
}
```

### Pattern 5: Heatmap Color Gradient from Theme

**What:** Generate activity heatmap colors based on theme's accent color rather than hardcoded green.

**When to use:** When adapting GitHub-style heatmaps to match theme palette.

**Example:**
```go
// Source: Current getActivityColor() pattern adapted for themes

func getActivityColor(hours float64, theme Theme) lipgloss.Color {
    // Use theme's Primary accent color and create gradient
    baseColor := theme.Primary  // e.g., Dracula purple #bd93f9

    // Create darkened variants for lower activity
    // For simplicity, use predefined light/medium/dark variants in Theme struct
    switch {
    case hours < 0.5:
        return theme.Background  // Matches background
    case hours < 2:
        return darken(theme.Primary, 0.7)  // Dark variant
    case hours < 4:
        return darken(theme.Primary, 0.5)  // Medium variant
    case hours < 6:
        return theme.Primary
    default:
        return lighten(theme.Primary, 1.2)  // Bright variant
    }
}

// Alternative: Define heatmap colors explicitly in Theme struct
type Theme struct {
    // ... other fields
    HeatmapColors [5]lipgloss.Color  // None, Low, Medium, High, VeryHigh
}
```

### Anti-Patterns to Avoid

- **Scattering color definitions across files:** All theme colors should be in theme/presets.go
- **Using ANSI color codes (e.g., "205") in themes:** Use hex colors for consistency and predictability
- **Creating styles globally at package init:** Styles need theme parameter, must be functions or methods
- **Mutating global theme state:** Pass theme as Model field, not global variable
- **Over-engineering with file-based theme loading:** 6 hardcoded themes are simpler and faster

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Color degradation for terminal profiles | Manual 256/16/8 color fallbacks | lipgloss automatic downsampling | Lipgloss detects terminal capabilities and converts colors automatically |
| Terminal background detection | Manual ANSI queries | lipgloss.AdaptiveColor (if needed) | Not needed for fixed dark themes; lipgloss handles it if adding light theme support later |
| INI file parsing with sections | Full INI parser library | Manual line-by-line parsing (already in config.go) | Existing pattern works; file is simple; no nested sections needed |
| Dynamic color interpolation | Custom color math library | Predefined shade variants in Theme struct | Heatmap needs 4-5 colors, not continuous gradients; hardcode them |

**Key insight:** Lipgloss handles terminal color profile complexity. You only need to define hex colors in Theme structs and let lipgloss handle downsampling to 256/16/8 color terminals. Don't write color conversion code.

## Common Pitfalls

### Pitfall 1: AdaptiveColor Causing Startup Hang

**What goes wrong:** Using lipgloss.AdaptiveColor without proper initialization can cause the application to hang on startup while detecting terminal background.

**Why it happens:** BubbleTea v0.27.0 and earlier had a race condition with background detection. As of v0.27.1+, this is fixed, but it requires calling colorprofile detection early.

**How to avoid:** Wakadash uses BubbleTea v1.3.10+ which has the fix. However, since all 6 target themes are dark themes, avoid AdaptiveColor entirely. Use fixed hex colors.

**Warning signs:** App hangs on startup before rendering anything; terminal seems frozen.

**Source:** Mentioned in .planning/STATE.md research flags

### Pitfall 2: Hardcoded Colors Persisting After Theme Switch

**What goes wrong:** Some UI elements don't update when theme changes because they use hardcoded lipgloss.Color() values instead of theme references.

**Why it happens:** Global style variables initialized at package load time can't react to runtime theme changes.

**How to avoid:**
1. Replace all global `var xStyle = lipgloss.NewStyle()...` with functions `func XStyle(theme Theme) lipgloss.Style`
2. Audit all files for `lipgloss.Color("...")` or `lipgloss.Color("#...")` calls
3. Replace with theme field references: `theme.Border`, `theme.Title`, etc.

**Warning signs:** After selecting a new theme, some panels/text colors don't change. Grep for `lipgloss.Color(` to find violations.

### Pitfall 3: Dynamic Width Styles Causing Rendering Corruption

**What goes wrong:** Using `.Width()` with dynamic values can cause visual corruption when terminal resizes or content changes.

**Why it happens:** Lipgloss caching and BubbleTea rendering can conflict when widths change between frames.

**How to avoid:** Use `.MaxWidth()` instead of `.Width()` for dynamic layouts. Only use `.Width()` for fixed-size elements.

**Warning signs:** Boxes overflow, text wraps incorrectly, or borders misalign after terminal resize.

**Source:** Mentioned in .planning/STATE.md v2.1 integration risks

### Pitfall 4: First-Run Detection Failing

**What goes wrong:** Theme picker doesn't show on first run because "no theme set" detection fails.

**Why it happens:** Empty string vs missing key ambiguity in INI parsing; or leftover blank `theme=` line from failed previous run.

**How to avoid:**
1. Treat both empty string and missing key as "no theme set"
2. Add explicit first-run flag: `theme_configured = true` after first selection
3. Check for valid theme name, not just presence of key

**Warning signs:** User sees default theme without being prompted to choose; or picker shows every time despite having selected a theme.

### Pitfall 5: Language Bar Colors Not Visually Distinct in Some Themes

**What goes wrong:** GitHub Linguist colors (cyan Go, yellow JavaScript, etc.) may clash with theme colors or lose contrast.

**Why it happens:** Linguist colors were designed for GitHub's light/dark UI, not for arbitrary theme palettes. For example, Nord's accent colors are all cool blues/greens, but Go's Linguist color is also cyan.

**How to avoid:**
- **Option A (Recommended):** Keep Linguist colors for familiarity; they're well-known and tested
- **Option B:** Define language colors per theme in Theme struct (LanguageColors map[string]lipgloss.Color)
- **Test each theme:** Render sample language chart and verify visual distinction

**Warning signs:** Two languages appear nearly identical in color; language chart looks washed out or too similar.

## Code Examples

Verified patterns from research and existing codebase:

### Example 1: Defining Theme Presets

```go
// Source: Official color palettes from theme documentation

package theme

import "github.com/charmbracelet/lipgloss"

var Dracula = Theme{
    Name:       "Dracula",
    Background: lipgloss.Color("#282a36"),
    Foreground: lipgloss.Color("#f8f8f2"),
    Border:     lipgloss.Color("#6272a4"),  // Comment color
    Title:      lipgloss.Color("#bd93f9"),  // Purple
    Dim:        lipgloss.Color("#6272a4"),  // Comment color
    Error:      lipgloss.Color("#ff5555"),  // Red
    Warning:    lipgloss.Color("#ffb86c"),  // Orange
    Success:    lipgloss.Color("#50fa7b"),  // Green
    Primary:    lipgloss.Color("#bd93f9"),  // Purple
    Secondary:  lipgloss.Color("#8be9fd"),  // Cyan
    Accent1:    lipgloss.Color("#ff79c6"),  // Pink
    Accent2:    lipgloss.Color("#f1fa8c"),  // Yellow
    Accent3:    lipgloss.Color("#50fa7b"),  // Green
    Accent4:    lipgloss.Color("#ffb86c"),  // Orange
}

var Nord = Theme{
    Name:       "Nord",
    Background: lipgloss.Color("#2e3440"),  // nord0
    Foreground: lipgloss.Color("#d8dee9"),  // nord4
    Border:     lipgloss.Color("#4c566a"),  // nord3
    Title:      lipgloss.Color("#88c0d0"),  // nord8 (frost)
    Dim:        lipgloss.Color("#4c566a"),  // nord3
    Error:      lipgloss.Color("#bf616a"),  // nord11 (aurora red)
    Warning:    lipgloss.Color("#ebcb8b"),  // nord13 (aurora yellow)
    Success:    lipgloss.Color("#a3be8c"),  // nord14 (aurora green)
    Primary:    lipgloss.Color("#88c0d0"),  // nord8
    Secondary:  lipgloss.Color("#81a1c1"),  // nord9
    Accent1:    lipgloss.Color("#8fbcbb"),  // nord7
    Accent2:    lipgloss.Color("#5e81ac"),  // nord10
    Accent3:    lipgloss.Color("#b48ead"),  // nord15 (aurora purple)
    Accent4:    lipgloss.Color("#d08770"),  // nord12 (aurora orange)
}

var Gruvbox = Theme{
    Name:       "Gruvbox",
    Background: lipgloss.Color("#282828"),  // dark0
    Foreground: lipgloss.Color("#ebdbb2"),  // light1
    Border:     lipgloss.Color("#504945"),  // dark2
    Title:      lipgloss.Color("#fabd2f"),  // bright yellow
    Dim:        lipgloss.Color("#665c54"),  // dark3
    Error:      lipgloss.Color("#fb4934"),  // bright red
    Warning:    lipgloss.Color("#fe8019"),  // bright orange
    Success:    lipgloss.Color("#b8bb26"),  // bright green
    Primary:    lipgloss.Color("#83a598"),  // bright blue
    Secondary:  lipgloss.Color("#d3869b"),  // bright purple
    Accent1:    lipgloss.Color("#8ec07c"),  // bright aqua
    Accent2:    lipgloss.Color("#fabd2f"),  // bright yellow
    Accent3:    lipgloss.Color("#fb4934"),  // bright red
    Accent4:    lipgloss.Color("#fe8019"),  // bright orange
}

var Monokai = Theme{
    Name:       "Monokai",
    Background: lipgloss.Color("#272822"),
    Foreground: lipgloss.Color("#f8f8f2"),
    Border:     lipgloss.Color("#49483e"),
    Title:      lipgloss.Color("#66d9ef"),  // Cyan
    Dim:        lipgloss.Color("#75715e"),  // Comment
    Error:      lipgloss.Color("#f92672"),  // Pink/Red
    Warning:    lipgloss.Color("#fd971f"),  // Orange
    Success:    lipgloss.Color("#a6e22e"),  // Green
    Primary:    lipgloss.Color("#ae81ff"),  // Purple
    Secondary:  lipgloss.Color("#66d9ef"),  // Cyan
    Accent1:    lipgloss.Color("#f92672"),  // Pink
    Accent2:    lipgloss.Color("#fd971f"),  // Orange
    Accent3:    lipgloss.Color("#a6e22e"),  // Green
    Accent4:    lipgloss.Color("#e6db74"),  // Yellow
}

var Solarized = Theme{
    Name:       "Solarized",
    Background: lipgloss.Color("#002b36"),  // base03
    Foreground: lipgloss.Color("#839496"),  // base0
    Border:     lipgloss.Color("#073642"),  // base02
    Title:      lipgloss.Color("#268bd2"),  // blue
    Dim:        lipgloss.Color("#586e75"),  // base01
    Error:      lipgloss.Color("#dc322f"),  // red
    Warning:    lipgloss.Color("#cb4b16"),  // orange
    Success:    lipgloss.Color("#859900"),  // green
    Primary:    lipgloss.Color("#268bd2"),  // blue
    Secondary:  lipgloss.Color("#2aa198"),  // cyan
    Accent1:    lipgloss.Color("#6c71c4"),  // violet
    Accent2:    lipgloss.Color("#b58900"),  // yellow
    Accent3:    lipgloss.Color("#859900"),  // green
    Accent4:    lipgloss.Color("#d33682"),  // magenta
}

var TokyoNight = Theme{
    Name:       "Tokyo Night",
    Background: lipgloss.Color("#1a1b26"),  // Storm variant
    Foreground: lipgloss.Color("#c0caf5"),
    Border:     lipgloss.Color("#3b4261"),
    Title:      lipgloss.Color("#7aa2f7"),  // Blue
    Dim:        lipgloss.Color("#565f89"),
    Error:      lipgloss.Color("#f7768e"),  // Red
    Warning:    lipgloss.Color("#e0af68"),  // Yellow
    Success:    lipgloss.Color("#9ece6a"),  // Green
    Primary:    lipgloss.Color("#bb9af7"),  // Purple
    Secondary:  lipgloss.Color("#7aa2f7"),  // Blue
    Accent1:    lipgloss.Color("#73daca"),  // Cyan
    Accent2:    lipgloss.Color("#ff9e64"),  // Orange
    Accent3:    lipgloss.Color("#9ece6a"),  // Green
    Accent4:    lipgloss.Color("#2ac3de"),  // Teal
}

// GetTheme returns a theme by name
func GetTheme(name string) Theme {
    switch name {
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
        return Dracula  // Default to Dracula
    }
}

// AllThemes returns theme names in display order
func AllThemes() []string {
    return []string{"dracula", "nord", "gruvbox", "monokai", "solarized", "tokyonight"}
}
```

### Example 2: Migrating Global Styles to Theme-Aware Functions

```go
// Source: Current styles.go adapted for theme support

package tui

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/b00y0h/wakadash/internal/theme"
)

// OLD (before):
// var borderStyle = lipgloss.NewStyle().
//     Border(lipgloss.RoundedBorder()).
//     BorderForeground(lipgloss.Color("62"))

// NEW (after):
func BorderStyle(t theme.Theme) lipgloss.Style {
    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(t.Border)
}

func TitleStyle(t theme.Theme) lipgloss.Style {
    return lipgloss.NewStyle().
        Bold(true).
        Foreground(t.Title)
}

func DimStyle(t theme.Theme) lipgloss.Style {
    return lipgloss.NewStyle().
        Foreground(t.Dim)
}

func ErrorStyle(t theme.Theme) lipgloss.Style {
    return lipgloss.NewStyle().
        Foreground(t.Error)
}

func WarningStyle(t theme.Theme) lipgloss.Style {
    return lipgloss.NewStyle().
        Foreground(t.Warning)
}
```

### Example 3: Config Persistence (Read/Write Theme)

```go
// Source: Adapted from existing config/config.go pattern

package theme

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// LoadThemeFromConfig reads the theme name from ~/.wakatime.cfg
// Returns empty string if no theme is set (first run)
func LoadThemeFromConfig() (string, error) {
    configPath, err := configFilePath()
    if err != nil {
        return "", err
    }

    f, err := os.Open(configPath)
    if err != nil {
        return "", err
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())

        // Skip comments and blank lines
        if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
            continue
        }

        key, value, found := strings.Cut(line, "=")
        if !found {
            continue
        }

        key = strings.TrimSpace(key)
        value = strings.TrimSpace(value)

        if key == "theme" {
            return value, nil
        }
    }

    if err := scanner.Err(); err != nil {
        return "", err
    }

    return "", nil  // No theme key found (first run)
}

// SaveThemeToConfig writes the theme name to ~/.wakatime.cfg
// Updates existing theme= line or appends if not found
func SaveThemeToConfig(themeName string) error {
    configPath, err := configFilePath()
    if err != nil {
        return err
    }

    // Read entire file
    data, err := os.ReadFile(configPath)
    if err != nil && !os.IsNotExist(err) {
        return err
    }

    lines := strings.Split(string(data), "\n")
    updated := false

    // Find and update existing theme= line
    for i, line := range lines {
        trimmed := strings.TrimSpace(line)
        if strings.HasPrefix(trimmed, "theme") {
            if key, _, found := strings.Cut(trimmed, "="); found && strings.TrimSpace(key) == "theme" {
                lines[i] = fmt.Sprintf("theme = %s", themeName)
                updated = true
                break
            }
        }
    }

    // Append if not found
    if !updated {
        // Ensure file ends with newline
        if len(lines) > 0 && lines[len(lines)-1] != "" {
            lines = append(lines, "")
        }
        lines = append(lines, fmt.Sprintf("theme = %s", themeName))
    }

    // Write back
    return os.WriteFile(configPath, []byte(strings.Join(lines, "\n")), 0600)
}

func configFilePath() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(homeDir, ".wakatime.cfg"), nil
}
```

### Example 4: Theme Picker Model

```go
// Source: BubbleTea view-switching pattern

package tui

import (
    "fmt"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/b00y0h/wakadash/internal/theme"
)

type ThemePickerModel struct {
    themes      []string
    selectedIdx int
    width       int
    height      int
    confirmed   bool
    themeName   string  // Selected theme name
}

func NewThemePicker() ThemePickerModel {
    return ThemePickerModel{
        themes:      theme.AllThemes(),
        selectedIdx: 0,
    }
}

func (m ThemePickerModel) Init() tea.Cmd {
    return nil
}

func (m ThemePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        return m, nil

    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            m.selectedIdx = (m.selectedIdx - 1 + len(m.themes)) % len(m.themes)
        case "down", "j":
            m.selectedIdx = (m.selectedIdx + 1) % len(m.themes)
        case "enter":
            m.confirmed = true
            m.themeName = m.themes[m.selectedIdx]
            // Save theme to config
            theme.SaveThemeToConfig(m.themeName)
            // Signal to main app to switch to dashboard
            return m, tea.Quit
        case "q", "esc":
            // Cancel - use default theme
            m.confirmed = true
            m.themeName = "dracula"
            return m, tea.Quit
        }
        return m, nil
    }

    return m, nil
}

func (m ThemePickerModel) View() string {
    if m.width < 40 || m.height < 10 {
        return "Terminal too small for theme preview"
    }

    // Get current theme
    currentTheme := theme.GetTheme(m.themes[m.selectedIdx])

    // Render mini dashboard preview with sample data
    preview := m.renderMiniDashboard(currentTheme)

    // Theme name label
    label := fmt.Sprintf("\nTheme: %s\n↑/↓ to browse, Enter to select\n", currentTheme.Name)
    labelStyled := lipgloss.NewStyle().
        Foreground(currentTheme.Title).
        Bold(true).
        Render(label)

    return lipgloss.JoinVertical(lipgloss.Left, labelStyled, preview)
}

func (m ThemePickerModel) renderMiniDashboard(t theme.Theme) string {
    // Render scaled-down version of actual dashboard with sample data
    var sb strings.Builder

    // Title
    title := lipgloss.NewStyle().
        Bold(true).
        Foreground(t.Title).
        Render("WakaTime Stats (last_7_days)")
    sb.WriteString(title + "\n\n")

    // Sample stats
    sb.WriteString(fmt.Sprintf("  Total time:    %s\n", "42h 15m"))
    sb.WriteString(fmt.Sprintf("  Daily average: %s\n", "6h 2m"))
    sb.WriteString("\n")

    // Sample language bar (simplified)
    langLabel := lipgloss.NewStyle().Foreground(t.Title).Render("Languages")
    sb.WriteString(langLabel + "\n")
    sb.WriteString(m.renderSampleBar("Go", 15, t.Accent1) + "\n")
    sb.WriteString(m.renderSampleBar("TypeScript", 10, t.Accent2) + "\n")
    sb.WriteString(m.renderSampleBar("Python", 5, t.Accent3) + "\n")
    sb.WriteString("\n")

    // Sample heatmap blocks
    heatLabel := lipgloss.NewStyle().Foreground(t.Title).Render("Activity (Last 7 Days)")
    sb.WriteString(heatLabel + "\n")

    blocks := []string{}
    for i := 0; i < 7; i++ {
        intensity := (i % 4) + 1  // Vary intensity
        var color lipgloss.Color
        switch intensity {
        case 1:
            color = t.Background
        case 2:
            color = t.Accent1
        case 3:
            color = t.Primary
        default:
            color = t.Secondary
        }
        block := lipgloss.NewStyle().
            Background(color).
            Foreground(t.Foreground).
            Padding(0, 1).
            Render(fmt.Sprintf("D%d", i+1))
        blocks = append(blocks, block)
    }
    sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, blocks...))

    // Wrap in border
    content := sb.String()
    bordered := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(t.Border).
        Padding(1, 2).
        Render(content)

    return bordered
}

func (m ThemePickerModel) renderSampleBar(name string, value int, color lipgloss.Color) string {
    bar := strings.Repeat("█", value)
    barStyled := lipgloss.NewStyle().Foreground(color).Render(bar)
    return fmt.Sprintf("  %-12s %s", name, barStyled)
}
```

### Example 5: Integrating Theme into Main Model

```go
// Source: Existing model.go pattern extended

package tui

import (
    "github.com/b00y0h/wakadash/internal/theme"
    // ... other imports
)

type Model struct {
    // ... existing fields
    theme theme.Theme  // NEW: Active theme
}

func NewModel(client *api.Client, rangeStr string, refreshInterval time.Duration) Model {
    // Load theme from config
    themeName, _ := theme.LoadThemeFromConfig()
    if themeName == "" {
        themeName = "dracula"  // Default if not set
    }
    activeTheme := theme.GetTheme(themeName)

    // ... existing initialization

    return Model{
        // ... existing fields
        theme: activeTheme,
    }
}

// In View() methods, replace hardcoded styles with theme-aware calls:
func (m Model) renderDashboard() string {
    // OLD:
    // statsPanel := borderStyle.Width(m.width - 2).Render(content)

    // NEW:
    statsPanel := BorderStyle(m.theme).
        Width(m.width - 2).
        Height(panelHeight).
        Render(content)

    return lipgloss.JoinVertical(lipgloss.Left, statsPanel, statusBar)
}

// Update() handles 't' key to open theme picker:
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        // ... existing keys
        case key.Matches(msg, m.keys.ThemePicker):  // 't' key
            // Switch to theme picker view
            return NewThemePicker(), nil
        }
    }
    return m, nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Global hardcoded ANSI colors | Hex colors with automatic downsampling | lipgloss v0.5+ (2021) | Predictable colors across terminals; no manual fallback code |
| tview's color system | lipgloss styles | BubbleTea ecosystem (2020+) | Composable styles; better theme support |
| Manual INI parsing | gopkg.in/ini.v1 library | Not applicable (wakadash already uses manual parsing) | Simpler for .wakatime.cfg; no change needed |
| Light/dark auto-detection with AdaptiveColor | Fixed dark themes | Current trend (2024+) | Dark themes dominant in dev tools; simpler implementation |

**Deprecated/outdated:**
- **tui-go:** Unmaintained since 2019; replaced by BubbleTea and rivo/tview
- **termui:** Less active; BubbleTea more popular for Go TUIs in 2024+
- **AdaptiveColor for theme systems:** Useful for light/dark detection but not needed when offering 6 fixed dark themes

## Open Questions

1. **Should language bar colors use theme palette or keep GitHub Linguist colors?**
   - What we know: Current code uses hardcoded Linguist colors (Go=#00ADD8, Python=#3572A5, etc.)
   - What's unclear: Whether theme-specific language colors improve or harm UX
   - Recommendation: Start with Linguist colors (user expectation), add theme-specific override as Phase 9 enhancement if user feedback requests it

2. **How to handle heatmap color gradients across different theme palettes?**
   - What we know: Current heatmap uses GitHub green (#0e4429 to #39d353)
   - What's unclear: Whether to use theme's Primary color or define explicit HeatmapColors array in Theme struct
   - Recommendation: Define 4-5 explicit heatmap colors per theme (HeatmapNone, Low, Medium, High, VeryHigh) for consistent visual weight

3. **Should first-run detection be explicit flag or inferred from missing theme key?**
   - What we know: Missing theme= key indicates first run
   - What's unclear: Whether to add `theme_configured = true` flag for robustness
   - Recommendation: Use simple "empty or missing theme key" check initially; add explicit flag only if first-run detection proves unreliable in testing

4. **What happens if user manually edits ~/.wakatime.cfg with invalid theme name?**
   - What we know: GetTheme() has default case returning Dracula
   - What's unclear: Whether to show error message or silently fall back
   - Recommendation: Silent fallback to Dracula; optionally log warning. Don't block app startup on invalid config.

## Sources

### Primary (HIGH confidence)
- [lipgloss package - github.com/charmbracelet/lipgloss](https://pkg.go.dev/github.com/charmbracelet/lipgloss) - Core styling API
- [BubbleTea package - github.com/charmbracelet/bubbletea](https://pkg.go.dev/github.com/charmbracelet/bubbletea) - TUI framework patterns
- [lipgloss-theme package - github.com/purpleclay/lipgloss-theme](https://pkg.go.dev/github.com/purpleclay/lipgloss-theme) - Theme struct pattern reference
- [WakaTime CLI USAGE.md](https://github.com/wakatime/wakatime-cli/blob/develop/USAGE.md) - Official config file format
- [Dracula Theme Official](https://github.com/dracula/dracula-theme) - Official color palette
- [Nord Theme Official](https://www.nordtheme.com) - Official color palette
- [Solarized Official](https://ethanschoonover.com/solarized/) - Ethan Schoonover's official specification
- [Gruvbox Official](https://github.com/morhetz/gruvbox) - Original Vim theme repository

### Secondary (MEDIUM confidence)
- [gogh-themes/lipgloss](https://pkg.go.dev/github.com/willyv3/gogh-themes/lipgloss) - Wrapped color schemes as lipgloss.Color
- [Tokyo Night Theme](https://github.com/folke/tokyonight.nvim) - Popular theme with multiple variants
- [Monokai colors from Sublime Text 3](https://gist.github.com/r-malon/8fc669332215c8028697a0bbfbfbb32a) - Classic color values
- [Building TUI with BubbleTea article](https://themarkokovacevic.com/posts/terminal-ui-with-bubbletea/) - Implementation patterns
- [Gogh Color Schemes](https://gogh-co.github.io/Gogh/) - Terminal color scheme collection

### Tertiary (LOW confidence)
- Various theme palette sites (color-hex.com) - Used for verification only; official sources preferred

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - lipgloss and BubbleTea already in use; patterns verified in official docs
- Architecture: HIGH - Theme struct pattern seen in multiple projects; config persistence follows existing code
- Pitfalls: MEDIUM - AdaptiveColor hang documented in STATE.md; other pitfalls inferred from lipgloss issues/PRs
- Color palettes: HIGH - Official theme documentation provides exact hex values for 5 of 6 themes; Tokyo Night partially verified

**Research date:** 2026-02-20
**Valid until:** 60 days (theme colors are stable; lipgloss API is mature and stable)
