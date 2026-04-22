package models

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type Resource struct {
	schema.GroupVersionKind
	Name      string
	Namespace string

	ID       string
	Parent   *Resource
	Children []Resource
}

func NewResource(name, namespace string) *Resource {
	return &Resource{}
}

type Model struct {
	list list.Model

	resource Resource
}

func NewModel() *Model {
	return &Model{
		list:     list.New(list.Item, list.NewDefaultDelegate(), 0, 0),
		resource: *NewResource("name", "namespace"),
	}
}

func (m Model) Title() string       { return m.resource.Name }
func (m Model) Description() string { return m.resource.Namespace }
func (m Model) FilterValue() string { return "" }

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() tea.View {
	v := tea.NewView(docStyle.Render(m.list.View()))
	v.AltScreen = true
	return v
}
