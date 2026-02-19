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
