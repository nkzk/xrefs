package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const refreshInterval = 7 * time.Second

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		command, err := CreateKubectlCommand(m.config.ResourceName, m.config.ResourceGroup, m.config.ResourceVersion, m.config.Name, m.config.Namespace)
		if err != nil {
			return m, func() tea.Msg {
				return errMsg{err: fmt.Errorf("failed to create kubectl command, %w", err)}
			}
		}

		xr, err := m.client.GetXR(command)
		if err != nil {
			return m, func() tea.Msg {
				return errMsg{err: fmt.Errorf("failed to get XR, %w", err)}
			}
		}

		return m, tea.Batch(
			extractResourceRefs(xr),
			tick(),
		)

	case []row:
		m.applyData(msg)

	case tea.WindowSizeMsg:
		if m.table != nil {
			m.table = m.table.Width(msg.Width).Height(msg.Height)
		}
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			m.showViewport = false
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter", "y":
			r := m.rows[m.cursor]

			row, err := toRow(r)
			if err != nil {
				m.err = fmt.Errorf("failed to convert row string to row: %w", err)
			}

			command, err := CreateKubectlCommand(row.Kind, "", row.ApiVersion, row.Name, row.Namespace)
			if err != nil {
				m.err = fmt.Errorf("failed to create kubectl command: %w", err)
				return m, nil
			}

			result, err := m.client.Get(command)
			if err != nil {
				m.err = fmt.Errorf("failed to get resource with command '%s': %w", command, err)
				return m, nil
			}

			m.viewport.SetContent(result)
			m.showViewport = true
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
		}

		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, nil
}
