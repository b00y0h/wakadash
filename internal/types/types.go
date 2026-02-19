// Package types defines the data structures for WakaTime API responses.
//
// Attribution: Rewritten from github.com/sahaj-b/wakafetch (MIT License)
package types

// StatItem represents a single coding activity entry (language, project, editor, etc.).
type StatItem struct {
	Name         string  `json:"name"`
	TotalSeconds float64 `json:"total_seconds"`
}

// GrandTotal holds the aggregated totals for a single day.
type GrandTotal struct {
	Digital      string  `json:"digital"`
	Hours        int     `json:"hours"`
	Minutes      int     `json:"minutes"`
	Text         string  `json:"text"`
	TotalSeconds float64 `json:"total_seconds"`
}

// DateRange describes the time range for a single day's data.
type DateRange struct {
	Date     string `json:"date"`
	End      string `json:"end"`
	Start    string `json:"start"`
	Text     string `json:"text"`
	Timezone string `json:"timezone"`
}

// DayData contains all activity data for a single day from the /summaries endpoint.
type DayData struct {
	Entities         []StatItem `json:"entities"`
	Branches         []StatItem `json:"branches"`
	Categories       []StatItem `json:"categories"`
	Dependencies     []StatItem `json:"dependencies"`
	Editors          []StatItem `json:"editors"`
	Languages        []StatItem `json:"languages"`
	Machines         []StatItem `json:"machines"`
	OperatingSystems []StatItem `json:"operating_systems"`
	Projects         []StatItem `json:"projects"`
	GrandTotal       GrandTotal `json:"grand_total"`
	Range            DateRange  `json:"range"`
}

// CumulativeTotal holds the aggregated total across all days in a summary range.
type CumulativeTotal struct {
	Digital string  `json:"digital"`
	Seconds float64 `json:"seconds"`
	Text    string  `json:"text"`
}

// DailyAverage holds daily average statistics for a summary range.
type DailyAverage struct {
	DaysIncludingHolidays int     `json:"days_including_holidays"`
	DaysMinusHolidays     int     `json:"days_minus_holidays"`
	Holidays              int     `json:"holidays"`
	Seconds               float64 `json:"seconds"`
	Text                  string  `json:"text"`
}

// SummaryResponse is the top-level response from the /v1/users/current/summaries endpoint.
type SummaryResponse struct {
	Data            []DayData       `json:"data"`
	CumulativeTotal CumulativeTotal `json:"cumulative_total"`
	DailyAverage    DailyAverage    `json:"daily_average"`
	End             string          `json:"end"`
	Start           string          `json:"start"`
}

// StatsData holds aggregated statistics for the /v1/users/current/stats/{range} endpoint.
type StatsData struct {
	Branches                  []StatItem `json:"branches"`
	Categories                []StatItem `json:"categories"`
	Editors                   []StatItem `json:"editors"`
	Languages                 []StatItem `json:"languages"`
	Machines                  []StatItem `json:"machines"`
	OperatingSystems          []StatItem `json:"operating_systems"`
	Projects                  []StatItem `json:"projects"`
	Range                     string     `json:"range"`
	Status                    string     `json:"status"`
	TotalSeconds              float64    `json:"total_seconds"`
	UserID                    string     `json:"user_id"`
	Username                  string     `json:"username"`
	DailyAverage              float64    `json:"daily_average"`
	DaysIncludingHolidays     int        `json:"days_including_holidays"`
	Start                     string     `json:"start"`
	End                       string     `json:"end"`
	HumanReadableDailyAverage string     `json:"human_readable_daily_average"`
	HumanReadableRange        string     `json:"human_readable_range"`
	HumanReadableTotal        string     `json:"human_readable_total"`
	IsCodingActivityVisible   bool       `json:"is_coding_activity_visible"`
	IsOtherUsageVisible       bool       `json:"is_other_usage_visible"`
}

// StatsResponse is the top-level response from the /v1/users/current/stats/{range} endpoint.
type StatsResponse struct {
	Data StatsData `json:"data"`
}

// Duration represents a single coding session from the /durations endpoint.
type Duration struct {
	Time     float64 `json:"time"`     // UNIX timestamp (start of duration)
	Duration float64 `json:"duration"` // Duration in seconds
	Project  string  `json:"project"`  // Project name
	Language string  `json:"language"` // Primary language
}

// DurationsResponse is the response from /v1/users/current/durations endpoint.
type DurationsResponse struct {
	Data     []Duration `json:"data"`
	Start    string     `json:"start"`
	End      string     `json:"end"`
	Timezone string     `json:"timezone"`
}
