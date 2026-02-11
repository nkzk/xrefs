package models

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nkzk/xrefs/internal/ui/v2/messages"
	"github.com/nkzk/xrefs/internal/ui/v2/store"
	"github.com/nkzk/xrefs/internal/ui/v2/theme"
)

const refreshInterval = 7 * time.Second

type RootModel struct {
	table   *TableScreen
	content *ContentScreen
	current tea.Model

	store *store.Store

	spinner  spinner.Model
	help     help.Model
	keys     theme.KeyMap
	width    int
	height   int
	title    string
	err      error
	updating bool
}

func NewRootModel(s *store.Store, title string) *RootModel {
	h := help.New()
	h.ShowAll = false

	table := NewTableScreen()
	content := NewContentScreen()

	return &RootModel{
		table:   table,
		content: content,
		current: table,
		store:   s,
		spinner: theme.NewSpinner(),
		help:    h,
		keys:    theme.Keys,
		title:   title,
	}
}

func (m *RootModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.store.FetchResources(),
		store.Tick(refreshInterval),
	)
}

func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.updateSizes()
		return m, nil

	case messages.TickMsg:
		if !m.updating && m.current == m.table {
			m.updating = true
			return m, tea.Batch(m.store.FetchResources(), store.Tick(refreshInterval))
		}
		return m, store.Tick(refreshInterval)

	case messages.ResourcesLoadedMsg:
		m.table.Update(msg)
		return m, m.store.UpdateStatus(msg.Resources)

	case messages.ResourceStatusUpdatedMsg:
		m.updating = false
		m.table.Update(msg)
		return m, nil

	case messages.ContentLoadedMsg:
		m.content.Update(msg)
		return m, nil

	case messages.ShowContentMsg:
		m.content.Reset()
		m.current = m.content
		if msg.Mode == "yaml" {
			return m, m.store.FetchYAML(msg.Resource)
		}
		return m, m.store.FetchDescribe(msg.Resource)

	case messages.GoBackMsg:
		m.current = m.table
		return m, nil

	case messages.ErrMsg:
		m.err = msg.Err
		m.updating = false
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Back):
			if m.current != m.table {
				m.current = m.table
				return m, nil
			}
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keys.Refresh):
			if !m.updating {
				m.updating = true
				return m, m.store.FetchResources()
			}
			return m, nil
		}
	}

	_, cmd := m.current.Update(msg)
	return m, cmd
}

func (m *RootModel) View() string {
	if m.err != nil {
		return theme.Error(m.err.Error())
	}

	header := theme.Header(m.title, m.spinner, m.updating, m.width)
	footer := theme.Footer(m.help.View(m.keys), m.width)
	bodyHeight := max(0, m.height-lipgloss.Height(header)-lipgloss.Height(footer))

	m.table.SetSize(m.width, bodyHeight)
	m.content.SetSize(m.width, bodyHeight)

	return lipgloss.JoinVertical(lipgloss.Top, header, m.current.View(), footer)
}

func (m *RootModel) updateSizes() {
	header := theme.Header(m.title, m.spinner, m.updating, m.width)
	footer := theme.Footer(m.help.View(m.keys), m.width)
	bodyHeight := max(0, m.height-lipgloss.Height(header)-lipgloss.Height(footer))
	m.table.SetSize(m.width, bodyHeight)
	m.content.SetSize(m.width, bodyHeight)
}
