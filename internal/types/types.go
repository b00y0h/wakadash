// Package types defines the data structures for WakaTime API responses.
//
// Attribution: Rewritten from github.com/sahaj-b/wakafetch (MIT License)
package types

import "fmt"

// StatItem represents a single coding activity entry (language, project, editor, etc.).
type StatItem struct {
	Name         string  `json:"name"`
	TotalSeconds float64 `json:"total_seconds"`
	Percent      float64 `json:"percent"` // Percentage of total time
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

// BestDay represents the day with the most coding activity.
type BestDay struct {
	Date         string  `json:"date"`
	Text         string  `json:"text"`
	TotalSeconds float64 `json:"total_seconds"`
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
	BestDay                   BestDay    `json:"best_day"`
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

// AggregateFromSummary creates StatsData by aggregating daily summaries.
// This is more reliable than the /stats endpoint which can return incomplete data.
func AggregateFromSummary(summary *SummaryResponse) *StatsData {
	if summary == nil || len(summary.Data) == 0 {
		return nil
	}

	// Maps to aggregate by name
	languages := make(map[string]float64)
	projects := make(map[string]float64)
	editors := make(map[string]float64)
	categories := make(map[string]float64)
	operatingSystems := make(map[string]float64)
	machines := make(map[string]float64)

	var totalSeconds float64
	var bestDay BestDay

	// Aggregate across all days
	for _, day := range summary.Data {
		dayTotal := day.GrandTotal.TotalSeconds
		totalSeconds += dayTotal

		// Track best day
		if dayTotal > bestDay.TotalSeconds {
			bestDay.Date = day.Range.Date
			bestDay.TotalSeconds = dayTotal
			bestDay.Text = day.GrandTotal.Text
		}

		// Aggregate each dimension
		for _, item := range day.Languages {
			languages[item.Name] += item.TotalSeconds
		}
		for _, item := range day.Projects {
			projects[item.Name] += item.TotalSeconds
		}
		for _, item := range day.Editors {
			editors[item.Name] += item.TotalSeconds
		}
		for _, item := range day.Categories {
			categories[item.Name] += item.TotalSeconds
		}
		for _, item := range day.OperatingSystems {
			operatingSystems[item.Name] += item.TotalSeconds
		}
		for _, item := range day.Machines {
			machines[item.Name] += item.TotalSeconds
		}
	}

	// Convert maps to sorted slices
	stats := &StatsData{
		Languages:        mapToStatItems(languages, totalSeconds),
		Projects:         mapToStatItems(projects, totalSeconds),
		Editors:          mapToStatItems(editors, totalSeconds),
		Categories:       mapToStatItems(categories, totalSeconds),
		OperatingSystems: mapToStatItems(operatingSystems, totalSeconds),
		Machines:         mapToStatItems(machines, totalSeconds),
		TotalSeconds:     totalSeconds,
		BestDay:          bestDay,
		DailyAverage:     summary.DailyAverage.Seconds,
		Start:            summary.Start,
		End:              summary.End,
		Status:           "ok",
	}

	// Format human-readable strings
	stats.HumanReadableTotal = formatDuration(totalSeconds)
	stats.HumanReadableDailyAverage = formatDuration(summary.DailyAverage.Seconds)

	return stats
}

// MergeDayData aggregates multiple daily DayData into a single combined DayData.
// Used to combine all days of a week into one view.
func MergeDayData(days []DayData) *DayData {
	if len(days) == 0 {
		return nil
	}
	if len(days) == 1 {
		return &days[0]
	}

	languages := make(map[string]float64)
	projects := make(map[string]float64)
	editors := make(map[string]float64)
	categories := make(map[string]float64)
	operatingSystems := make(map[string]float64)
	machines := make(map[string]float64)

	var totalSeconds float64

	for _, day := range days {
		totalSeconds += day.GrandTotal.TotalSeconds
		for _, item := range day.Languages {
			languages[item.Name] += item.TotalSeconds
		}
		for _, item := range day.Projects {
			projects[item.Name] += item.TotalSeconds
		}
		for _, item := range day.Editors {
			editors[item.Name] += item.TotalSeconds
		}
		for _, item := range day.Categories {
			categories[item.Name] += item.TotalSeconds
		}
		for _, item := range day.OperatingSystems {
			operatingSystems[item.Name] += item.TotalSeconds
		}
		for _, item := range day.Machines {
			machines[item.Name] += item.TotalSeconds
		}
	}

	return &DayData{
		Languages:        mapToStatItems(languages, totalSeconds),
		Projects:         mapToStatItems(projects, totalSeconds),
		Editors:          mapToStatItems(editors, totalSeconds),
		Categories:       mapToStatItems(categories, totalSeconds),
		OperatingSystems: mapToStatItems(operatingSystems, totalSeconds),
		Machines:         mapToStatItems(machines, totalSeconds),
		GrandTotal: GrandTotal{
			TotalSeconds: totalSeconds,
			Text:         formatDuration(totalSeconds),
			Hours:        int(totalSeconds) / 3600,
			Minutes:      (int(totalSeconds) % 3600) / 60,
		},
		Range: days[0].Range,
	}
}

// mapToStatItems converts a name->seconds map to a sorted slice of StatItems.
func mapToStatItems(m map[string]float64, total float64) []StatItem {
	items := make([]StatItem, 0, len(m))
	for name, secs := range m {
		percent := 0.0
		if total > 0 {
			percent = (secs / total) * 100
		}
		items = append(items, StatItem{
			Name:         name,
			TotalSeconds: secs,
			Percent:      percent,
		})
	}

	// Sort by TotalSeconds descending
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].TotalSeconds > items[i].TotalSeconds {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	return items
}

// formatDuration converts seconds to human-readable format like "2 hrs 15 mins".
func formatDuration(secs float64) string {
	total := int(secs)
	if total == 0 {
		return "0 secs"
	}

	hours := total / 3600
	mins := (total % 3600) / 60

	if hours > 0 && mins > 0 {
		return formatPlural(hours, "hr") + " " + formatPlural(mins, "min")
	} else if hours > 0 {
		return formatPlural(hours, "hr")
	} else if mins > 0 {
		return formatPlural(mins, "min")
	}
	return formatPlural(total, "sec")
}

// formatPlural returns "N unit" or "N units" based on count.
func formatPlural(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", n, unit)
}
