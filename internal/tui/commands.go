package tui

import (
	"fmt"

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
