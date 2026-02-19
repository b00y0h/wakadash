package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/b00y0h/wakadash/internal/api"
)

// fetchStatsCmd returns a tea.Cmd that fetches WakaTime stats asynchronously.
// The result flows back into Update as statsFetchedMsg or fetchErrMsg.
// A recover() guard is included per research pitfall #2: panics inside tea.Cmd
// goroutines break the terminal; explicit error returns keep it recoverable.
func fetchStatsCmd(client *api.Client, rangeStr string) tea.Cmd {
	return func() (msg tea.Msg) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch v := r.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("panic in fetchStatsCmd: %v", r)
				}
				msg = fetchErrMsg{err: err}
			}
		}()

		stats, err := client.FetchStats(rangeStr)
		if err != nil {
			return fetchErrMsg{err: err}
		}
		return statsFetchedMsg{stats: stats}
	}
}

// scheduleRefresh returns a Cmd that fires refreshMsg after interval.
// CRITICAL: Only call from statsFetchedMsg/fetchErrMsg handler to avoid double tickers (pitfall #3).
// The self-loop happens when refreshMsg triggers a new fetch, and statsFetchedMsg schedules the next refresh.
func scheduleRefresh(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return refreshMsg(t)
	})
}

// tickEverySecond returns a Cmd that fires countdownTickMsg after 1 second.
// Used for countdown display; self-loops from countdownTickMsg handler.
func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return countdownTickMsg(t)
	})
}
