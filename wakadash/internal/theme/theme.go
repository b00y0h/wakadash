// Package theme provides theming support for wakadash.
package theme

import (
	"log"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DefaultTheme is the default theme name.
const DefaultTheme = "dracula"

// Theme defines a color palette for the dashboard.
type Theme struct {
	// Name is the theme identifier.
	Name string

	// Base colors
	Background lipgloss.Color
	Foreground lipgloss.Color

	// UI element colors
	Border lipgloss.Color
	Title  lipgloss.Color
	Dim    lipgloss.Color

	// Status colors
	Error   lipgloss.Color
	Warning lipgloss.Color
	Success lipgloss.Color

	// Accent colors for charts and highlights
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent1   lipgloss.Color
	Accent2   lipgloss.Color
	Accent3   lipgloss.Color
	Accent4   lipgloss.Color

	// HeatmapColors defines the 5-level intensity gradient for activity heatmaps.
	// [0]=None, [1]=Low, [2]=Medium, [3]=High, [4]=VeryHigh
	HeatmapColors [5]lipgloss.Color
}

// GetTheme returns a Theme by name (case-insensitive).
// Returns the Dracula theme if the name is not recognized.
// Logs a warning for non-empty invalid names to help users debug config typos.
func GetTheme(name string) Theme {
	normalizedName := strings.ToLower(strings.TrimSpace(name))

	switch normalizedName {
	case "dracula":
		return Dracula
	case "nord":
		return Nord
	case "gruvbox":
		return Gruvbox
	case "monokai":
		return Monokai
	case "solarized":
		return Solarized
	case "tokyonight":
		return TokyoNight
	default:
		// Log warning for non-empty invalid names (empty = first run, expected)
		if normalizedName != "" {
			log.Printf("Warning: unknown theme %q, using 'dracula' instead. Available: %v",
				name, strings.Join(AllThemes(), ", "))
		}
		return Dracula
	}
}

// AllThemes returns all available theme names in display order.
func AllThemes() []string {
	return []string{"dracula", "nord", "gruvbox", "monokai", "solarized", "tokyonight"}
}
