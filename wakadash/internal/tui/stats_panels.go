package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/b00y0h/wakadash/internal/types"
)

// barItem represents a single item in a bar chart.
type barItem struct {
	name    string
	seconds float64
}

// Bar character - matches wakafetch style (has natural vertical spacing)
const barChar = "━"

// renderBarChart renders a wakafetch-style horizontal bar chart.
// Format: name ━━━━━━━━━━━━░░░░░ time
// Uses two-tone bars: colored fill + dim background track
func renderBarChart(items []barItem, maxSeconds float64, barColor lipgloss.Color, panelWidth int) string {
	if len(items) == 0 {
		return "  No data"
	}

	// Find max name length for alignment
	maxNameLen := 0
	for _, item := range items {
		if len(item.name) > maxNameLen {
			maxNameLen = len(item.name)
		}
	}
	// Cap name length to avoid overflow, but allow more space
	if maxNameLen > 20 {
		maxNameLen = 20
	}

	// Calculate bar width (panel width - name - spacing - time)
	// Time format: "XXXh XXm" = ~9 chars, plus spacing = ~12
	barWidth := panelWidth - maxNameLen - 12
	if barWidth < 10 {
		barWidth = 10
	}
	// Allow bars to stretch wider for larger terminals
	if barWidth > 80 {
		barWidth = 80
	}

	var sb strings.Builder
	barStyle := lipgloss.NewStyle().Foreground(barColor)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))

	for _, item := range items {
		// Truncate long names
		name := item.name
		if len(name) > maxNameLen {
			name = name[:maxNameLen-1] + "…"
		}

		// Calculate bar length proportional to max
		barLen := 0
		if maxSeconds > 0 {
			barLen = int(float64(barWidth) * (item.seconds / maxSeconds))
		}
		if barLen < 1 && item.seconds > 0 {
			barLen = 1 // Show at least 1 char for non-zero values
		}

		// Build two-tone bar: colored fill + dim background track
		filledBar := strings.Repeat(barChar, barLen)
		emptyBar := strings.Repeat(barChar, barWidth-barLen)

		// Format time
		timeStr := formatSecondsCompact(item.seconds)

		// Render line: name (padded) + colored bar + dim track + time
		line := fmt.Sprintf("%-*s %s%s %s\n",
			maxNameLen, name,
			barStyle.Render(filledBar),
			dimStyle.Render(emptyBar),
			timeStr,
		)
		sb.WriteString(line)
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// formatSecondsCompact formats seconds as "XXh XXm" or "XXm XXs".
func formatSecondsCompact(secs float64) string {
	total := int(secs)
	if total == 0 {
		return "    0s"
	}

	hours := total / 3600
	mins := (total % 3600) / 60
	seconds := total % 60

	if hours > 0 {
		return fmt.Sprintf("%3dh %2dm", hours, mins)
	}
	if mins > 0 {
		return fmt.Sprintf("%3dm %2ds", mins, seconds)
	}
	return fmt.Sprintf("    %2ds", seconds)
}

// getTopItems extracts top N items from a slice, with optional "Other" aggregation.
func getTopItems(items []types.StatItem, limit int) []barItem {
	result := make([]barItem, 0, limit+1)

	var otherSeconds float64
	for i, item := range items {
		if i < limit {
			result = append(result, barItem{
				name:    item.Name,
				seconds: item.TotalSeconds,
			})
		} else {
			otherSeconds += item.TotalSeconds
		}
	}

	// Add "Other" if there are more items
	if len(items) > limit && otherSeconds > 0 {
		result = append(result, barItem{
			name:    "Other",
			seconds: otherSeconds,
		})
	}

	return result
}

// getMaxSeconds finds the maximum seconds value in a slice.
func getMaxSeconds(items []barItem) float64 {
	var max float64
	for _, item := range items {
		if item.seconds > max {
			max = item.seconds
		}
	}
	return max
}

// renderLanguagesPanel renders the languages panel with wakafetch-style bars.
func (m Model) renderLanguagesPanel() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle(m.theme).Render("Languages") + "\n")

	if m.stats == nil || len(m.stats.Data.Languages) == 0 {
		sb.WriteString("  No data")
		return sb.String()
	}

	items := getTopItems(m.stats.Data.Languages, 10)
	maxSecs := getMaxSeconds(items)
	panelWidth := m.width/2 - 4
	sb.WriteString(renderBarChart(items, maxSecs, m.theme.Primary, panelWidth))
	return sb.String()
}

// renderProjectsPanel renders the projects panel with wakafetch-style bars.
func (m Model) renderProjectsPanel() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle(m.theme).Render("Projects") + "\n")

	if m.stats == nil || len(m.stats.Data.Projects) == 0 {
		sb.WriteString("  No data")
		return sb.String()
	}

	items := getTopItems(m.stats.Data.Projects, 10)
	maxSecs := getMaxSeconds(items)
	panelWidth := m.width/2 - 4
	// Use secondary/accent color for projects (cyan-ish)
	projectColor := lipgloss.Color("#00d7ff")
	sb.WriteString(renderBarChart(items, maxSecs, projectColor, panelWidth))
	return sb.String()
}

// renderCategoriesPanel renders the categories panel with wakafetch-style bars.
func (m Model) renderCategoriesPanel() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle(m.theme).Render("Categories") + "\n")

	if m.stats == nil || len(m.stats.Data.Categories) == 0 {
		sb.WriteString("  No data")
		return sb.String()
	}

	items := getTopItems(m.stats.Data.Categories, 10)
	maxSecs := getMaxSeconds(items)
	panelWidth := m.width/2 - 4
	sb.WriteString(renderBarChart(items, maxSecs, m.theme.Primary, panelWidth))
	return sb.String()
}

// renderEditorsPanel renders the editors panel with wakafetch-style bars.
func (m Model) renderEditorsPanel() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle(m.theme).Render("Editors") + "\n")

	if m.stats == nil || len(m.stats.Data.Editors) == 0 {
		sb.WriteString("  No data")
		return sb.String()
	}

	items := getTopItems(m.stats.Data.Editors, 10)
	maxSecs := getMaxSeconds(items)
	panelWidth := m.width/2 - 4
	sb.WriteString(renderBarChart(items, maxSecs, m.theme.Primary, panelWidth))
	return sb.String()
}

// renderOSPanel renders the operating systems panel with wakafetch-style bars.
func (m Model) renderOSPanel() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle(m.theme).Render("Operating Systems") + "\n")

	if m.stats == nil || len(m.stats.Data.OperatingSystems) == 0 {
		sb.WriteString("  No data")
		return sb.String()
	}

	items := getTopItems(m.stats.Data.OperatingSystems, 10)
	maxSecs := getMaxSeconds(items)
	panelWidth := m.width/2 - 4
	sb.WriteString(renderBarChart(items, maxSecs, m.theme.Primary, panelWidth))
	return sb.String()
}

// renderMachinesPanel renders the machines panel with wakafetch-style bars.
func (m Model) renderMachinesPanel() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle(m.theme).Render("Machines") + "\n")

	if m.stats == nil || len(m.stats.Data.Machines) == 0 {
		sb.WriteString("  No data")
		return sb.String()
	}

	items := getTopItems(m.stats.Data.Machines, 10)
	maxSecs := getMaxSeconds(items)
	panelWidth := m.width/2 - 4
	sb.WriteString(renderBarChart(items, maxSecs, m.theme.Primary, panelWidth))
	return sb.String()
}
