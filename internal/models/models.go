package models

import (
	"fmt"
	"io"
	"strings"

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
	Depth    int
}

func flatten(rs []Resource, depth int) []list.Item {
	var out []list.Item
	for _, r := range rs {
		r.Depth = depth
		out = append(out, r)
		out = append(out, flatten(r.Children, depth+1)...)
	}
	return out
}

// implement item interface
func (r Resource) Title() string       { return r.Name }
func (r Resource) Description() string { return r.Namespace }
func (r Resource) FilterValue() string { return r.Name }

func NewResource(name, namespace string) *Resource {
	return &Resource{}
}

type Model struct {
	list list.Model

	resource Resource
}

type resourceDelegate struct {
	selected lipgloss.Style
	normal   lipgloss.Style
}

func (d resourceDelegate) Height() int  { return 1 }
func (d resourceDelegate) Spacing() int { return 0 }
func (d resourceDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d resourceDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	r := item.(Resource)

	indent := strings.Repeat("  ", r.Depth)

	prefix := ""
	if r.Depth > 0 {
		prefix = "└─ "
	}

	line := indent + prefix + r.Name
	if r.Namespace != "" {
		line += " " + r.Namespace
	}

	if index == m.Index() {
		fmt.Fprint(w, d.selected.Render(line))
		return
	}
	fmt.Fprint(w, d.normal.Render(line))
}
func NewModel() *Model {
	white := lipgloss.Color("#ffffff")
	grey := lipgloss.Color("#9b9b9b")

	delegate := resourceDelegate{
		selected: lipgloss.NewStyle().Foreground(white),
		normal:   lipgloss.NewStyle().Foreground(grey),
	}

	l := list.New(make([]list.Item, 0), delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)

	return &Model{
		list:     l,
		resource: *NewResource("name", "namespace"),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

type addResourcesMsg struct {
	resources []Resource
}

func addResourceCmd(rs []Resource) tea.Cmd {
	return func() tea.Msg {
		return addResourcesMsg{resources: rs}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case addResourcesMsg:
		items := append([]list.Item{}, m.list.Items()...)
		items = append(items, flatten(msg.resources, 0)...)
		return m, m.list.SetItems(items)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "a":
			return m, addResourceCmd([]Resource{
				{
					Name: "parent",
					Children: []Resource{
						{
							Name: "child-1",
							Children: []Resource{
								{
									Name:      "child-child",
									Namespace: "namespace",
								},
							},
						},
						{Name: "child-2"},
					},
				},
			})
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
