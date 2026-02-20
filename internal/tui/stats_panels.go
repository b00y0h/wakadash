package tui

import (
	"fmt"
	"strings"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/charmbracelet/lipgloss"
)

// formatTimeWithPercent formats seconds and percentage as "2h 15m (65%)".
func formatTimeWithPercent(secs float64, percent float64) string {
	timeStr := formatSeconds(secs)
	return fmt.Sprintf("%s (%.0f%%)", timeStr, percent)
}

// updateCategoriesChart populates the categories chart from stats data.
func (m *Model) updateCategoriesChart() {
	if m.stats == nil {
		return
	}

	m.categoriesChart.Clear()
	data := m.stats.Data

	// Limit to top 10 categories
	limit := 10
	if len(data.Categories) < limit {
		limit = len(data.Categories)
	}

	// Calculate total for percentages
	var total float64
	for _, cat := range data.Categories {
		total += cat.TotalSeconds
	}

	// Add top 10 categories
	var otherSeconds float64
	for i, cat := range data.Categories {
		if i < limit {
			hours := cat.TotalSeconds / 3600.0
			percent := (cat.TotalSeconds / total) * 100
			label := fmt.Sprintf("%s: %s", cat.Name, formatTimeWithPercent(cat.TotalSeconds, percent))
			barStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
			m.categoriesChart.Push(barchart.BarData{
				Label: label,
				Values: []barchart.BarValue{
					{
						Name:  "",
						Value: hours,
						Style: barStyle,
					},
				},
			})
		} else {
			otherSeconds += cat.TotalSeconds
		}
	}

	// Add "Other" category if there are remaining items
	if len(data.Categories) > limit {
		hours := otherSeconds / 3600.0
		percent := (otherSeconds / total) * 100
		label := fmt.Sprintf("Other: %s", formatTimeWithPercent(otherSeconds, percent))
		barStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
		m.categoriesChart.Push(barchart.BarData{
			Label: label,
			Values: []barchart.BarValue{
				{
					Name:  "",
					Value: hours,
					Style: barStyle,
				},
			},
		})
	}

	m.categoriesChart.Draw()
}

// updateEditorsChart populates the editors chart from stats data.
func (m *Model) updateEditorsChart() {
	if m.stats == nil {
		return
	}

	m.editorsChart.Clear()
	data := m.stats.Data

	// Limit to top 10 editors
	limit := 10
	if len(data.Editors) < limit {
		limit = len(data.Editors)
	}

	// Calculate total for percentages
	var total float64
	for _, ed := range data.Editors {
		total += ed.TotalSeconds
	}

	// Add top 10 editors
	var otherSeconds float64
	for i, ed := range data.Editors {
		if i < limit {
			hours := ed.TotalSeconds / 3600.0
			percent := (ed.TotalSeconds / total) * 100
			label := fmt.Sprintf("%s: %s", ed.Name, formatTimeWithPercent(ed.TotalSeconds, percent))
			barStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
			m.editorsChart.Push(barchart.BarData{
				Label: label,
				Values: []barchart.BarValue{
					{
						Name:  "",
						Value: hours,
						Style: barStyle,
					},
				},
			})
		} else {
			otherSeconds += ed.TotalSeconds
		}
	}

	// Add "Other" category if there are remaining items
	if len(data.Editors) > limit {
		hours := otherSeconds / 3600.0
		percent := (otherSeconds / total) * 100
		label := fmt.Sprintf("Other: %s", formatTimeWithPercent(otherSeconds, percent))
		barStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
		m.editorsChart.Push(barchart.BarData{
			Label: label,
			Values: []barchart.BarValue{
				{
					Name:  "",
					Value: hours,
					Style: barStyle,
				},
			},
		})
	}

	m.editorsChart.Draw()
}

// updateOSChart populates the operating systems chart from stats data.
func (m *Model) updateOSChart() {
	if m.stats == nil {
		return
	}

	m.osChart.Clear()
	data := m.stats.Data

	// Limit to top 10 operating systems
	limit := 10
	if len(data.OperatingSystems) < limit {
		limit = len(data.OperatingSystems)
	}

	// Calculate total for percentages
	var total float64
	for _, os := range data.OperatingSystems {
		total += os.TotalSeconds
	}

	// Add top 10 operating systems
	var otherSeconds float64
	for i, os := range data.OperatingSystems {
		if i < limit {
			hours := os.TotalSeconds / 3600.0
			percent := (os.TotalSeconds / total) * 100
			label := fmt.Sprintf("%s: %s", os.Name, formatTimeWithPercent(os.TotalSeconds, percent))
			barStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
			m.osChart.Push(barchart.BarData{
				Label: label,
				Values: []barchart.BarValue{
					{
						Name:  "",
						Value: hours,
						Style: barStyle,
					},
				},
			})
		} else {
			otherSeconds += os.TotalSeconds
		}
	}

	// Add "Other" category if there are remaining items
	if len(data.OperatingSystems) > limit {
		hours := otherSeconds / 3600.0
		percent := (otherSeconds / total) * 100
		label := fmt.Sprintf("Other: %s", formatTimeWithPercent(otherSeconds, percent))
		barStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
		m.osChart.Push(barchart.BarData{
			Label: label,
			Values: []barchart.BarValue{
				{
					Name:  "",
					Value: hours,
					Style: barStyle,
				},
			},
		})
	}

	m.osChart.Draw()
}

// updateMachinesChart populates the machines chart from stats data.
func (m *Model) updateMachinesChart() {
	if m.stats == nil {
		return
	}

	m.machinesChart.Clear()
	data := m.stats.Data

	// Limit to top 10 machines
	limit := 10
	if len(data.Machines) < limit {
		limit = len(data.Machines)
	}

	// Calculate total for percentages
	var total float64
	for _, mach := range data.Machines {
		total += mach.TotalSeconds
	}

	// Add top 10 machines
	var otherSeconds float64
	for i, mach := range data.Machines {
		if i < limit {
			hours := mach.TotalSeconds / 3600.0
			percent := (mach.TotalSeconds / total) * 100
			label := fmt.Sprintf("%s: %s", mach.Name, formatTimeWithPercent(mach.TotalSeconds, percent))
			barStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
			m.machinesChart.Push(barchart.BarData{
				Label: label,
				Values: []barchart.BarValue{
					{
						Name:  "",
						Value: hours,
						Style: barStyle,
					},
				},
			})
		} else {
			otherSeconds += mach.TotalSeconds
		}
	}

	// Add "Other" category if there are remaining items
	if len(data.Machines) > limit {
		hours := otherSeconds / 3600.0
		percent := (otherSeconds / total) * 100
		label := fmt.Sprintf("Other: %s", formatTimeWithPercent(otherSeconds, percent))
		barStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
		m.machinesChart.Push(barchart.BarData{
			Label: label,
			Values: []barchart.BarValue{
				{
					Name:  "",
					Value: hours,
					Style: barStyle,
				},
			},
		})
	}

	m.machinesChart.Draw()
}

// renderCategoriesPanel renders the categories panel with title.
func (m Model) renderCategoriesPanel() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle(m.theme).Render("Categories") + "\n")
	if m.stats == nil || len(m.stats.Data.Categories) == 0 {
		sb.WriteString("  No data")
	} else {
		sb.WriteString(m.categoriesChart.View())
	}
	return sb.String()
}

// renderEditorsPanel renders the editors panel with title.
func (m Model) renderEditorsPanel() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle(m.theme).Render("Editors") + "\n")
	if m.stats == nil || len(m.stats.Data.Editors) == 0 {
		sb.WriteString("  No data")
	} else {
		sb.WriteString(m.editorsChart.View())
	}
	return sb.String()
}

// renderOSPanel renders the operating systems panel with title.
func (m Model) renderOSPanel() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle(m.theme).Render("Operating Systems") + "\n")
	if m.stats == nil || len(m.stats.Data.OperatingSystems) == 0 {
		sb.WriteString("  No data")
	} else {
		sb.WriteString(m.osChart.View())
	}
	return sb.String()
}

// renderMachinesPanel renders the machines panel with title.
func (m Model) renderMachinesPanel() string {
	var sb strings.Builder
	sb.WriteString(TitleStyle(m.theme).Render("Machines") + "\n")
	if m.stats == nil || len(m.stats.Data.Machines) == 0 {
		sb.WriteString("  No data")
	} else {
		sb.WriteString(m.machinesChart.View())
	}
	return sb.String()
}
