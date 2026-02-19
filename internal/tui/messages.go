// Package tui implements the bubbletea TUI model for the wakadash dashboard.
package tui

import "github.com/b00y0h/wakadash/internal/types"

// statsFetchedMsg is sent by fetchStatsCmd when the API call succeeds.
type statsFetchedMsg struct {
	stats *types.StatsResponse
}

// fetchErrMsg is sent by fetchStatsCmd when the API call fails.
type fetchErrMsg struct {
	err error
}
