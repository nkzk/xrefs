package models

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/nkzk/xrefs/internal/ui/v2/messages"
	"github.com/nkzk/xrefs/internal/ui/v2/theme"
)

type ContentScreen struct {
	viewport viewport.Model
	keys     theme.KeyMap
	content  string
	loaded   bool
	width    int
	height   int
}

func NewContentScreen() *ContentScreen {
	return &ContentScreen{
		viewport: viewport.New(0, 0),
		keys:     theme.Keys,
	}
}

func (m *ContentScreen) Init() tea.Cmd { return nil }

func (m *ContentScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case messages.ContentLoadedMsg:
		m.content = msg.Content
		m.loaded = true
		m.viewport.SetContent(msg.Content)
		m.viewport.GotoTop()
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Back) {
			return m, func() tea.Msg { return messages.GoBackMsg{} }
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *ContentScreen) View() string {
	if !m.loaded {
		return "\n  Loading...\n"
	}
	return m.viewport.View()
}

func (m *ContentScreen) SetSize(w, h int) {
	m.width, m.height = w, h
	m.viewport.Width = max(0, w-2)
	m.viewport.Height = max(0, h-2)
}

func (m *ContentScreen) Reset() {
	m.content = ""
	m.loaded = false
	m.viewport.SetContent("")
	m.viewport.GotoTop()
}
