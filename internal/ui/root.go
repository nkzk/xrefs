package ui

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/nkzk/xrefs/internal/models"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Tab interface {
	tea.Model
	Title() string
	FullScreen() bool
	Status() string
	Msg() string
	ShortHelpView() string
}

type Root struct {
	keymap Keymap

	tabs     []Tab
	tabIndex int

	help help.Model
}

type Keymap struct {
	NextTab key.Binding
	PrevTab key.Binding
	Quit    key.Binding
}

func (m Root) Init() tea.Cmd {
	return nil
}

func NewModel(root *models.Resource) *Root {
	h := help.New()
	h.Styles = help.DefaultDarkStyles()

	r := &Root{
		help: h,
		keymap: Keymap{
			NextTab: key.NewBinding(key.WithKeys("tab"), key.WithHelp("\u21E5", "next tab")),
			PrevTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("⇧\u21E5", "prev tab")),
			Quit:    key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		},
	}

	r.tabs = []Tab{
		NewOverview(root, r),
		NewGraph([]unstructured.Unstructured{}, r),
	}

	return r
}

func (m Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.SetWidth(msg.Width)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.NextTab):
			m.tabIndex++
			if m.tabIndex >= len(m.tabs) {
				m.tabIndex = 0
			}

		case key.Matches(msg, m.keymap.PrevTab):
			m.tabIndex--
			if m.tabIndex < 0 {
				m.tabIndex = len(m.tabs) - 1
			}

		case key.Matches(msg, m.keymap.Quit):
			return m, tea.Quit
		}
	}

	updatedModel, cmd := m.tabs[m.tabIndex].Update(msg)
	updatedTab, ok := updatedModel.(Tab)
	if ok {
		m.tabs[m.tabIndex] = updatedTab
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Root) View() tea.View {
	activeTab := m.tabs[m.tabIndex]
	tabView := activeTab.View()

	if activeTab.FullScreen() {
		v := tea.NewView(tabView.Content)
		v.AltScreen = true
		return v
	}

	var s strings.Builder
	s.WriteString("\n ")

	for i, t := range m.tabs {
		style := lipgloss.NewStyle().Bold(true)
		if i == m.tabIndex {
			style = style.
				Underline(true).
				Foreground(lipgloss.Color("212"))
		}
		s.WriteString(style.Render(t.Title()) + "  ")
	}

	s.WriteString("\n" + tabView.Content)

	helpLine := m.help.ShortHelpView([]key.Binding{
		m.keymap.NextTab,
		m.keymap.PrevTab,
		m.keymap.Quit,
	})

	if tabHelpLine := activeTab.ShortHelpView(); tabHelpLine != "" {
		helpLine += " " + tabHelpLine
	}

	if status := activeTab.Status(); status != "" {
		helpLine += " • " + status
	}

	if msg := activeTab.Msg(); msg != "" {
		helpLine += " • " + msg
	}

	s.WriteString("\n" + helpLine)

	v := tea.NewView(s.String())
	v.AltScreen = true
	return v
}
