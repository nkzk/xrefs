package constants

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DocStyle styling for viewports
var DocStyle = lipgloss.NewStyle().Margin(0, 2)

// ErrStyle provides styling for error messages
var ErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#bd534b")).Render

// AlertStyle provides styling for alert messages
var AlertStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Render

var WindowSize tea.WindowSizeMsg

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
		key.WithHelp("enter/y", "yaml"),
	),
	Describe: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "describe"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("ctrl+c/q", "quit"),
	),

	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("arrow-up/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("down/j", "down"),
	),
	Top: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "top"),
	),

	Help: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "help"),
	),
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Enter, k.Describe},
		{k.Up, k.Down, k.Top, k.Bottom},
		{k.Back, k.Quit},
	}
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Back, k.Quit}
}
