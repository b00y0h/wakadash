package tui

import "github.com/charmbracelet/lipgloss"

var (
	// borderStyle wraps panels in a rounded border with a purple-blue color.
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	// titleStyle renders section titles in bold magenta.
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	// dimStyle renders secondary text in a muted gray.
	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// errorStyle renders error messages in red.
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)
