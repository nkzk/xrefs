package models

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/nkzk/xrefs/internal/ui/v2/messages"
	"github.com/nkzk/xrefs/internal/ui/v2/theme"
)

type TableScreen struct {
	resources []messages.Resource
	rows      [][]string
	cursor    int
	offset    int
	width     int
	height    int
	loaded    bool
	table     *table.Table
	keys      theme.KeyMap
}

func NewTableScreen() *TableScreen {
	m := &TableScreen{
		keys:  theme.Keys,
		table: table.New(),
	}

	m.table = table.New().
		Headers(messages.Headers()...).
		Border(lipgloss.RoundedBorder()).
		BorderRow(false).
		BorderColumn(false).
		BorderTop(true).
		BorderBottom(true).
		BorderRight(false).
		BorderLeft(false).
		StyleFunc(m.cellStyleFunc)

	return m
}

func (m *TableScreen) Init() tea.Cmd { return nil }

func (m *TableScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.ResourcesLoadedMsg:
		m.setResources(msg.Resources)
		m.loaded = true

	case messages.ResourceStatusUpdatedMsg:
		m.setResources(msg.Resources)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *TableScreen) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
			m.adjustOffset()
		}

	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.rows)-1 {
			m.cursor++
			m.adjustOffset()
		}

	case key.Matches(msg, m.keys.Top):
		m.cursor = 0
		m.adjustOffset()

	case key.Matches(msg, m.keys.Bottom):
		if len(m.rows) > 0 {
			m.cursor = len(m.rows) - 1
			m.adjustOffset()
		}

	case key.Matches(msg, m.keys.Enter):
		if r, ok := m.SelectedResource(); ok {
			return m, func() tea.Msg {
				return messages.ShowContentMsg{Resource: r, Mode: "yaml"}
			}
		}

	case key.Matches(msg, m.keys.Describe):
		if r, ok := m.SelectedResource(); ok {
			return m, func() tea.Msg {
				return messages.ShowContentMsg{Resource: r, Mode: "describe"}
			}
		}
	}

	m.table.Offset(m.offset)
	return m, nil
}

func (m *TableScreen) View() string {
	if !m.loaded {
		return "\n  Loading...\n"
	}
	if len(m.rows) == 0 {
		return "\n  No resource references found.\n"
	}
	return m.table.String()
}

func (m *TableScreen) SetSize(w, h int) {
	m.width, m.height = w, h
	m.table.Width(max(0, w-2))
	m.table.Height(max(0, h-2))
}

func (m *TableScreen) SelectedResource() (messages.Resource, bool) {
	if m.cursor < 0 || m.cursor >= len(m.resources) {
		return messages.Resource{}, false
	}
	return m.resources[m.cursor], true
}

func (m *TableScreen) setResources(resources []messages.Resource) {
	m.resources = resources
	m.rows = make([][]string, len(resources))
	for i, r := range resources {
		m.rows[i] = r.ToRow()
	}
	m.table.ClearRows()
	m.table.Rows(m.rows...)
}

func (m *TableScreen) visibleRows() int {
	return max(0, m.height-8)
}

func (m *TableScreen) adjustOffset() {
	visible := m.visibleRows()
	if visible <= 0 {
		return
	}

	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}

	maxOffset := len(m.rows) - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m *TableScreen) cellStyleFunc(row, col int) lipgloss.Style {
	isHeader := row == table.HeaderRow
	value := ""
	if row >= 0 && row < len(m.rows) && col >= 0 && col < len(m.rows[row]) {
		value = m.rows[row][col]
	}
	return theme.TableCellStyle(row, col, m.cursor, value, isHeader)
}
