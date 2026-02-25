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
	// Parse the input date
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false // Invalid date format is not recent
	}

	// Get today's date (midnight)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Calculate 7 days ago
	sevenDaysAgo := today.AddDate(0, 0, -7)

	// Date is recent if it's >= 7 days ago and <= today
	return !targetDate.Before(sevenDaysAgo) && !targetDate.After(today)
}

// Fetch retrieves data for the given date from API (recent) or archive (old).
// Returns (*types.DayData, error) for both API and archive sources.
func (ds *DataSource) Fetch(date string) (*types.DayData, error) {
	if ds.IsRecent(date) {
		// Recent date: use API client
		// API's FetchSummary returns data for a range, so we need to extract the single day
		summary, err := ds.api.FetchSummary(7) // Fetch last 7 days to ensure we have the date
		if err != nil {
			return nil, err
		}

		// Extract the specific date from the summary response
		return ds.extractDay(summary, date), nil
	}

	// Old date: use archive fetcher
	// Nil fetcher is graceful no-op (returns nil, nil)
	if ds.archive == nil {
		return nil, nil
	}

	return ds.archive.FetchArchive(date)
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
