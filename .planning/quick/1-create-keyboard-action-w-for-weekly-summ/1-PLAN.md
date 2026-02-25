---
phase: quick-weekly-browser
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/weekly_browser.go
  - internal/tui/keymap.go
  - internal/tui/model.go
  - internal/tui/messages.go
  - internal/tui/commands.go
autonomous: true

must_haves:
  truths:
    - "User presses 'w' and sees a scrollable list of available weeks (up to 52)"
    - "Each week in the list shows date range, total coding time, and top language"
    - "User navigates weeks with up/down arrow keys"
    - "User presses Enter on a week to see that week's detailed dashboard view"
    - "User presses Esc to return to the dashboard without changing the viewed week"
    - "Loading spinner shown while weekly data is being fetched from GitHub archive"
  artifacts:
    - path: "internal/tui/weekly_browser.go"
      provides: "WeeklyBrowserModel with list rendering, navigation, and selection"
      min_lines: 150
    - path: "internal/tui/keymap.go"
      provides: "WeeklyBrowser key binding for 'w'"
      contains: "WeeklyBrowser"
    - path: "internal/tui/messages.go"
      provides: "weeklyDataFetchedMsg for async week scanning results"
      contains: "weeklyDataFetchedMsg"
    - path: "internal/tui/commands.go"
      provides: "fetchWeeklySummariesCmd that scans up to 52 weeks"
      contains: "fetchWeeklySummariesCmd"
  key_links:
    - from: "internal/tui/model.go"
      to: "internal/tui/weekly_browser.go"
      via: "showWeeklyBrowser flag and WeeklyBrowserModel field"
      pattern: "showWeeklyBrowser"
    - from: "internal/tui/weekly_browser.go"
      to: "internal/tui/commands.go"
      via: "fetchWeeklySummariesCmd called on browser open"
      pattern: "fetchWeeklySummariesCmd"
    - from: "internal/tui/model.go"
      to: "selectedWeekStart"
      via: "Week selection sets selectedWeekStart and fetches that week's data"
      pattern: "selectedWeekStart.*SelectedWeek"
---

<objective>
Add a 'w' keyboard shortcut that opens a weekly summary browser overlay. The browser
shows up to 52 weeks of historical data fetched from the GitHub archive, displaying
each week's date range, total coding time, and top language. Users navigate with
arrow keys and press Enter to jump to a specific week's detailed dashboard view.

Purpose: Give users a bird's-eye view of their coding history and quick navigation
to any historical week, rather than having to arrow-key one week at a time.

Output: New weekly_browser.go file, updated keymap/model/messages/commands files.
</objective>

<context>
@internal/tui/model.go
@internal/tui/picker.go
@internal/tui/keymap.go
@internal/tui/messages.go
@internal/tui/commands.go
@internal/tui/styles.go
@internal/tui/stats_panels.go
@internal/datasource/source.go
@internal/archive/fetcher.go
@internal/types/types.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add weekly data fetching infrastructure and WeeklyBrowserModel</name>
  <files>
    internal/tui/messages.go
    internal/tui/commands.go
    internal/tui/weekly_browser.go
  </files>
  <action>
**messages.go** - Add two new message types:

```go
// WeeklySummary holds summary data for a single week (for the weekly browser list).
type WeeklySummary struct {
    WeekStart   string  // YYYY-MM-DD (Sunday)
    WeekEnd     string  // YYYY-MM-DD (Saturday)
    TotalSeconds float64
    TopLanguage string
    ProjectCount int
    HasData     bool
}

// weeklyDataFetchedMsg is sent when all weekly summaries have been scanned.
type weeklyDataFetchedMsg struct {
    weeks []WeeklySummary
    err   error
}
```

**commands.go** - Add `fetchWeeklySummariesCmd`:

```go
func fetchWeeklySummariesCmd(ds *datasource.DataSource, maxWeeks int) tea.Cmd
```

This command scans backwards from the current week up to `maxWeeks` (52). For each week:
1. Calculate the Sunday start date for that week using `getWeekStart(time.Now())` then subtract 7 days per iteration
2. Call `ds.Fetch(weekStartDate)` to get data for that week's Sunday
3. If data exists and has `GrandTotal.TotalSeconds > 0`, create a `WeeklySummary` with:
   - `WeekStart` / `WeekEnd` (Sunday to Saturday)
   - `TotalSeconds` from `GrandTotal.TotalSeconds`
   - `TopLanguage` from `Languages[0].Name` (if any)
   - `ProjectCount` from `len(Projects)`
   - `HasData = true`
4. Include the current week (week 0) using live data awareness — for the current week, always include it with HasData=true (recent data from API)
5. Skip weeks with no data (don't include them in the list)
6. Return `weeklyDataFetchedMsg{weeks: collectedWeeks}` or `weeklyDataFetchedMsg{err: err}` on failure

Include the standard `defer recover()` panic guard matching the pattern in existing commands.

**weekly_browser.go** - Create `WeeklyBrowserModel` following the `ThemePickerModel` pattern:

```go
type WeeklyBrowserModel struct {
    weeks       []WeeklySummary // Available weeks with data
    selectedIdx int             // Currently highlighted week
    scrollOffset int            // For scrolling when list exceeds viewport
    width       int
    height      int
    loading     bool            // True while fetching weekly data
    confirmed   bool            // True when user pressed Enter
    cancelled   bool            // True when user pressed Esc
    selectedWeek string         // Week start date when confirmed
    err         error           // Fetch error
}
```

Constructor: `NewWeeklyBrowser() WeeklyBrowserModel` — returns model with `loading: true`.

**Update method** handles:
- `tea.WindowSizeMsg` — update width/height
- `weeklyDataFetchedMsg` — store weeks, set loading=false, handle err
- `tea.KeyMsg`:
  - `up/k` — move selection up (with wrapping)
  - `down/j` — move selection down (with wrapping)
  - `enter` — set confirmed=true, store selectedWeek from weeks[selectedIdx].WeekStart
  - `esc/q` — set cancelled=true
  - `home/g` — jump to top of list
  - `end/G` — jump to bottom of list

**View method** renders:
- If loading: show spinner text "Scanning weekly history..." (simple text, no spinner component needed — just use a static message since this is a sub-model)
- If error: show error message with hint to press Esc
- Otherwise: render the week list as a scrollable table/list:

```
Weekly History (23 weeks found)
Arrow keys to browse, Enter to select, Esc to cancel

  > Feb 23 - Mar 1    12h 35m   Go           3 projects
    Feb 16 - Feb 22    8h 12m   TypeScript   2 projects
    Feb 9 - Feb 15    15h 44m   Go           4 projects
    ...
```

Each row format: `[cursor] [date range]  [total time]  [top language]  [N projects]`
- Use `>` marker for selected row
- Selected row styled with theme.Primary foreground
- Non-selected rows use theme.Foreground
- Dim style for the "projects" count
- Date range formatted as "Mon DD - Mon DD" using `formatWeekRange` from model.go (already exists)
- Time formatted using `formatSeconds` (already exists in model.go) or types.formatDuration pattern
- Handle scrolling: show `maxVisible` rows (height - 6 for header/footer), adjust scrollOffset when selection moves beyond visible range

Add `IsConfirmed() bool`, `IsCancelled() bool`, `SelectedWeek() string` accessor methods (matching picker pattern).
  </action>
  <verify>
Run `cd /Users/BobbySmith/source/github/b00y0h/wakafetch-brew/wakadash && go build ./...` — must compile without errors.
  </verify>
  <done>
messages.go has WeeklySummary type and weeklyDataFetchedMsg. commands.go has fetchWeeklySummariesCmd that scans up to 52 weeks. weekly_browser.go has complete WeeklyBrowserModel with Update/View/accessors. All files compile.
  </done>
</task>

<task type="auto">
  <name>Task 2: Wire weekly browser into main Model and add 'w' keybinding</name>
  <files>
    internal/tui/keymap.go
    internal/tui/model.go
  </files>
  <action>
**keymap.go** - Add the WeeklyBrowser binding:

1. Add field to `keymap` struct: `WeeklyBrowser key.Binding`
2. Add to `FullHelp()` — put it in the navigation row alongside PrevDay/NextDay/Today: `{k.PrevDay, k.NextDay, k.Today, k.WeeklyBrowser}`
3. Add to `defaultKeymap`:
```go
WeeklyBrowser: key.NewBinding(
    key.WithKeys("w"),
    key.WithHelp("w", "weekly browser"),
),
```

**model.go** - Wire the browser into the main Model:

1. Add fields to `Model` struct:
```go
// Weekly browser
showWeeklyBrowser bool
weeklyBrowser     WeeklyBrowserModel
```

2. In `View()` method, add a check AFTER the picker check and BEFORE end-of-history check:
```go
if m.showWeeklyBrowser {
    return m.weeklyBrowser.View()
}
```

3. In `Update()` method, add a delegation block AFTER the picker delegation block (before the main `switch msg` on line ~207). Follow the exact same pattern as the picker delegation:

```go
if m.showWeeklyBrowser {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.weeklyBrowser.width = msg.Width
        m.weeklyBrowser.height = msg.Height
        m.width = msg.Width
        m.height = msg.Height
    case weeklyDataFetchedMsg:
        // Forward data to browser
        newBrowser, _ := m.weeklyBrowser.Update(msg)
        m.weeklyBrowser = newBrowser.(WeeklyBrowserModel)
    case tea.KeyMsg:
        newBrowser, _ := m.weeklyBrowser.Update(msg)
        m.weeklyBrowser = newBrowser.(WeeklyBrowserModel)
        if m.weeklyBrowser.IsConfirmed() {
            // User selected a week — navigate to it
            selectedWeek := m.weeklyBrowser.SelectedWeek()
            m.showWeeklyBrowser = false
            // Check if selected week is current week
            currentWeekStart := getWeekStart(time.Now()).Format("2006-01-02")
            if selectedWeek == currentWeekStart {
                m.selectedWeekStart = ""
                m.atOldestData = false
                return m, fetchDataCmd(m.dataSource, time.Now().Format("2006-01-02"))
            }
            // Navigate to historical week
            m.selectedWeekStart = selectedWeek
            m.atOldestData = !m.dataSource.HasOlderData(selectedWeek)
            m.showEndOfHistory = false
            return m, fetchDataCmd(m.dataSource, selectedWeek)
        }
        if m.weeklyBrowser.IsCancelled() {
            m.showWeeklyBrowser = false
            return m, nil
        }
    }
    return m, nil
}
```

4. In the `tea.KeyMsg` switch block (inside the main Update), add a case for the new keybinding after the `ChangeTheme` case:
```go
case key.Matches(msg, m.keys.WeeklyBrowser):
    // Only open if dataSource is available (needs archive access)
    if m.dataSource != nil {
        m.showWeeklyBrowser = true
        m.weeklyBrowser = NewWeeklyBrowser()
        return m, fetchWeeklySummariesCmd(m.dataSource, 52)
    }
    return m, nil
```

5. Update the status bar hint in `renderStatusBar()` to include 'w':
Change `"? help  1-9 panels  a/h all  r refresh  q quit"` to `"? help  w weeks  1-9 panels  a/h all  r refresh  q quit"`
  </action>
  <verify>
Run `cd /Users/BobbySmith/source/github/b00y0h/wakafetch-brew/wakadash && go build ./...` — must compile without errors. Run `go vet ./...` — no issues. Run existing tests: `go test ./...` — all pass.
  </verify>
  <done>
Pressing 'w' opens the weekly browser overlay. Browser shows loading state, then list of available weeks. Up/down navigates, Enter selects a week and navigates the main dashboard to it. Esc returns to dashboard. Help text updated. Status bar hint includes 'w weeks'. All tests pass.
  </done>
</task>

</tasks>

<verification>
1. `go build ./...` compiles successfully
2. `go vet ./...` reports no issues
3. `go test ./...` all existing tests pass
4. The 'w' key does not conflict with any existing keybinding (verified: existing keys are q, ?, r, t, 1-9, a, h, left, right, 0, home, ctrl+c)
5. WeeklyBrowserModel follows the same sub-model delegation pattern as ThemePickerModel
</verification>

<success_criteria>
- 'w' opens weekly browser overlay showing up to 52 weeks of history
- Each week row shows: date range, total time, top language, project count
- Arrow keys navigate, Enter selects, Esc cancels
- Selecting a week navigates the main dashboard to that week (reusing existing selectedWeekStart mechanism)
- Loading state shown during async data scan
- Graceful handling when no archive data is available
- All existing tests continue to pass
</success_criteria>

<output>
After completion, create `.planning/quick/1-create-keyboard-action-w-for-weekly-summ/1-SUMMARY.md`
</output>
