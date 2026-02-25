package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/b00y0h/wakadash/internal/api"
	"github.com/b00y0h/wakadash/internal/archive"
	"github.com/b00y0h/wakadash/internal/datasource"
	"github.com/b00y0h/wakadash/internal/types"
)

// isRetryableError returns true for transient errors that should be retried.
func isRetryableError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "504") ||
		strings.Contains(errStr, "timed out")
}

// fetchWithRetry wraps an operation with exponential backoff for transient errors.
func fetchWithRetry[T any](operation func() (*T, error)) (*T, error) {
	op := backoff.Operation[*T](func() (*T, error) {
		res, err := operation()
		if err != nil {
			if isRetryableError(err) {
				return nil, err // Retry
			}
			// Permanent error
			return nil, backoff.Permanent(err)
		}
		return res, nil
	})

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 1 * time.Second
	b.MaxInterval = 30 * time.Second
	b.Multiplier = 2.0
	b.RandomizationFactor = 0.5

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := backoff.Retry(ctx, op, backoff.WithBackOff(b))
	return result, err
}

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

		stats, err := fetchWithRetry(func() (*types.StatsResponse, error) {
			return client.FetchStats(rangeStr)
		})
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

// fetchDurationsCmd fetches today's durations for sparkline visualization.
func fetchDurationsCmd(client *api.Client) tea.Cmd {
	return func() (msg tea.Msg) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch v := r.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("panic in fetchDurationsCmd: %v", r)
				}
				msg = fetchErrMsg{err: err}
			}
		}()

		today := time.Now().Format("2006-01-02")
		durations, err := fetchWithRetry(func() (*types.DurationsResponse, error) {
			return client.FetchDurations(today)
		})
		if err != nil {
			return fetchErrMsg{err: err}
		}
		return durationsFetchedMsg{durations: durations}
	}
}

// fetchSummaryCmd fetches last 7 days of summaries for heatmap.
func fetchSummaryCmd(client *api.Client) tea.Cmd {
	return func() (msg tea.Msg) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch v := r.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("panic in fetchSummaryCmd: %v", r)
				}
				msg = fetchErrMsg{err: err}
			}
		}()

		summary, err := fetchWithRetry(func() (*types.SummaryResponse, error) {
			return client.FetchSummary(7) // Last 7 days
		})
		if err != nil {
			return fetchErrMsg{err: err}
		}
		return summaryFetchedMsg{summary: summary}
	}
}

// fetchArchiveCmd fetches archived data for a specific date.
// Returns archiveFetchedMsg with data=nil if archive not found (graceful).
func fetchArchiveCmd(fetcher *archive.Fetcher, date string) tea.Cmd {
	return func() (msg tea.Msg) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch v := r.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("panic in fetchArchiveCmd: %v", r)
				}
				msg = fetchErrMsg{err: err}
			}
		}()

		// Fetcher may be nil if history_repo not configured
		data, err := fetcher.FetchArchive(date)
		if err != nil {
			return fetchErrMsg{err: err}
		}
		// data may be nil if 404 (archive not found) - that's OK
		return archiveFetchedMsg{data: data, date: date}
	}
}

// fetchDataCmd fetches day data using the hybrid DataSource.
// Routes to API for recent dates, archive for older dates.
func fetchDataCmd(ds *datasource.DataSource, date string) tea.Cmd {
	return func() (msg tea.Msg) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch v := r.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("panic in fetchDataCmd: %v", r)
				}
				msg = fetchErrMsg{err: err}
			}
		}()

		data, err := ds.Fetch(date)
		if err != nil {
			return fetchErrMsg{err: err}
		}
		return dataFetchedMsg{data: data, date: date}
	}
}
