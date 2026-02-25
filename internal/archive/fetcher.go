// Package archive provides functionality to fetch historical WakaTime data from GitHub.
package archive

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/b00y0h/wakadash/internal/types"
)

const requestTimeout = 10 * time.Second

// Fetcher retrieves archived WakaTime data from a GitHub repository.
type Fetcher struct {
	HistoryRepo string       // Format: "owner/repo"
	httpCli     *http.Client // HTTP client with 10s timeout
}

// New creates a Fetcher for the given GitHub repository.
// Returns nil if historyRepo is empty or has invalid format.
// Format validation: historyRepo must be "owner/repo" (exactly one slash).
func New(historyRepo string) *Fetcher {
	// Return nil for empty repo (graceful no-op)
	if historyRepo == "" {
		return nil
	}

	// Validate format: must be "owner/repo" (exactly one slash)
	if strings.Count(historyRepo, "/") != 1 {
		return nil // Invalid format, defer error to fetch time
	}

	return &Fetcher{
		HistoryRepo: historyRepo,
		httpCli:     &http.Client{Timeout: requestTimeout},
	}
}

// FetchArchive retrieves archived data for the given date (YYYY-MM-DD format).
// Returns (nil, nil) if the file doesn't exist (404) - not an error.
// Returns (*types.DayData, nil) if data exists and is valid.
// Returns (nil, error) for other failures (network, JSON parsing, etc.).
func (f *Fetcher) FetchArchive(date string) (*types.DayData, error) {
	// Graceful no-op if fetcher is nil (no history_repo configured)
	if f == nil {
		return nil, nil
	}

	// Build GitHub raw URL
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/data/%s.json",
		f.HistoryRepo, date)

	// Make GET request (no auth needed for public repos)
	resp, err := f.httpCli.Get(url)
	if err != nil {
		if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
			return nil, fmt.Errorf("request timed out after %s", requestTimeout)
		}
		return nil, fmt.Errorf("unable to reach GitHub: check your internet connection")
	}
	defer resp.Body.Close()

	// Handle 404: data doesn't exist (not an error)
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	// Handle other non-200 responses
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub returned status %d: %s", resp.StatusCode, resp.Status)
	}

	// Decode JSON response
	var data types.DayData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("invalid JSON in archive file: %w", err)
	}

	return &data, nil
}
