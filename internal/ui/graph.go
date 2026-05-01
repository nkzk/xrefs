package ui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Graph struct {
	help   help.Model
	list   list.Model
	keyMap OverviewKeymap

	resources []unstructured.Unstructured
}

var _ Tab = (*Graph)(nil)

type DependencyKeymap struct {
	Back key.Binding
}

func NewGraph(resources []unstructured.Unstructured, parent *Root) *Graph {
	l := list.New(toList(resources), list.NewDefaultDelegate(), 120, 24)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)

	return &Graph{
		list: l,
		help: help.New(),
		keyMap: OverviewKeymap{
			Back: key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "back")),
		},
		resources: resources,
	}
}

func (m Graph) Init() tea.Cmd { return nil }

func (m Graph) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m Graph) View() tea.View {
	return tea.NewView("helo")
}

func (m Graph) Title() string {
	return "Graph"
}

func (m Graph) FullScreen() bool {
	return false
}

func (m Graph) Status() string {
	return ""
}

func (m Graph) ShortHelpView() string {
	return m.help.ShortHelpView([]key.Binding{
		m.keyMap.Back,
	})
}

func (m Graph) Msg() string {
	return ""
}

func toList(resources []unstructured.Unstructured) []list.Item {
	l := []list.Item{}

	return l
}
