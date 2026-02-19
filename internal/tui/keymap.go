package tui

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	Quit    key.Binding
	Help    key.Binding
	Refresh key.Binding
	Toggle1 key.Binding // Languages
	Toggle2 key.Binding // Projects
	Toggle3 key.Binding // Sparkline
	Toggle4 key.Binding // Heatmap
}

// ShortHelp returns bindings shown in compact help view
func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns bindings shown in expanded help view
func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help, k.Quit},
		{k.Refresh},
		{k.Toggle1, k.Toggle2, k.Toggle3, k.Toggle4},
	}
}

var defaultKeymap = keymap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh now"),
	),
	Toggle1: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "toggle languages"),
	),
	Toggle2: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "toggle projects"),
	),
	Toggle3: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "toggle sparkline"),
	),
	Toggle4: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "toggle heatmap"),
	),
}
