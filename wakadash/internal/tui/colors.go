package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// languageColors maps programming languages to their GitHub Linguist colors
var languageColors = map[string]string{
	"go":         "#00ADD8",
	"javascript": "#f1e05a",
	"typescript": "#3178c6",
	"python":     "#3572A5",
	"rust":       "#dea584",
	"ruby":       "#701516",
	"java":       "#b07219",
	"c":          "#555555",
	"c++":        "#f34b7d",
	"html":       "#e34c26",
	"css":        "#563d7c",
	"shell":      "#89e051",
	"markdown":   "#083fa1",
	"php":        "#4F5D95",
	"swift":      "#F05138",
	"kotlin":     "#A97BFF",
	"scala":      "#c22d40",
	"vue":        "#41b883",
	"dart":       "#00B4AB",
	"elixir":     "#6e4a7e",
}

// getLanguageColor returns the GitHub Linguist color for a given language.
// Returns a default gray for unknown languages.
func getLanguageColor(name string) lipgloss.Color {
	// Case-insensitive matching
	lowerName := strings.ToLower(name)

	if color, ok := languageColors[lowerName]; ok {
		return lipgloss.Color(color)
	}

	// Default gray for unknown languages
	return lipgloss.Color("#cccccc")
}
