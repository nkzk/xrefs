package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/nkzk/xrefs/internal/models"
)

type Model struct {
	list list.Model

	resourceViewModel resourceViewModel
	showViewport      bool

	root          *models.Resource
	rootUpdatedAt time.Time
}

func NewModel(root *models.Resource) *Model {
	delegate := NewResourceDelegate()

	l := list.New(flatten(*root, 0), delegate, 120, 24)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)

	return &Model{
		list:              l,
		root:              root,
		resourceViewModel: newResourceViewModel(),
	}
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type (
	UpdateResourceMsg struct {
		Resource *models.Resource
	}

	ExpandNodeMsg struct {
		Resource *models.Resource
	}

	QuitMsg    struct{}
	RootErrMsg struct {
		Err error
	}
	RootDeletedMsg struct{}
)

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case UpdateResourceMsg:
		m.root = msg.Resource
		m.rootUpdatedAt = time.Now()
		return m, m.list.SetItems(flatten(*msg.Resource, 0))
	case tea.KeyPressMsg:
		if m.list.FilterState() == list.Filtering {
			break // let list handle it
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q":
			if m.showViewport {
				m.showViewport = false
				return m, nil
			}
			return m, nil

		case "y", "enter":
			if !m.showViewport {
				selected, ok := m.list.SelectedItem().(models.Resource)
				if ok {
					m.resourceViewModel.SetResource(&selected)
					m.showViewport = true
					return m, nil
				}
			}

		case "x", "space":
			if !m.showViewport {
				selected, ok := m.list.SelectedItem().(models.Resource)
				if ok {
					node := findResourceByID(m.root, selected.ID)
					if node != nil && len(node.Children) > 0 {
						node.Expanded = !node.Expanded
						cmd := m.list.SetItems(flatten(*m.root, 0))
						if node.Expanded && !node.ChildrenLoaded {
							return m, tea.Batch(cmd, func() tea.Msg {
								return ExpandNodeMsg{Resource: node}
							})
						}
						return m, cmd
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		var cmd tea.Cmd
		h, v := docStyle.GetFrameSize()
		m.resourceViewModel, cmd = m.resourceViewModel.Update(msg)
		m.list.SetSize(msg.Width-h, msg.Height-v-1)
		return m, cmd
	}

	if m.showViewport {
		var cmd tea.Cmd
		m.resourceViewModel, cmd = m.resourceViewModel.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() tea.View {
	if m.showViewport {
		return m.resourceViewModel.View()
	}

	columns := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6f6f6f")).
		Render(fmt.Sprintf(
			"%-64s  %-15s %-13s %-14s %s",
			"RESOURCE",
			"NAMESPACE",
			"READY",
			"SYNCED",
			"REASON",
		))

	status := "not updated yet"
	if !m.rootUpdatedAt.IsZero() {
		status = fmt.Sprintf("last update: %s", m.rootUpdatedAt.Format("15:04:05"))
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6f6f6f")).
		Render("↑/↓ navigate • y inspect • ctrl+c quit • " + status)

	body := strings.Join([]string{
		columns,
		m.list.View(),
		footer,
	}, "\n")

	v := tea.NewView(docStyle.Render(body))
	v.AltScreen = true
	return v
}

type resourceDelegate struct {
	selected lipgloss.Style
	normal   lipgloss.Style
	notFound lipgloss.Style
	err      lipgloss.Style
}

func NewResourceDelegate() resourceDelegate {
	return resourceDelegate{
		selected: lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true),
		normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("#9b9b9b")),
		notFound: lipgloss.NewStyle().Foreground(lipgloss.Color("#ff9898")),
		err:      lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f5f")),
	}
}

func (d resourceDelegate) Height() int  { return 1 }
func (d resourceDelegate) Spacing() int { return 0 }
func (d resourceDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d resourceDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	r := item.(models.Resource)

	ready := condStatus(r, "Ready")
	synced := condStatus(r, "Synced")
	reason := shorten(condReason(r), 40)

	// 🔥 override if error
	if r.Error != nil {
		ready = "-"
		synced = "-"
		reason = shorten(r.Error.Error(), 40)
	}

	if r.NotFound {
		ready = "-"
		synced = "-"
		reason = "NOT FOUND"
	}

	row := fmt.Sprintf(
		"%-64s  %-15s %-13s %-14s %s",
		treeName(r),
		namespace(r),
		ready,
		synced,
		reason,
	)

	isSelected := index == m.Index()

	switch {
	case r.NotFound && isSelected:
		fmt.Fprint(w, d.notFound.Bold(true).Render(row))

	case r.NotFound:
		fmt.Fprint(w, d.notFound.Render(row))

	case isSelected:
		fmt.Fprint(w, d.selected.Render(row))

	default:
		fmt.Fprint(w, d.normal.Render(row))
	}
}

func flatten(r models.Resource, depth int) []list.Item {
	return flattenWithPrefix(r, depth, true, "")
}

func flattenWithPrefix(r models.Resource, depth int, isLast bool, prefix string) []list.Item {
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

	if !r.Expanded {
		return out
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

// findResourceByID traverses the resource tree and returns a pointer to the node with the given ID.
func findResourceByID(r *models.Resource, id string) *models.Resource {
	if r.ID == id {
		return r
	}
	for i := range r.Children {
		if found := findResourceByID(&r.Children[i], id); found != nil {
			return found
		}
	}
	return nil
}

func treeName(r models.Resource) string {
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

	label := fmt.Sprintf("%s/%s", kind, name)

	if len(r.Children) > 0 {
		if r.Expanded {
			label = "▼ " + label
		} else {
			label = "▶ " + label
		}
	}

	const maxCol = 64
	maxLabel := maxCol - len(prefix)
	if maxLabel > 0 && len(label) > maxLabel {
		label = label[:maxLabel-1] + "…"
	}

	return prefix + label
}

func namespace(r models.Resource) string {
	if r.Unstructured != nil && r.Unstructured.GetNamespace() != "" {
		return r.Unstructured.GetNamespace()
	}
	if r.Ref != nil {
		return r.Ref.Namespace
	}
	return "-"
}

func condStatus(r models.Resource, name string) string {
	c := r.Conditions.Get(name)
	if c.Status == "" {
		return "-"
	}
	return c.Status
}

func condReason(r models.Resource) string {
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
	return s[:head] + "(...)" + s[len(s)-tail:]
}
