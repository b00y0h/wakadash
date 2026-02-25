// Package datasource provides a unified interface for fetching WakaTime data
// from either the API (recent dates) or GitHub archive (older dates).
package datasource

import (
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
func (ds *DataSource) IsRecent(date string) bool {
	return false // TODO: implement
}

// Fetch retrieves data for the given date from API (recent) or archive (old).
func (ds *DataSource) Fetch(date string) (*types.DayData, error) {
	return nil, nil // TODO: implement
}
