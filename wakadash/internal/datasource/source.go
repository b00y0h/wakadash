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
func (ds *DataSource) Fetch(date string) (*types.DayData, error) {
	return nil, nil // TODO: implement
}

// extractDay finds the matching date in a SummaryResponse.
// Returns nil if no matching date is found.
func (ds *DataSource) extractDay(summary *types.SummaryResponse, date string) *types.DayData {
	return nil // TODO: implement
}
