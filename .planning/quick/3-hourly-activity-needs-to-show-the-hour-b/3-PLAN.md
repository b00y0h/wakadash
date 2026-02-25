---
phase: quick-3
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/tui/model.go
autonomous: true
must_haves:
  truths:
    - "Hour labels are visible below the sparkline bars in the Hourly Activity panel"
    - "Labels align correctly with the corresponding sparkline columns"
    - "Labels are readable and not cluttered at various terminal widths"
  artifacts:
    - path: "internal/tui/model.go"
      provides: "Hour labels row appended below sparkline chart output"
      contains: "renderSparkline"
  key_links:
    - from: "renderSparkline()"
      to: "sparklineChart.View()"
      via: "appending hour label row to sparkline content before passing to renderBorderedPanel"
      pattern: "renderSparkline"
---

<objective>
Add hour labels below the bars in the Hourly Activity sparkline panel.

Purpose: The sparkline shows 24 bars (one per hour) but the user cannot tell which bar corresponds to which hour. Adding hour labels below the bars makes the chart readable.
Output: Modified renderSparkline() that displays hour numbers below the chart bars.
</objective>

<execution_context>
@/Users/BobbySmith/.claude/get-shit-done/workflows/execute-plan.md
@/Users/BobbySmith/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/tui/model.go (renderSparkline at line ~824, updateSparkline at ~817, sparkline resize at ~258)
@internal/tui/styles.go (renderBorderedPanel, DimStyle)
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add hour labels row below sparkline bars</name>
  <files>internal/tui/model.go</files>
  <action>
Modify the `renderSparkline()` method to append a row of hour labels below the sparkline chart output.

Key facts about the sparkline layout:
- The sparkline canvas width equals `m.width - 4` (set in WindowSizeMsg handler at line ~258)
- 24 data points are pushed (one per hour 0-23)
- In `DrawColumnsOnly()`, columns are right-aligned: they start at column `canvasWidth - 24` and fill to the right edge
- Each column is exactly 1 character wide

Implementation:

1. In `renderSparkline()`, after getting `content := m.sparklineChart.View()`, build an hour label string:

2. Calculate the starting offset: `startCol := m.sparklineChart.Width() - 24` (this matches where the sparkline library places the first bar)

3. Build the label row as a string:
   - First `startCol` characters are spaces (padding to align with bars)
   - Then 24 single characters for hours 0-23
   - For hours 0-23, show the hour number. Since each bar is only 1 char wide, use single-digit representation:
     - Show the hour number for every 3rd hour (0, 3, 6, 9, 12, 15, 18, 21) to avoid crowding
     - Use a space for the other hours
   - Format: "0" at position 0, "3" at position 3, "6" at position 6, "9" at position 9, then two-digit hours need special handling
   - Better approach: since columns are 1-char wide, label every 6th hour for readability: 0, 6, 12, 18 (using the tens digit at one position and units at next for two-digit hours, which conveniently works since 12 is at position 12 and 18 is at position 18)

   Simplest clean approach: Build a 24-char label string where every 6th position shows the hour number. For two-digit hours (12, 18), show both digits by using positions 12-13 and 18-19. This works because the intervening positions are spaces anyway.

   Actually, the cleanest approach that works at 1-char-per-column:
   ```
   // Build labels for key hours
   labels := make([]byte, 24)
   for i := range labels {
       labels[i] = ' '
   }
   // Mark key hours: 0, 3, 6, 9, 12, 15, 18, 21
   keyHours := []int{0, 3, 6, 9, 12, 15, 18, 21}
   ```

   For each key hour, write a condensed label. Since each column is 1 char, use single chars where possible. For two-digit hours, we can show just every 6 hours and use 2 chars: write the number starting at that position (it will overlap the next space). Build as a rune slice or string builder for proper handling.

   **Final approach (recommended):** Use `strings.Builder`. Create a label row that is exactly `canvasWidth` chars wide. For positions 0 through `startCol-1`, write spaces. For the 24 hour positions, write labels at every 3rd hour. Use `fmt.Sprintf("%-3d", hour)` for 3-hour intervals, which gives "0  ", "3  ", "6  ", "9  ", "12 ", "15 ", "18 ", "21 " - each taking 3 chars, which perfectly fills 24 columns (8 labels * 3 chars = 24). This provides clean, evenly-spaced labels.

   Wait - that does not account for the fact that each position maps to exactly one column. Let me reconsider.

   The 24 bars occupy 24 consecutive single-character columns. The best approach:
   - Create a `strings.Builder`
   - Add `startCol` spaces for left padding
   - Then iterate through hours 0-23. For every 3rd hour (0, 3, 6, 9, 12, 15, 18, 21), write the hour number then pad with spaces to fill the remaining chars up to the next label. Concretely:
     - For each group of 3 positions, the first position gets the hour label, subsequent positions get spaces
     - Single digit hours (0, 3, 6, 9): write digit + 2 spaces
     - Double digit hours (12, 15, 18, 21): write 2 digits + 1 space
   - This produces exactly 24 characters: 8 groups of 3

4. Style the label row with `DimStyle(m.theme)` so it appears as secondary text.

5. Combine: `content = content + "\n" + dimLabelRow`

6. Pass combined content to `renderBorderedPanel("Hourly Activity (Today)", content, m.width-4, m.theme)`

7. Also increase the sparkline height by 1 to account for the label row NOT eating into bar space. In the WindowSizeMsg handler (around line 259), change `sparklineHeight := 5` to `sparklineHeight := 4` -- actually no, the label row is appended OUTSIDE the sparkline canvas. The sparkline canvas renders bars in its own height. We append the label below. The bordered panel will naturally accommodate the extra line. Leave sparklineHeight at 5. No change needed there.

Important: The `renderBorderedPanel` function wraps content with a border and adds padding. The content we pass will have the sparkline view + newline + label row. This should render correctly within the bordered panel.
  </action>
  <verify>
Run `go build ./...` to confirm compilation succeeds. Then run the application with `go run ./cmd/wakadash/ --help` to verify it starts without errors. Visual verification: the Hourly Activity panel should show hour labels (0, 3, 6, 9, 12, 15, 18, 21) below the sparkline bars.
  </verify>
  <done>
The Hourly Activity panel displays hour labels below the sparkline bars. Labels appear at every 3rd hour (0, 3, 6, 9, 12, 15, 18, 21), are properly aligned with the corresponding bar columns, and are styled as dim/secondary text. The labels remain correctly aligned when the terminal is resized.
  </done>
</task>

</tasks>

<verification>
- `go build ./...` compiles without errors
- Running the dashboard shows hour labels below the sparkline bars in the "Hourly Activity (Today)" panel
- Labels align with the bars at various terminal widths
- Labels show key hours (0, 3, 6, 9, 12, 15, 18, 21) in dim text
</verification>

<success_criteria>
- Hour labels are visible below sparkline bars
- Labels are correctly aligned with their corresponding bars
- Labels are readable and not cluttered
- No compilation errors or visual regressions
</success_criteria>

<output>
After completion, create `.planning/quick/3-hourly-activity-needs-to-show-the-hour-b/3-SUMMARY.md`
</output>
