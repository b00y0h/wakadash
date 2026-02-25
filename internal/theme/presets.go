package theme

import "github.com/charmbracelet/lipgloss"

// Dracula theme based on draculatheme.com
var Dracula = Theme{
	Name:       "Dracula",
	Background: lipgloss.Color("#282a36"),
	Foreground: lipgloss.Color("#f8f8f2"),
	Border:     lipgloss.Color("#6272a4"),
	Title:      lipgloss.Color("#bd93f9"),
	Dim:        lipgloss.Color("#6272a4"),
	Error:      lipgloss.Color("#ff5555"),
	Warning:    lipgloss.Color("#ffb86c"),
	Success:    lipgloss.Color("#50fa7b"),
	Primary:    lipgloss.Color("#bd93f9"),
	Secondary:  lipgloss.Color("#8be9fd"),
	Accent1:    lipgloss.Color("#ff79c6"),
	Accent2:    lipgloss.Color("#f1fa8c"),
	Accent3:    lipgloss.Color("#50fa7b"),
	Accent4:    lipgloss.Color("#ffb86c"),
	HeatmapColors: [5]lipgloss.Color{
		lipgloss.Color("#282a36"), // None
		lipgloss.Color("#583c7c"), // Low
		lipgloss.Color("#7952a3"), // Medium
		lipgloss.Color("#9d6fca"), // High
		lipgloss.Color("#bd93f9"), // VeryHigh
	},
}

// Nord theme based on nordtheme.com
var Nord = Theme{
	Name:       "Nord",
	Background: lipgloss.Color("#2e3440"),
	Foreground: lipgloss.Color("#d8dee9"),
	Border:     lipgloss.Color("#4c566a"),
	Title:      lipgloss.Color("#88c0d0"),
	Dim:        lipgloss.Color("#4c566a"),
	Error:      lipgloss.Color("#bf616a"),
	Warning:    lipgloss.Color("#ebcb8b"),
	Success:    lipgloss.Color("#a3be8c"),
	Primary:    lipgloss.Color("#88c0d0"),
	Secondary:  lipgloss.Color("#81a1c1"),
	Accent1:    lipgloss.Color("#8fbcbb"),
	Accent2:    lipgloss.Color("#5e81ac"),
	Accent3:    lipgloss.Color("#b48ead"),
	Accent4:    lipgloss.Color("#d08770"),
	HeatmapColors: [5]lipgloss.Color{
		lipgloss.Color("#2e3440"), // None
		lipgloss.Color("#4d7a8c"), // Low
		lipgloss.Color("#5d96aa"), // Medium
		lipgloss.Color("#6fabbf"), // High
		lipgloss.Color("#88c0d0"), // VeryHigh
	},
}

// Gruvbox theme based on github.com/morhetz/gruvbox
var Gruvbox = Theme{
	Name:       "Gruvbox",
	Background: lipgloss.Color("#282828"),
	Foreground: lipgloss.Color("#ebdbb2"),
	Border:     lipgloss.Color("#504945"),
	Title:      lipgloss.Color("#fabd2f"),
	Dim:        lipgloss.Color("#665c54"),
	Error:      lipgloss.Color("#fb4934"),
	Warning:    lipgloss.Color("#fe8019"),
	Success:    lipgloss.Color("#b8bb26"),
	Primary:    lipgloss.Color("#83a598"),
	Secondary:  lipgloss.Color("#d3869b"),
	Accent1:    lipgloss.Color("#8ec07c"),
	Accent2:    lipgloss.Color("#fabd2f"),
	Accent3:    lipgloss.Color("#fb4934"),
	Accent4:    lipgloss.Color("#fe8019"),
	HeatmapColors: [5]lipgloss.Color{
		lipgloss.Color("#282828"), // None
		lipgloss.Color("#4d6656"), // Low
		lipgloss.Color("#5e8066"), // Medium
		lipgloss.Color("#709a77"), // High
		lipgloss.Color("#8ec07c"), // VeryHigh
	},
}

// Monokai theme based on monokai.pro
var Monokai = Theme{
	Name:       "Monokai",
	Background: lipgloss.Color("#272822"),
	Foreground: lipgloss.Color("#f8f8f2"),
	Border:     lipgloss.Color("#49483e"),
	Title:      lipgloss.Color("#66d9ef"),
	Dim:        lipgloss.Color("#75715e"),
	Error:      lipgloss.Color("#f92672"),
	Warning:    lipgloss.Color("#fd971f"),
	Success:    lipgloss.Color("#a6e22e"),
	Primary:    lipgloss.Color("#ae81ff"),
	Secondary:  lipgloss.Color("#66d9ef"),
	Accent1:    lipgloss.Color("#f92672"),
	Accent2:    lipgloss.Color("#fd971f"),
	Accent3:    lipgloss.Color("#a6e22e"),
	Accent4:    lipgloss.Color("#e6db74"),
	HeatmapColors: [5]lipgloss.Color{
		lipgloss.Color("#272822"), // None
		lipgloss.Color("#5d4a7c"), // Low
		lipgloss.Color("#7a5ca3"), // Medium
		lipgloss.Color("#9770ca"), // High
		lipgloss.Color("#ae81ff"), // VeryHigh
	},
}

// Solarized (Dark) theme based on ethanschoonover.com/solarized
var Solarized = Theme{
	Name:       "Solarized",
	Background: lipgloss.Color("#002b36"),
	Foreground: lipgloss.Color("#839496"),
	Border:     lipgloss.Color("#073642"),
	Title:      lipgloss.Color("#268bd2"),
	Dim:        lipgloss.Color("#586e75"),
	Error:      lipgloss.Color("#dc322f"),
	Warning:    lipgloss.Color("#cb4b16"),
	Success:    lipgloss.Color("#859900"),
	Primary:    lipgloss.Color("#268bd2"),
	Secondary:  lipgloss.Color("#2aa198"),
	Accent1:    lipgloss.Color("#6c71c4"),
	Accent2:    lipgloss.Color("#b58900"),
	Accent3:    lipgloss.Color("#859900"),
	Accent4:    lipgloss.Color("#d33682"),
	HeatmapColors: [5]lipgloss.Color{
		lipgloss.Color("#002b36"), // None
		lipgloss.Color("#0f5173"), // Low
		lipgloss.Color("#1a6d96"), // Medium
		lipgloss.Color("#2088b8"), // High
		lipgloss.Color("#268bd2"), // VeryHigh
	},
}

// TokyoNight (Storm variant) theme based on github.com/folke/tokyonight.nvim
var TokyoNight = Theme{
	Name:       "Tokyo Night",
	Background: lipgloss.Color("#1a1b26"),
	Foreground: lipgloss.Color("#c0caf5"),
	Border:     lipgloss.Color("#3b4261"),
	Title:      lipgloss.Color("#7aa2f7"),
	Dim:        lipgloss.Color("#565f89"),
	Error:      lipgloss.Color("#f7768e"),
	Warning:    lipgloss.Color("#e0af68"),
	Success:    lipgloss.Color("#9ece6a"),
	Primary:    lipgloss.Color("#bb9af7"),
	Secondary:  lipgloss.Color("#7aa2f7"),
	Accent1:    lipgloss.Color("#73daca"),
	Accent2:    lipgloss.Color("#ff9e64"),
	Accent3:    lipgloss.Color("#9ece6a"),
	Accent4:    lipgloss.Color("#2ac3de"),
	HeatmapColors: [5]lipgloss.Color{
		lipgloss.Color("#1a1b26"), // None
		lipgloss.Color("#6d5bab"), // Low
		lipgloss.Color("#8574c3"), // Medium
		lipgloss.Color("#9d8edb"), // High
		lipgloss.Color("#bb9af7"), // VeryHigh
	},
}
