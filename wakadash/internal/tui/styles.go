package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/b00y0h/wakadash/internal/theme"
)

// BorderStyle returns a rounded border style using the theme's border color.
func BorderStyle(t theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border)
}

// TitleStyle returns a bold style for section titles using the theme's title color.
func TitleStyle(t theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Title)
}

// DimStyle returns a style for secondary text using the theme's dim color.
func DimStyle(t theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Dim)
}

// ErrorStyle returns a style for error messages using the theme's error color.
func ErrorStyle(t theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Error)
}

// WarningStyle returns a style for warnings using the theme's warning color.
func WarningStyle(t theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Warning)
}

// SuccessStyle returns a style for success messages using the theme's success color.
func SuccessStyle(t theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Success)
}
