package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// calculateItemsPerPanel determines how many items to show per panel based on available height.
// Returns (itemCount, wasTruncated) where itemCount is minimum 3 and maximum 10.
func calculateItemsPerPanel(availableHeight, visiblePanelCount int) (itemCount int, wasTruncated bool) {
	if visiblePanelCount == 0 {
		return 10, false
	}

	// Estimate rows per panel (title + items + padding)
	estimatedRowHeight := 3
	calculated := availableHeight / (visiblePanelCount * estimatedRowHeight)

	// Enforce minimum of 3 items per panel when truncating
	if calculated < 3 {
		calculated = 3
	}

	// Cap at maximum 10 items
	itemCount = calculated
	if itemCount > 10 {
		itemCount = 10
	}

	// Mark as truncated if we couldn't show full 10 items
	wasTruncated = (calculated < 10)

	return itemCount, wasTruncated
}

// renderStatsGrid renders the 4-panel stats grid (Categories, Editors, OS, Machines)
// with responsive layout based on terminal width.
func (m Model) renderStatsGrid() string {
	// Build array of visible panels in order
	var visiblePanels []string

	if m.showCategories {
		visiblePanels = append(visiblePanels, m.renderCategoriesPanel())
	}
	if m.showEditors {
		visiblePanels = append(visiblePanels, m.renderEditorsPanel())
	}
	if m.showOS {
		visiblePanels = append(visiblePanels, m.renderOSPanel())
	}
	if m.showMachines {
		visiblePanels = append(visiblePanels, m.renderMachinesPanel())
	}

	// Handle empty case
	if len(visiblePanels) == 0 {
		return ""
	}

	// Very small terminals - show friendly message
	if m.width < 40 {
		return DimStyle(m.theme).Render("Terminal too narrow")
	}

	// Wide terminals (>=80 cols): 2-column layout
	if m.width >= 80 {
		var rows []string
		for i := 0; i < len(visiblePanels); i += 2 {
			if i+1 < len(visiblePanels) {
				// Two panels in this row
				row := lipgloss.JoinHorizontal(
					lipgloss.Top,
					visiblePanels[i],
					strings.Repeat(" ", 2), // Gap between columns
					visiblePanels[i+1],
				)
				rows = append(rows, row)
			} else {
				// Single panel in this row
				rows = append(rows, visiblePanels[i])
			}
		}
		return lipgloss.JoinVertical(lipgloss.Left, rows...)
	}

	// Narrow terminals (40-79 cols): Stack vertically
	return lipgloss.JoinVertical(lipgloss.Left, visiblePanels...)
}

// renderDashboardLayout builds the complete dashboard layout with all sections.
func (m Model) renderDashboardLayout() string {
	var sections []string

	// Summary panel at top (full width)
	if m.showSummary {
		sections = append(sections, m.renderSummaryPanel())
	}

	// Existing stats section (Languages and Projects)
	if m.showLanguages || m.showProjects {
		statsContent := m.renderStats()
		if statsContent != "" {
			sections = append(sections, statsContent)
		}
	}

	// New stats grid (Categories, Editors, OS, Machines)
	gridContent := m.renderStatsGrid()
	if gridContent != "" {
		sections = append(sections, gridContent)
	}

	// Sparkline
	if m.showSparkline {
		sections = append(sections, m.renderSparkline())
	}

	// Heatmap
	if m.showHeatmap {
		sections = append(sections, m.renderHeatmapPanel())
	}

	// Join all sections vertically
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
