package ui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nkzk/xrefs/internal/models"
	v1 "k8s.io/api/core/v1"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type (
	UpdateResourceMsg struct {
		Resource *models.Resource
	}

	QuitMsg    struct{}
	RootErrMsg struct {
		Err error
	}
	RootDeletedMsg  struct{}
	RootNotFoundMsg struct {
		Ref *v1.ObjectReference
	}
)

type Model struct {
	list list.Model
	root *models.Resource
}

func NewModel(root *models.Resource) *Model {
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
	r := item.(models.Resource)

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
