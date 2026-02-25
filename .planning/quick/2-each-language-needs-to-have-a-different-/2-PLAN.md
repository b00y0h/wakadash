---
phase: quick-2
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/stats_panels.go
  - internal/tui/colors.go
autonomous: true
must_haves:
  truths:
    - "Each language in the Languages panel displays its bar in a distinct color from the theme accent palette"
    - "Unknown languages fall back to a theme accent color instead of all being the same"
    - "All other panels (Projects, Categories, Editors, OS, Machines) remain unchanged"
  artifacts:
    - path: "internal/tui/stats_panels.go"
      provides: "Per-language colored bar rendering"
      contains: "getLanguageColor"
    - path: "internal/tui/colors.go"
      provides: "Language-to-color mapping with theme fallback"
      contains: "languageColors"
  key_links:
    - from: "internal/tui/stats_panels.go"
      to: "internal/tui/colors.go"
      via: "getLanguageColor call per language item"
      pattern: "getLanguageColor\\("
---

<objective>
Give each language in the Languages panel its own distinct bar color using the existing GitHub Linguist color map in colors.go.

Purpose: Currently all language bars render in a single theme.Primary color, making it hard to visually distinguish languages at a glance. The colors.go file already has a languageColors map with 20 GitHub Linguist colors and a getLanguageColor() function, but renderLanguagesPanel() never uses it.

Output: Languages panel with per-language colored bars; all other panels unchanged.
</objective>

<execution_context>
@/Users/BobbySmith/.claude/get-shit-done/workflows/execute-plan.md
@/Users/BobbySmith/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/tui/stats_panels.go
@internal/tui/colors.go
@internal/theme/theme.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add per-item color support to bar chart and wire language colors</name>
  <files>internal/tui/stats_panels.go, internal/tui/colors.go</files>
  <action>
  In stats_panels.go, create a new function `renderColoredBarChart` that renders bars where each item has its own color. The approach:

  1. Add a `color` field to the `barItem` struct:
     ```go
     type barItem struct {
         name    string
         seconds float64
         color   lipgloss.Color // optional per-item color; zero value means use default
     }
     ```

  2. Create `renderColoredBarChart(items []barItem, maxSeconds float64, fallbackColor lipgloss.Color, panelWidth int) string` that works identically to `renderBarChart` except: inside the item loop, if `item.color != ""` use `item.color` for the barStyle, otherwise use `fallbackColor`. This keeps the existing `renderBarChart` callers unaffected since the zero value of `lipgloss.Color` is `""`.

     Actually, the simpler approach: modify `renderBarChart` itself. Since `barItem.color` zero value is `""`, check inside the loop: if `item.color != ""` use it, else use the passed `barColor`. This is backward-compatible because all existing callers create barItems without setting `.color`, so they get the default.

  3. In `renderLanguagesPanel`, after calling `getTopItems()`, iterate through the returned `items` slice and set each item's color using `getLanguageColor(item.name)` from colors.go. The "Other" aggregation item should use a theme accent color (e.g., `m.theme.Accent4`) since it represents multiple languages.

     ```go
     for i := range items {
         if items[i].name == "Other" {
             items[i].color = m.theme.Accent4
         } else {
             items[i].color = getLanguageColor(items[i].name)
         }
     }
     ```

  4. In `getLanguageColor` in colors.go, update the fallback for unknown languages. Instead of returning a hardcoded "#cccccc" gray, keep the current behavior (the theme accent colors are not accessible from colors.go without passing the theme, so "#cccccc" is fine as a visible neutral fallback for unknown languages).

  No changes needed to any other panel rendering functions (Projects, Categories, Editors, OS, Machines) - they will continue to work exactly as before since they never set `.color` on barItems.
  </action>
  <verify>
  Run `cd /Users/BobbySmith/source/github/b00y0h/wakafetch-brew/wakadash && go build ./...` to confirm compilation succeeds.
  Run `go vet ./...` to check for issues.
  Run `go test ./...` to ensure existing tests pass.
  Grep to confirm: `grep -c "getLanguageColor" internal/tui/stats_panels.go` returns >= 1 (language colors are wired in).
  Grep to confirm: `grep "item.color" internal/tui/stats_panels.go` shows the per-item color check in renderBarChart.
  </verify>
  <done>
  The Languages panel renders each language bar in its GitHub Linguist color (Go=#00ADD8 cyan, Python=#3572A5 blue, TypeScript=#3178c6, etc.). The "Other" row uses theme.Accent4. All other panels (Projects, Categories, Editors, OS, Machines) continue rendering with their existing single-color behavior. The app compiles and all tests pass.
  </done>
</task>

</tasks>

<verification>
- `go build ./...` compiles without errors
- `go vet ./...` reports no issues
- `go test ./...` all tests pass
- barItem struct has a `color` field
- renderBarChart uses per-item color when set
- renderLanguagesPanel calls getLanguageColor for each language
- No other panel functions were modified (Projects, Categories, Editors, OS, Machines unchanged)
</verification>

<success_criteria>
Each language in the Languages dashboard panel displays its bar in a unique color matching GitHub Linguist conventions. Unknown languages show a neutral gray. The "Other" aggregated row uses a theme accent. All other panels are visually unchanged.
</success_criteria>

<output>
After completion, create `.planning/quick/2-each-language-needs-to-have-a-different-/2-SUMMARY.md`
</output>
