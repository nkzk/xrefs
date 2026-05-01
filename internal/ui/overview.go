package ui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/nkzk/xrefs/internal/models"
)

type Overview struct {
	parent            *Root
	list              list.Model
	resourceViewModel resourceViewModel
	showViewport      bool

	root          *models.Resource
	rootUpdatedAt time.Time

	keyMap OverviewKeymap

	help help.Model

	msg string
}

var _ Tab = (*Overview)(nil)

type OverviewKeymap struct {
	Navigate key.Binding
	Yaml     key.Binding
	Back     key.Binding
	Expand   key.Binding
}

func NewOverview(root *models.Resource, parent *Root) *Overview {
	delegate := NewResourceListDelegate()

	l := list.New(visibleResourceItems(root), delegate, 120, 24)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)

	return &Overview{
		parent:            parent,
		list:              l,
		root:              root,
		help:              help.New(),
		resourceViewModel: newResourceViewModel(),
		keyMap: OverviewKeymap{
			Navigate: key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "navigate")),
			Yaml:     key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yaml")),
			Back:     key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			Expand:   key.NewBinding(key.WithKeys("x", "space"), key.WithHelp("x/space", "expand")),
		},
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

func (m Overview) Init() tea.Cmd { return nil }

func (m Overview) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case UpdateResourceMsg:
		m.root = msg.Resource
		m.rootUpdatedAt = time.Now()
		return m, m.list.SetItems(visibleResourceItems(msg.Resource))
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

			return m, tea.Quit

		case "y", "enter":
			if !m.showViewport {
				if r := m.selectedResource(); r != nil {
					m.resourceViewModel.SetResource(r)
					m.showViewport = true
					return m, nil
				}
			}

		case "x", "space":
			if !m.showViewport {
				if r := m.selectedResource(); r != nil && len(r.Children) > 0 {
					r.Expanded = !r.Expanded
					cmd := m.list.SetItems(visibleResourceItems(m.root))
					if r.Expanded && !r.ChildrenLoaded {
						return m, tea.Batch(cmd, func() tea.Msg {
							return ExpandNodeMsg{Resource: r}
						})
					}
					return m, cmd
				}
			}
		}

	case tea.WindowSizeMsg:
		var cmd tea.Cmd
		h, v := docStyle.GetFrameSize()
		m.resourceViewModel, cmd = m.resourceViewModel.Update(msg)
		m.list.SetSize(msg.Width-h, msg.Height-v-4)
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

func (m Overview) selectedResource() *models.Resource {
	item, ok := m.list.SelectedItem().(ResourceItem)
	if !ok {
		return nil
	}
	return item.Resource
}

func (m Overview) View() tea.View {
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

	body := columns + "\n" + m.list.View()

	return tea.NewView(docStyle.Render(body))
}

func (m Overview) Title() string {
	return "Overview"
}

func (m Overview) FullScreen() bool {
	return m.showViewport
}

func (m Overview) Msg() string {
	return m.msg
}

func (m Overview) Status() string {
	if !m.rootUpdatedAt.IsZero() {
		return fmt.Sprintf("last update: %s", m.rootUpdatedAt.Format("15:04:05"))
	}
	return "not updated yet"
}

func (m Overview) ShortHelpView() string {
	return m.help.ShortHelpView([]key.Binding{
		m.keyMap.Navigate,
		m.keyMap.Back,
		m.keyMap.Yaml,
		m.keyMap.Expand,
	})
}
