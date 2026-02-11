package theme

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

var (
	ColorPrimary    = lipgloss.Color("#A1D6FE")
	ColorSecondary  = lipgloss.Color("241")
	ColorError      = lipgloss.Color("#bd534b")
	ColorAlert      = lipgloss.Color("62")
	ColorSuccess    = lipgloss.Color("#00FF7F")
	ColorFailure    = lipgloss.Color("#FF5555")
	ColorSelected   = lipgloss.Color("#ffffffff")
	ColorSelectedFg = lipgloss.Color("#000000")
)

var renderer = lipgloss.NewRenderer(os.Stdout)

var (
	BaseStyle     = renderer.NewStyle().Bold(true).Foreground(ColorPrimary)
	HeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Align(lipgloss.Left)
	FooterStyle   = lipgloss.NewStyle().Foreground(ColorSecondary).Faint(true).PaddingLeft(2)
	SelectedStyle = lipgloss.NewStyle().Background(ColorSelected).Foreground(ColorSelectedFg).Bold(true)
	ErrorStyle    = lipgloss.NewStyle().Foreground(ColorError)
	AlertStyle    = lipgloss.NewStyle().Foreground(ColorAlert)
	SuccessStyle  = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	FailureStyle  = lipgloss.NewStyle().Foreground(ColorFailure).Bold(true)
	DocStyle      = lipgloss.NewStyle().Margin(0, 2)
)

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

func Error(msg string) string {
	return ErrorStyle.Render(fmt.Sprintf("error: %s", msg))
}

func TableCellStyle(row, col, cursor int, value string, isHeader bool) lipgloss.Style {
	if isHeader {
		return HeaderStyle.Inherit(BaseStyle)
	}

	isSelected := row == cursor

	switch value {
	case "False", "no":
		if isSelected {
			return SelectedStyle.Inherit(BaseStyle).Foreground(ColorFailure).Bold(true)
		}
		return BaseStyle.Foreground(ColorFailure).Bold(true)
	case "True", "yes":
		if isSelected {
			return SelectedStyle.Inherit(BaseStyle).Foreground(ColorSuccess).Bold(true)
		}
		return BaseStyle.Foreground(ColorSuccess).Bold(true)
	}

	if isSelected {
		return SelectedStyle.Inherit(BaseStyle)
	}
	return BaseStyle
}

func NewSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Jump
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return s
}
