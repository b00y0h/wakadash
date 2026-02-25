package tui

import (
	"strings"

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

// renderBorderedPanel renders content inside a bordered panel with a centered title.
// Uses lipgloss's built-in border handling for proper ANSI code support.
func renderBorderedPanel(title, content string, width int, t theme.Theme) string {
	// Style for borders
	borderColor := lipgloss.Color("#666666")

	// Create border style with rounded corners
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(width)

	// Wrap content in border
	bordered := borderStyle.Render(content)

	// Get the bordered content as lines
	lines := strings.Split(bordered, "\n")
	if len(lines) == 0 {
		return bordered
	}

	// Replace the top border line with one that has the centered title
	topLine := lines[0]
	topLineWidth := lipgloss.Width(topLine)

	// Build new top line with centered title
	titleStyle := TitleStyle(t)
	styledTitle := titleStyle.Render(title)
	titleWidth := lipgloss.Width(styledTitle)

	// Calculate padding for centered title (account for corners)
	innerWidth := topLineWidth - 2 // minus corner characters
	if innerWidth < titleWidth {
		innerWidth = titleWidth
	}
	availableSpace := innerWidth - titleWidth
	leftPad := availableSpace / 2
	rightPad := availableSpace - leftPad

	// Build the new top border with title
	dimStyle := lipgloss.NewStyle().Foreground(borderColor)
	newTopLine := dimStyle.Render("╭") +
		dimStyle.Render(strings.Repeat("─", leftPad)) +
		styledTitle +
		dimStyle.Render(strings.Repeat("─", rightPad)) +
		dimStyle.Render("╮")

	lines[0] = newTopLine
	return strings.Join(lines, "\n")
}

// EndOfHistoryStyle creates a centered box/border style for the end-of-history banner.
// Per user decision: Box/border around message — emphasized style, draws attention like a modal.
func EndOfHistoryStyle(t theme.Theme, width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-4).
		Height(height-4).
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Warning).
		Padding(2, 4)
}

// EndOfHistoryTitleStyle creates the title style for the banner.
func EndOfHistoryTitleStyle(t theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Warning).
		Bold(true).
		MarginBottom(2)
}

// EndOfHistoryTextStyle creates the body text style for the banner.
func EndOfHistoryTextStyle(t theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Foreground).
		MarginBottom(1)
}

// EndOfHistoryHintStyle creates the navigation hint style.
func EndOfHistoryHintStyle(t theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Dim).
		Italic(true)
}
