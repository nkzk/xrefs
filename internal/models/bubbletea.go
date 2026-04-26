package models

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type UpdateResourceMsg struct {
	Resource *Resource
}

type Condition struct {
	Status             string `json:"status"`
	ConditionType      string `json:"type"`
	Reason             string `json:"reason"`
	LastTransitionTime string `json:"lastTransitionTime"`
}

type Conditions []Condition

func (c Conditions) Get(t string) Condition {
	for _, cond := range c {
		if cond.ConditionType == t {
			return cond
		}
	}
	return Condition{}
}

type Resource struct {
	Ref          *v1.ObjectReference
	Unstructured *unstructured.Unstructured
	Conditions   Conditions

	ID       string
	Parent   *Resource
	Children []Resource

	Depth  int
	IsLast bool
	Prefix string

	Error error
}

func NewResource(
	parent *Resource,
	unstructured *unstructured.Unstructured,
	ref *v1.ObjectReference,
) *Resource {
	return &Resource{
		ID:           uuid.New().String(),
		Parent:       parent,
		Unstructured: unstructured,
		Ref:          ref,
	}
}

// ---------- tree flatten ----------

func flatten(r Resource, depth int) []list.Item {
	return flattenWithPrefix(r, depth, true, "")
}

func flattenWithPrefix(r Resource, depth int, isLast bool, prefix string) []list.Item {
	r.Depth = depth
	r.IsLast = isLast
	r.Prefix = prefix

	out := []list.Item{r}

	childPrefix := prefix
	if depth > 0 {
		if isLast {
			childPrefix += "   "
		} else {
			childPrefix += "│  "
		}
	}

	for i, child := range r.Children {
		out = append(out, flattenWithPrefix(
			child,
			depth+1,
			i == len(r.Children)-1,
			childPrefix,
		)...)
	}

	return out
}

// ---------- helpers ----------

func treeName(r Resource) string {
	prefix := ""
	if r.Depth > 0 {
		if r.IsLast {
			prefix = r.Prefix + "└─ "
		} else {
			prefix = r.Prefix + "├─ "
		}
	}

	kind := r.Ref.Kind
	name := r.Ref.Name

	if r.Unstructured != nil {
		if r.Unstructured.GetKind() != "" {
			kind = r.Unstructured.GetKind()
		}
		if r.Unstructured.GetName() != "" {
			name = r.Unstructured.GetName()
		}
	}

	return fmt.Sprintf("%s%s/%s", prefix, kind, name)
}

func namespace(r Resource) string {
	if r.Unstructured != nil && r.Unstructured.GetNamespace() != "" {
		return r.Unstructured.GetNamespace()
	}
	if r.Ref != nil {
		return r.Ref.Namespace
	}
	return "-"
}

func condStatus(r Resource, name string) string {
	c := r.Conditions.Get(name)
	if c.Status == "" {
		return "-"
	}
	return c.Status
}

func condReason(r Resource) string {
	if r := r.Conditions.Get("Ready").Reason; r != "" {
		return r
	}
	if r := r.Conditions.Get("Synced").Reason; r != "" {
		return r
	}
	return "-"
}

func shorten(s string, max int) string {
	if len(s) <= max {
		return s
	}
	head := max / 2
	tail := max - head - 1
	return s[:head] + "…" + s[len(s)-tail:]
}

// ---------- list item impl ----------

func (r Resource) Title() string       { return r.Unstructured.GetName() }
func (r Resource) Description() string { return r.Unstructured.GetNamespace() }
func (r Resource) FilterValue() string { return r.Unstructured.GetName() }

// ---------- delegate ----------

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

	row := fmt.Sprintf(
		"%-48s  %-15s %-13s %-14s %s",
		treeName(r),
		namespace(r),
		condStatus(r, "Ready"),
		condStatus(r, "Synced"),
		shorten(condReason(r), 40),
	)

	if index == m.Index() {
		fmt.Fprint(w, d.selected.Render(row))
		return
	}
	fmt.Fprint(w, d.normal.Render(row))
}

// ---------- model ----------

type Model struct {
	list list.Model
	root *Resource
}

func NewModel(root *Resource) *Model {
	delegate := resourceDelegate{
		selected: lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true),
		normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("#9b9b9b")),
	}

	l := list.New(flatten(*root, 0), delegate, 120, 24)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	return &Model{
		list: l,
		root: root,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case UpdateResourceMsg:
		return m, m.list.SetItems(flatten(*msg.Resource, 0))

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
	columns := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6f6f6f")).
		Render(fmt.Sprintf(
			"%-48s  %-15s %-13s %-14s %s",
			"RESOURCE",
			"NAMESPACE",
			"READY",
			"SYNCED",
			"REASON",
		))

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6f6f6f")).
		Render("↑/↓ navigate • ctrl+c quit")

	body := strings.Join([]string{
		columns,
		m.list.View(),
		footer,
	}, "\n")

	v := tea.NewView(docStyle.Render(body))
	v.AltScreen = true
	return v
}
