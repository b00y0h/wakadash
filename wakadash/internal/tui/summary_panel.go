package tui

import (
	"fmt"
	"strings"

	"github.com/b00y0h/wakadash/internal/types"
)

// calculateStreaks calculates current and best streaks from 7-day window.
// Returns (current, best) where current is consecutive days from most recent
// and best is the longest consecutive run in the 7-day data.
func calculateStreaks(summaryData *types.SummaryResponse) (current int, best int) {
	if summaryData == nil || len(summaryData.Data) == 0 {
		return 0, 0
	}

	// Calculate current streak: count consecutive days with activity from most recent backwards
	current = 0
	for i := len(summaryData.Data) - 1; i >= 0; i-- {
		if summaryData.Data[i].GrandTotal.TotalSeconds > 0 {
			current++
		} else {
			break
		}
	}

	// Calculate best streak: longest consecutive run in available data
	best = 0
	currentRun := 0
	for _, day := range summaryData.Data {
		if day.GrandTotal.TotalSeconds > 0 {
			currentRun++
			if currentRun > best {
				best = currentRun
			}
		} else {
			currentRun = 0
		}
	}

	return current, best
}

// renderSummaryPanel renders the summary panel with 30-day overview statistics.
func (m Model) renderSummaryPanel() string {
	statsData := m.getActiveStatsData()
	if statsData == nil {
		return ""
	}

	var sb strings.Builder
	data := *statsData

	// Total and averages
	sb.WriteString(fmt.Sprintf("Total: %s\n", data.HumanReadableTotal))
	sb.WriteString(fmt.Sprintf("Daily average: %s\n", data.HumanReadableDailyAverage))

	// Best day (if available)
	if data.BestDay.Date != "" {
		sb.WriteString(fmt.Sprintf("Best day: %s (%s)\n", data.BestDay.Date, data.BestDay.Text))
	}

	// Streak information
	currentStreak, bestStreak := calculateStreaks(m.summaryData)
	sb.WriteString(fmt.Sprintf("Streak: Current: %d days | Best: %d days\n", currentStreak, bestStreak))

	sb.WriteString("\n")

	// Top items
	if len(data.Languages) > 0 {
		sb.WriteString(fmt.Sprintf("Top language: %s\n", data.Languages[0].Name))
	}
	if len(data.Projects) > 0 {
		sb.WriteString(fmt.Sprintf("Top project: %s\n", data.Projects[0].Name))
	}
	if len(data.Editors) > 0 {
		sb.WriteString(fmt.Sprintf("Top editor: %s\n", data.Editors[0].Name))
	}
	if len(data.Categories) > 0 {
		sb.WriteString(fmt.Sprintf("Top category: %s\n", data.Categories[0].Name))
	}

	sb.WriteString("\n")

	// Counts
	sb.WriteString(fmt.Sprintf("Languages: %d\n", len(data.Languages)))
	sb.WriteString(fmt.Sprintf("Projects: %d", len(data.Projects)))

	content := sb.String()
	return renderBorderedPanel("Summary", content, m.width-4, m.theme)
}
