package tui

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	Quit        key.Binding
	Help        key.Binding
	Refresh     key.Binding
	ChangeTheme key.Binding
	Toggle1     key.Binding // Languages
	Toggle2     key.Binding // Projects
	Toggle3     key.Binding // Sparkline
	Toggle4     key.Binding // Heatmap
	Toggle5     key.Binding // Categories
	Toggle6     key.Binding // Editors
	Toggle7     key.Binding // OS
	Toggle8     key.Binding // Machines
	Toggle9     key.Binding // Summary
	ShowAll     key.Binding // Show all panels
	HideAll     key.Binding // Hide all panels
	PrevDay     key.Binding // Navigate to previous day
	NextDay     key.Binding // Navigate to next day
	Today       key.Binding // Return to today
}

// ShortHelp returns bindings shown in compact help view
func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.ChangeTheme, k.Quit}
}

// FullHelp returns bindings shown in expanded help view
func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help, k.Quit},
		{k.Refresh, k.ChangeTheme},
		{k.Toggle1, k.Toggle2, k.Toggle3, k.Toggle4},
		{k.Toggle5, k.Toggle6, k.Toggle7, k.Toggle8, k.Toggle9},
		{k.ShowAll, k.HideAll},
		{k.PrevDay, k.NextDay, k.Today},
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
	ChangeTheme: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "change theme"),
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
	Toggle5: key.NewBinding(
		key.WithKeys("5"),
		key.WithHelp("5", "toggle categories"),
	),
	Toggle6: key.NewBinding(
		key.WithKeys("6"),
		key.WithHelp("6", "toggle editors"),
	),
	Toggle7: key.NewBinding(
		key.WithKeys("7"),
		key.WithHelp("7", "toggle OS"),
	),
	Toggle8: key.NewBinding(
		key.WithKeys("8"),
		key.WithHelp("8", "toggle machines"),
	),
	Toggle9: key.NewBinding(
		key.WithKeys("9"),
		key.WithHelp("9", "toggle summary"),
	),
	ShowAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "show all panels"),
	),
	HideAll: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "hide all panels"),
	),
	PrevDay: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "previous day"),
	),
	NextDay: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "next day"),
	),
	Today: key.NewBinding(
		key.WithKeys("0", "home"),
		key.WithHelp("0/home", "return to today"),
	),
}
