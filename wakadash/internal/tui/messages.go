// Package tui implements the bubbletea TUI model for the wakadash dashboard.
package tui

import (
	"time"

	"github.com/b00y0h/wakadash/internal/types"
)

// statsFetchedMsg is sent by fetchStatsCmd when the API call succeeds.
type statsFetchedMsg struct {
	stats *types.StatsResponse
}

// fetchErrMsg is sent by fetchStatsCmd when the API call fails.
type fetchErrMsg struct {
	err error
}

// refreshMsg is sent when the refresh interval elapses - triggers new stats fetch.
type refreshMsg time.Time

// countdownTickMsg is sent every second for countdown display updates.
type countdownTickMsg time.Time

// durationsFetchedMsg is sent when hourly durations are fetched for sparkline.
type durationsFetchedMsg struct {
	durations *types.DurationsResponse
}

// summaryFetchedMsg is sent when daily summaries are fetched for heatmap.
type summaryFetchedMsg struct {
	summary *types.SummaryResponse
}

// archiveFetchedMsg is sent when archive data fetch completes.
type archiveFetchedMsg struct {
	data *types.DayData // nil if archive not found (404)
	date string         // Date that was fetched
}

// dataFetchedMsg is sent when DataSource.Fetch completes.
// Works for both API (recent) and archive (older) dates.
type dataFetchedMsg struct {
	data *types.DayData
	date string
}
