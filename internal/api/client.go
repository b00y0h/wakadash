// Package api provides a client for the WakaTime v1 REST API.
//
// Attribution: Rewritten from github.com/sahaj-b/wakafetch (MIT License)
package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/b00y0h/wakadash/internal/types"
)

const requestTimeout = 10 * time.Second

// Client holds the API credentials and base URL for WakaTime API calls.
type Client struct {
	APIKey  string // #nosec G117
	APIURL  string
	httpCli *http.Client
}

// New creates a Client with the given credentials and a 10-second timeout.
func New(apiKey, apiURL string) *Client {
	return &Client{
		APIKey:  apiKey,
		APIURL:  strings.TrimSuffix(apiURL, "/"),
		httpCli: &http.Client{Timeout: requestTimeout},
	}
}

// FetchStats retrieves aggregated statistics for the given time range.
// rangeStr is one of: last_7_days, last_30_days, last_6_months, last_year, all_time.
func (c *Client) FetchStats(rangeStr string) (*types.StatsResponse, error) {
	url := c.buildURL(fmt.Sprintf("/v1/users/current/stats/%s", rangeStr))
	result, err := fetchJSON[types.StatsResponse](c, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stats: %w", err)
	}
	return result, nil
}

// FetchSummary retrieves daily summaries for the past `days` days.
func (c *Client) FetchSummary(days int) (*types.SummaryResponse, error) {
	today := time.Now()
	todayDate := today.Format("2006-01-02")
	startDate := today.AddDate(0, 0, -days+1).Format("2006-01-02")

	url := c.buildURL(fmt.Sprintf(
		"/v1/users/current/summaries?start=%s&end=%s",
		startDate, todayDate,
	))

	result, err := fetchJSON[types.SummaryResponse](c, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch summary: %w", err)
	}
	return result, nil
}

// FetchDurations retrieves coding sessions for a specific date (YYYY-MM-DD format).
// Returns time-stamped durations that can be grouped by hour for sparkline visualization.
func (c *Client) FetchDurations(date string) (*types.DurationsResponse, error) {
	url := c.buildURL(fmt.Sprintf("/v1/users/current/durations?date=%s", date))
	result, err := fetchJSON[types.DurationsResponse](c, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch durations: %w", err)
	}
	return result, nil
}

// buildURL constructs the full request URL, handling Wakapi's /v1 path prefix.
func (c *Client) buildURL(path string) string {
	if strings.HasSuffix(c.APIURL, "/v1") {
		// Strip the leading /v1 from path since it's already in the base URL.
		path = strings.TrimPrefix(path, "/v1")
	}
	return c.APIURL + path
}

// fetchJSON makes an authenticated GET request and decodes the JSON response into T.
func fetchJSON[T any](c *Client, urlStr string) (*T, error) {
	// Validate URL to mitigate SSRF
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme: %s", parsedURL.Scheme)
	}

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	encodedKey := base64.StdEncoding.EncodeToString([]byte(c.APIKey))
	req.Header.Set("Authorization", "Basic "+encodedKey)

	// #nosec G704 - URL is validated above to have http/https scheme
	resp, err := c.httpCli.Do(req)
	if err != nil {
		if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
			return nil, fmt.Errorf("request timed out after %s", requestTimeout)
		}
		return nil, fmt.Errorf("unable to reach server: check your internet connection")
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// continue to decode
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("authentication failed (401): check your API key")
	case http.StatusForbidden:
		return nil, fmt.Errorf("access forbidden (403): your API key may lack permissions")
	case http.StatusNotFound:
		return nil, fmt.Errorf("endpoint not found (404): verify the API URL")
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("rate limit exceeded (429): try again later")
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return nil, fmt.Errorf("server unavailable (%s): try again later", resp.Status)
	default:
		return nil, fmt.Errorf("API request failed: %s", resp.Status)
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("invalid JSON response from server: %w", err)
	}

	return &result, nil
}
