package constants

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	Help key.Binding

	Enter    key.Binding
	Back     key.Binding
	Quit     key.Binding
	Yaml     key.Binding
	Describe key.Binding

	Up   key.Binding
	Down key.Binding

	Top    key.Binding
	Bottom key.Binding
}

var Keymap = keymap{
	Enter: key.NewBinding(
		key.WithKeys("enter", "y"),
		key.WithHelp("<y>", "yaml "),
	),
	Describe: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("<d>", "describe "),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q", "esc"),
		key.WithHelp("<esc>/<q>", "quit "),
	),

	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("<↑>/<k>", "up "),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("<↓>/<j>", "down "),
	),
	Top: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("<g>", "top "),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("<G>", "top "),
	),

	Help: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("<h>", "help "),
	),
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Top, k.Bottom},
		{k.Enter, k.Describe, k.Quit},
	}
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Help,
		k.Quit,
	}
}
