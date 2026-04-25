package models

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type RootDeletedMsg struct{}
type RootNotFoundMsg struct{}
type RootErrMsg struct {
	Err error
}

type UpdateResourceMsg struct {
	Resource *Resource
}

type QuitMsg struct{}

type ResourceList struct {
	Items []*Resource
	Error error
}

type Resource struct {
	Unstructured unstructured.Unstructured

	ID       string
	Parent   *Resource
	Children []Resource
	Depth    int

	Error error
}

func flatten(r Resource, depth int) []list.Item {
	r.Depth = depth

	out := []list.Item{r}

	for _, child := range r.Children {
		out = append(out, flatten(child, depth+1)...)
	}

	return out
}

// implement item interface
func (r Resource) Title() string       { return r.Unstructured.GetName() }
func (r Resource) Description() string { return r.Unstructured.GetNamespace() }
func (r Resource) FilterValue() string { return r.Unstructured.GetName() }

func NewResource(name, namespace string) *Resource {
	return &Resource{}
}

type Model struct {
	list list.Model

	root *Resource
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

	line := fmt.Sprintf(
		"%s%s%s (%s)\n%s  %s | %s",
		indent,
		prefix,
		r.Unstructured.GetName(),
		r.Unstructured.GetNamespace(),
		indent, "status", "r")
	if index == m.Index() {
		fmt.Fprint(w, d.selected.Render(line))
		return
	}

	fmt.Fprint(w, d.normal.Render(line))
}

func NewModel(root *Resource) *Model {
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
		list: l,
		root: root,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

type updateResourceMsg struct {
	resources Resource
}

func updateresourcesCmd(rs Resource) tea.Cmd {
	return func() tea.Msg {
		return updateResourceMsg{resources: rs}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case updateResourceMsg:
		items := append([]list.Item{}, m.list.Items()...)
		items = append(items, flatten(msg.resources, 0)...)
		return m, m.list.SetItems(items)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
			// case "a":
			// 	return m, updateresourcesCmd(Resource{
			// 		Unstructured: unstructured.Unstructured{
			// 			Object: map[string]any{
			// 				"apiVersion": "v1",
			// 				"kind":       "Deployment",
			// 				"metadata": map[string]any{
			// 					"name":      "parent-resource",
			// 					"namespace": "default",
			// 				},
			// 			},
			// 		},
			// 		Children: []Resource{
			// 			{
			// 				Unstructured: unstructured.Unstructured{
			// 					Object: map[string]any{
			// 						"apiVersion": "v1",
			// 						"kind":       "ConfigMap",
			// 						"metadata": map[string]any{
			// 							"name":      "child-1",
			// 							"namespace": "default",
			// 						},
			// 					},
			// 				},
			// 			},
			// 			{
			// 				Unstructured: unstructured.Unstructured{
			// 					Object: map[string]any{
			// 						"apiVersion": "v1",
			// 						"kind":       "ConfigMap",
			// 						"metadata": map[string]any{
			// 							"name":      "child-2",
			// 							"namespace": "default",
			// 						},
			// 					},
			// 				},
			// 			},
			// 		},
			// 	})
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
