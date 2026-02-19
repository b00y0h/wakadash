package tui

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	Quit    key.Binding
	Help    key.Binding
	Refresh key.Binding
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
}
