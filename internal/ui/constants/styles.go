package constants

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// DocStyle styling for viewports
var DocStyle = lipgloss.NewStyle().Margin(0, 2)

// ErrStyle provides styling for error messages
var ErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#bd534b")).Render

// AlertStyle provides styling for alert messages
var AlertStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Render

// Styling for selected rows
var SelectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#ffffffff")).Foreground(lipgloss.Color("#000000")).Bold(true)

var BaseStyle = lipgloss.NewRenderer(os.Stdout).NewStyle().Bold(true).Foreground(lipgloss.Color("#A1D6FE"))

// styles.go
var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#A1D6FE")).
			Align(lipgloss.Left)

	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Faint(true).
			PaddingLeft(2)
)

// Header renders the header text; shows spinner if updating.
func Header(text string, spin spinner.Model, updating bool, width int) string {
	s := text
	if updating {
		s = fmt.Sprintf("%s %s", text, spin.View())
	}
	return HeaderStyle.Width(width).Render(s)
}

func Footer(helpView string, width int) string {
	return FooterStyle.Width(width).Render(helpView)
}
