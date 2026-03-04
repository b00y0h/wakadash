---
phase: quick-5
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/archive/fetcher.go
  - internal/archive/fetcher_test.go
autonomous: true
must_haves:
  truths:
    - "Archive fetcher constructs URL as data/YYYY/MM/DD/summary.json (not data/YYYY-MM-DD.json)"
    - "Archive fetcher unwraps SummaryResponse wrapper to extract DayData from Data[0]"
    - "Archive fetcher returns a clear error when repo returns 404 explaining the repo may be private"
  artifacts:
    - path: "internal/archive/fetcher.go"
      provides: "Fixed archive fetcher with correct URL pattern and SummaryResponse unwrapping"
      contains: "data/%s/%s/%s/summary.json"
    - path: "internal/archive/fetcher_test.go"
      provides: "Updated tests for new URL pattern, wrapper unwrapping, and 404 private repo warning"
  key_links:
    - from: "internal/archive/fetcher.go"
      to: "internal/types/types.go"
      via: "types.SummaryResponse for JSON decoding"
      pattern: "types\\.SummaryResponse"
---

<objective>
Fix the archive fetcher to match wakasync's actual data structure.

Purpose: The archive fetcher currently uses the wrong URL pattern and decodes JSON directly
into DayData, but wakasync stores data as `data/YYYY/MM/DD/summary.json` wrapped in a
SummaryResponse envelope `{"data": [...]}`. Also, private repos return 404 (not 403) from
raw.githubusercontent.com, and users need a helpful message to diagnose this.

Output: Updated fetcher.go with correct URL, SummaryResponse unwrapping, and private-repo
warning on 404. Updated tests to match.
</objective>

<execution_context>
@/Users/BobbySmith/.claude/get-shit-done/workflows/execute-plan.md
@/Users/BobbySmith/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/archive/fetcher.go
@internal/archive/fetcher_test.go
@internal/types/types.go
@internal/datasource/source.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Fix URL pattern, unwrap SummaryResponse, and add private-repo 404 warning</name>
  <files>internal/archive/fetcher.go</files>
  <action>
Three changes to `FetchArchive` in `internal/archive/fetcher.go`:

1. **Change URL pattern** (line 53): Parse the `date` string (YYYY-MM-DD format) into year, month, day components. Build URL as:
   ```
   https://raw.githubusercontent.com/%s/main/data/%s/%s/%s/summary.json
   ```
   where the format args are: HistoryRepo, year (4-digit), month (2-digit with leading zero), day (2-digit with leading zero).
   Use `strings.Split(date, "-")` to extract parts (date is already validated as YYYY-MM-DD by callers). Add a length check on the split result (must be 3 parts) and return an error if invalid.

2. **Unwrap SummaryResponse** (lines 77-82): Instead of decoding directly into `types.DayData`, decode into `types.SummaryResponse`. Then extract `Data[0]` as the DayData. If `Data` is empty, return `(nil, nil)` (treat as no data available, same as current 404 behavior).

3. **Change 404 handling** (lines 67-69): Instead of silently returning `(nil, nil)` on 404, return a descriptive error:
   ```go
   return nil, fmt.Errorf("archive not found for %s (if using a private repo, make it public — raw.githubusercontent.com cannot access private repos)", date)
   ```
   This is important because raw.githubusercontent.com returns 404 (not 403) for private repos, making it indistinguishable from "file doesn't exist". The error message helps users diagnose the issue.

   **Important downstream impact:** `internal/datasource/source.go` calls `FetchArchive` in several places and checks `err == nil && data != nil`. Changing 404 from `(nil, nil)` to `(nil, error)` means callers will now get an error for missing dates. This is the DESIRED behavior — it surfaces the private-repo issue. However, in `FindNonEmptyWeek` and `HasOlderData`, the error will simply cause those dates to be skipped (the `err == nil` check filters them out), so behavior is functionally unchanged for valid public repos where files legitimately don't exist. For private repos, the error will now propagate up through `Fetch()` to the TUI and display the helpful message.

   Actually, on reflection: if the repo is public but a specific date simply has no data file, 404 is expected and normal. Returning an error for every missing date would be noisy and break the week-scanning logic in `FindNonEmptyWeek`.

   **Revised approach:** Keep 404 returning `(nil, nil)` for the normal case. Instead, add a **new exported function** `CheckRepoAccess(repo string) error` that does a HEAD request to `https://raw.githubusercontent.com/{repo}/main/` and returns a helpful error if it gets 404. OR, simpler: just log/return a more descriptive error message at a higher level.

   **Simplest correct approach:** Keep `(nil, nil)` on 404 (backward compatible, doesn't break scanning). But add a **comment** in the 404 handler noting that private repos also return 404. Then add a new method `(f *Fetcher) CheckAccess() error` that tests repo accessibility by requesting a known path (e.g., the repo root raw URL). If that returns 404, return an error like "history_repo may be private or not exist — raw.githubusercontent.com cannot access private repos". This can be called once at startup.

   Implementation:
   - Keep existing 404 handler returning `(nil, nil)`
   - Add comment: `// Note: private repos also return 404 from raw.githubusercontent.com`
   - Add new method:
     ```go
     // CheckAccess verifies the history repo is accessible (public and exists).
     // Returns nil if accessible, error with guidance if not.
     // Call once at startup to surface private-repo issues early.
     func (f *Fetcher) CheckAccess() error {
         if f == nil {
             return nil
         }
         url := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/", f.HistoryRepo)
         resp, err := f.httpCli.Head(url)
         if err != nil {
             return fmt.Errorf("unable to reach history repo: %w", err)
         }
         defer resp.Body.Close()
         if resp.StatusCode == http.StatusNotFound {
             return fmt.Errorf("history repo '%s' is not accessible — if this is a private repo, make it public (raw.githubusercontent.com cannot access private repos)", f.HistoryRepo)
         }
         return nil
     }
     ```
  </action>
  <verify>
    `cd /Users/BobbySmith/source/github/b00y0h/wakafetch-brew/wakadash && go build ./...` compiles without errors.
  </verify>
  <done>
    - URL pattern uses `data/YYYY/MM/DD/summary.json` format
    - JSON decodes into SummaryResponse and extracts Data[0]
    - CheckAccess method exists for startup repo validation
    - 404 in FetchArchive still returns (nil, nil) with comment about private repos
    - Code compiles
  </done>
</task>

<task type="auto">
  <name>Task 2: Update tests for new URL pattern, SummaryResponse wrapper, and CheckAccess</name>
  <files>internal/archive/fetcher_test.go</files>
  <action>
Update `internal/archive/fetcher_test.go` to match the new behavior:

1. **Update TestFetchArchive_Success**: Change the mock JSON response from raw DayData to a SummaryResponse wrapper:
   ```json
   {
     "data": [{
       "grand_total": { "total_seconds": 3600, "digital": "1:00", "text": "1 hr", "hours": 1, "minutes": 0 },
       "languages": [{ "name": "Go", "total_seconds": 3600, "percent": 100 }],
       "projects": [{ "name": "wakadash", "total_seconds": 3600, "percent": 100 }],
       "range": { "date": "2026-02-24", "text": "Mon Feb 24", "start": "", "end": "", "timezone": "" },
       "categories": [], "editors": [], "machines": [], "operating_systems": [],
       "entities": [], "branches": [], "dependencies": []
     }],
     "cumulative_total": { "digital": "1:00", "seconds": 3600, "text": "1 hr" },
     "daily_average": { "days_including_holidays": 1, "days_minus_holidays": 1, "holidays": 0, "seconds": 3600, "text": "1 hr" },
     "start": "2026-02-24", "end": "2026-02-24"
   }
   ```
   The rest of the assertions (checking GrandTotal.TotalSeconds, Languages, Range.Date) remain the same since the function still returns *types.DayData.

2. **Add TestFetchArchive_EmptyDataArray**: Test that when SummaryResponse has `"data": []`, FetchArchive returns `(nil, nil)`.

3. **Add TestFetchArchive_URLPattern**: Verify the URL constructed by FetchArchive uses the new path format. Update the `testTransport.RoundTrip` to capture and validate the request URL path contains `/data/2026/02/24/summary.json` (not `/data/2026-02-24.json`).

4. **Add TestCheckAccess_NilFetcher**: Verify `CheckAccess()` on nil fetcher returns nil.

5. **Add TestCheckAccess_Accessible**: Use mockRoundTripper with 200 status, verify returns nil.

6. **Add TestCheckAccess_NotFound**: Use mockRoundTripper with 404 status, verify returns error containing "private repo" or "not accessible".

7. **Update TestFetchArchive_404_Mocked**: Should still pass (404 returns nil, nil) — no changes needed.

8. **Fix date format validation test**: Add TestFetchArchive_InvalidDate that passes a malformed date (e.g., "bad-date") and verifies it returns an error (from the new date parsing logic).
  </action>
  <verify>
    `cd /Users/BobbySmith/source/github/b00y0h/wakafetch-brew/wakadash && go test ./internal/archive/ -v` — all tests pass.
  </verify>
  <done>
    - TestFetchArchive_Success uses SummaryResponse-wrapped JSON and still passes
    - TestFetchArchive_EmptyDataArray confirms empty data returns (nil, nil)
    - TestFetchArchive_URLPattern confirms new URL format data/YYYY/MM/DD/summary.json
    - TestCheckAccess tests cover nil, accessible, and not-found cases
    - TestFetchArchive_InvalidDate confirms bad dates return error
    - All existing tests still pass
  </done>
</task>

</tasks>

<verification>
- `go build ./...` compiles successfully
- `go test ./internal/archive/ -v` all tests pass
- `go test ./...` full test suite passes (no regressions in datasource or other packages)
- Manual inspection: URL in fetcher.go contains `data/%s/%s/%s/summary.json`
- Manual inspection: JSON decoding uses `types.SummaryResponse` not `types.DayData` directly
</verification>

<success_criteria>
- Archive fetcher constructs URLs matching wakasync's `data/YYYY/MM/DD/summary.json` structure
- Archive fetcher correctly unwraps SummaryResponse to extract DayData
- CheckAccess method provides clear guidance about private repos
- All tests pass including new tests for wrapper handling and URL format
- No regressions in dependent packages (datasource, tui)
</success_criteria>

<output>
After completion, create `.planning/quick/5-fix-archive-fetcher-to-match-wakasync-da/5-SUMMARY.md`
</output>
