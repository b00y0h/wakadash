// Package datasource provides a unified interface for fetching WakaTime data
// from either the API (recent dates) or GitHub archive (older dates).
package datasource

import (
	"time"

	"github.com/b00y0h/wakadash/internal/api"
	"github.com/b00y0h/wakadash/internal/archive"
	"github.com/b00y0h/wakadash/internal/types"
)

// DataSource routes data requests to API or archive based on date.
type DataSource struct {
	api     *api.Client
	archive *archive.Fetcher
}

// New creates a DataSource with the given API client and archive fetcher.
func New(client *api.Client, fetcher *archive.Fetcher) *DataSource {
	return &DataSource{
		api:     client,
		archive: fetcher,
	}
}

// IsRecent returns true if the date is within 7 days of today.
// Date should be in YYYY-MM-DD format.
func (ds *DataSource) IsRecent(date string) bool {
	// Parse the input date (YYYY-MM-DD, treat as local date)
	targetDate, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return false // Invalid date format is not recent
	}

	// Get today's date (midnight, local timezone)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	// Calculate 7 days ago (inclusive boundary)
	sevenDaysAgo := today.AddDate(0, 0, -7)

	// Date is recent if it's >= 7 days ago and <= today
	return !targetDate.Before(sevenDaysAgo) && !targetDate.After(today)
}

// FetchResult holds the result of a Fetch call.
// DailyTotals is populated only for archive weeks (index 0=Sun, 6=Sat).
type FetchResult struct {
	Data        *types.DayData
	DailyTotals [7]float64 // Per-day total seconds (Sun-Sat), zero if not archive
}

// Fetch retrieves data for the given date from API (recent) or archive (old).
// For archive dates, fetches all 7 days of the week and aggregates them.
func (ds *DataSource) Fetch(date string) (*FetchResult, error) {
	if ds.IsRecent(date) {
		// Recent date: use API client
		summary, err := ds.api.FetchSummary(7)
		if err != nil {
			return nil, err
		}
		return &FetchResult{Data: ds.extractDay(summary, date)}, nil
	}

	// Old date: fetch the full week from archive
	return ds.fetchArchiveWeek(date)
}

// extractDay finds the matching date in a SummaryResponse.
// Returns nil if no matching date is found.
func (ds *DataSource) extractDay(summary *types.SummaryResponse, date string) *types.DayData {
	if summary == nil {
		return nil
	}

	for i := range summary.Data {
		if summary.Data[i].Range.Date == date {
			return &summary.Data[i]
		}
	}

	return nil
}

// fetchArchiveWeek fetches all 7 days of a week from the archive and merges them.
// weekStart should be a Sunday date in YYYY-MM-DD format.
// Returns merged data plus per-day totals (index 0=Sun, 6=Sat).
func (ds *DataSource) fetchArchiveWeek(weekStart string) (*FetchResult, error) {
	if ds.archive == nil {
		return &FetchResult{}, nil
	}

	start, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return nil, err
	}

	var days []types.DayData
	var dailyTotals [7]float64
	for i := 0; i < 7; i++ {
		date := start.AddDate(0, 0, i).Format("2006-01-02")
		data, err := ds.archive.FetchArchive(date)
		if err != nil {
			continue // Skip days that error
		}
		if data != nil && data.GrandTotal.TotalSeconds > 0 {
			days = append(days, *data)
			dailyTotals[i] = data.GrandTotal.TotalSeconds
		}
	}

	if len(days) == 0 {
		return &FetchResult{}, nil
	}

	return &FetchResult{
		Data:        types.MergeDayData(days),
		DailyTotals: dailyTotals,
	}, nil
}

// weekHasArchiveData checks if any day in a week has archive data.
// Short-circuits on first day found with data.
func (ds *DataSource) weekHasArchiveData(weekStart string) bool {
	if ds.archive == nil {
		return false
	}

	start, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return false
	}

	for i := 0; i < 7; i++ {
		date := start.AddDate(0, 0, i).Format("2006-01-02")
		data, err := ds.archive.FetchArchive(date)
		if err == nil && data != nil && data.GrandTotal.TotalSeconds > 0 {
			return true
		}
	}
	return false
}

// FindNonEmptyWeek searches backward from startWeek for a week with data.
// Returns the week start date (Sunday) of the first week with data, or empty string if none found.
// maxWeeksBack limits the search depth (e.g., 52 for one year).
// For recent dates (within 7 days), assumes data exists (API always has recent data).
// For archive dates, checks all 7 days of each week (not just Sunday).
func (ds *DataSource) FindNonEmptyWeek(startWeek string, direction int, maxWeeksBack int) (string, bool) {
	if direction != -1 && direction != 1 {
		direction = -1 // Default to backward search
	}

	// Parse start week
	current, err := time.Parse("2006-01-02", startWeek)
	if err != nil {
		return "", false
	}

	// Search for non-empty week
	for i := 0; i < maxWeeksBack; i++ {
		// Check if this week has data
		weekDateStr := current.Format("2006-01-02")

		// Recent dates always have data (API)
		if ds.IsRecent(weekDateStr) {
			return weekDateStr, true
		}

		// For older dates, check all 7 days of the week
		if ds.weekHasArchiveData(weekDateStr) {
			return weekDateStr, true
		}

		// Move to next week in search direction
		current = current.AddDate(0, 0, 7*direction)
	}

	// No data found within search limit
	return "", false
}

// HasOlderData checks if there is any data older than the given week.
// Used to determine if we're at the oldest available data.
func (ds *DataSource) HasOlderData(weekStart string) bool {
	parsed, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return false
	}

	// Check the previous week
	prevWeek := parsed.AddDate(0, 0, -7).Format("2006-01-02")

	// If previous week is recent, there's always data
	if ds.IsRecent(prevWeek) {
		return true
	}

	// For older dates, check all 7 days of each week
	if ds.weekHasArchiveData(prevWeek) {
		return true
	}
	// Try a few more weeks back (up to 4) to avoid false negatives
	for i := 2; i <= 4; i++ {
		checkWeek := parsed.AddDate(0, 0, -7*i).Format("2006-01-02")
		if ds.weekHasArchiveData(checkWeek) {
			return true
		}
	}

	return false
}
