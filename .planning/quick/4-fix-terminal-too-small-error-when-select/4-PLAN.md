---
phase: quick-4
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/model.go
  - internal/tui/picker.go
autonomous: true

must_haves:
  truths:
    - "Pressing 't' opens theme picker with correct preview, never shows 'terminal too small'"
    - "Theme picker still shows size warning if terminal is genuinely too small"
    - "Weekly browser also receives current dimensions on creation for consistency"
  artifacts:
    - path: "internal/tui/model.go"
      provides: "Passes current width/height to sub-models on creation"
    - path: "internal/tui/picker.go"
      provides: "ThemePickerModel accepts initial dimensions"
  key_links:
    - from: "internal/tui/model.go"
      to: "internal/tui/picker.go"
      via: "NewThemePicker constructor"
      pattern: "NewThemePicker.*width.*height"
---

<objective>
Fix bug where pressing 't' to open theme picker always shows "Terminal Too Small. Please resize." regardless of actual terminal size.

Purpose: The ThemePickerModel is created with zero-value width/height (Go struct defaults). Its View() method checks `if m.width < 40 || m.height < 10` which is always true at 0x0. The parent model has the correct dimensions in m.width/m.height but never passes them to the picker on creation. The fix is to pass current terminal dimensions when constructing the picker.

Output: Working theme picker that opens correctly when 't' is pressed.
</objective>

<execution_context>
@/Users/BobbySmith/.claude/get-shit-done/workflows/execute-plan.md
@/Users/BobbySmith/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/tui/model.go
@internal/tui/picker.go
@internal/tui/weekly_browser.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Pass terminal dimensions to sub-models on creation</name>
  <files>internal/tui/picker.go, internal/tui/model.go</files>
  <action>
  Bug analysis:
  - `picker.go` line 80: `View()` returns "Terminal too small" when `m.width < 40 || m.height < 10`
  - `picker.go` line 29-34: `NewThemePicker()` creates a model with width=0, height=0 (zero values)
  - `model.go` line 335-336: `ChangeTheme` handler creates picker but never sets width/height
  - The picker only receives dimensions if a `tea.WindowSizeMsg` arrives AFTER creation, which typically only happens on terminal resize

  Fix in `picker.go`:
  1. Update `NewThemePicker` signature to accept width and height parameters: `func NewThemePicker(isFirstRun bool, width, height int) ThemePickerModel`
  2. Set `width` and `height` fields in the returned struct from these parameters

  Fix in `model.go`:
  1. In the `ChangeTheme` handler (around line 335), update the call to: `NewThemePicker(false, m.width, m.height)`
  2. Also set `m.picker.width = m.width` and `m.picker.height = m.height` after creation (belt-and-suspenders, since the constructor now handles it)

  Also fix the same pattern for WeeklyBrowserModel for consistency:
  1. In `weekly_browser.go`, update `NewWeeklyBrowser` to accept width, height: `func NewWeeklyBrowser(t theme.Theme, width, height int) WeeklyBrowserModel`
  2. Set width/height in the returned struct
  3. In `model.go` WeeklyBrowser handler (around line 349), update the call to: `NewWeeklyBrowser(m.theme, m.width, m.height)`

  Search for any other callers of `NewThemePicker` or `NewWeeklyBrowser` (e.g., in main.go for first-run picker) and update those call sites too. The first-run picker in main.go likely has the same bug but may not manifest because bubbletea sends an initial WindowSizeMsg before the first render in standalone mode.
  </action>
  <verify>
  1. `cd /Users/BobbySmith/source/github/b00y0h/wakafetch-brew/wakadash && go build ./...` compiles without errors
  2. `go vet ./...` passes
  3. `go test ./internal/tui/...` passes (if tests exist)
  4. Run the app (`go run .`), press 't' -- theme picker should appear with preview, NOT "terminal too small"
  5. Press Esc to return to dashboard, confirm dashboard still works
  </verify>
  <done>Pressing 't' opens theme picker with full preview at any reasonable terminal size. The "terminal too small" message only appears if the terminal is genuinely under 40x10.</done>
</task>

</tasks>

<verification>
- `go build ./...` succeeds
- `go vet ./...` succeeds
- Manual test: launch app, press 't', see theme picker with preview (not size error)
- Manual test: press Esc from picker, return to dashboard, all panels functional
- Manual test: press 'w', weekly browser opens correctly (consistency fix)
</verification>

<success_criteria>
Theme picker opens correctly when 't' is pressed, showing the theme preview and navigation controls. The "terminal too small" error only appears when the terminal is genuinely smaller than 40 columns or 10 rows.
</success_criteria>

<output>
After completion, create `.planning/quick/4-fix-terminal-too-small-error-when-select/4-SUMMARY.md`
</output>
