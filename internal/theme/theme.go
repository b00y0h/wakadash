// Package theme provides theming support for wakadash.
package theme

import "github.com/charmbracelet/lipgloss"

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

// GetTheme returns a Theme by name.
// Returns the Dracula theme if the name is not recognized.
func GetTheme(name string) Theme {
	switch name {
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
		return Dracula
	}
}

// AllThemes returns all available theme names in display order.
func AllThemes() []string {
	return []string{"dracula", "nord", "gruvbox", "monokai", "solarized", "tokyonight"}
}
