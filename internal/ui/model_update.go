package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
	"github.com/nkzk/xrefs/internal/ui/constants"
)

const refreshInterval = 7 * time.Second

type tickMsg time.Time
type statusMsg []row
type resourceRefMsg []row

func tick() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.debugWriter != nil {
		spew.Fdump(m.debugWriter, msg)
	}

	switch msg := msg.(type) {
	case tickMsg:
		var xr string
		cmds := []tea.Cmd{tick()}

		if !m.updating {
			m.updating = true
			command, err := createGetYamlCommand(
				m.config.ResourceName,
				m.config.ResourceGroup,
				m.config.ResourceVersion,
				m.config.Name,
				m.config.Namespace)
			if err != nil {
				return m, func() tea.Msg {
					return errMsg{err: fmt.Errorf("failed to create kubectl command, %w", err)}
				}
			}

			xr, err = m.client.GetXR(command)
			if err != nil {
				return m, func() tea.Msg {
					return errMsg{err: fmt.Errorf("failed to get XR, %w", err)}
				}
			}

			cmds = append(cmds, getResourceRefs(xr, m.rowStatus))
		}

		return m, tea.Batch(cmds...)

	case resourceRefMsg:
		m.loaded = true
		_ = m.saveRowsToModel([]row(msg))

		return m, m.updateStatusCmd(m.rowStatus, []row(msg), m.client)

	case statusMsg:
		m.updating = false
		_ = m.saveRowsToModel([]row(msg))
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		m.width = msg.Width
		m.height = msg.Height

		if m.table != nil {
			m.table = m.table.Width(msg.Width).Height(msg.Height)
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keymap.Back):
			m.showViewport = false
		case key.Matches(msg, constants.Keymap.Quit):
			return m, tea.Quit
		case key.Matches(msg, constants.Keymap.Describe):
			if !m.showViewport {
				row, err := m.getSelectedRow()
				if err != nil {
					m.err = fmt.Errorf("failed to get selected row: %w", err)
				}

				command, err := createDescribeCommand(row)
				if err != nil {
					m.err = fmt.Errorf("failed to create describe command: %w", err)
				}

				result, err := m.client.Run(command)
				if err != nil {
					m.err = fmt.Errorf("failed to get resource with command '%s': %w", command, err)
				}

				m.viewport.SetContent(highlightDescribe(result))
				m.viewport.GotoTop()
				m.showViewport = true
			}
		case key.Matches(msg, constants.Keymap.Enter):
			if !m.showViewport {

				row, err := m.getSelectedRow()
				if err != nil {
					m.err = fmt.Errorf("failed to get selected row: %w", err)
				}

				command, err := createGetYamlCommand(row.Kind, "", row.ApiVersion, row.Name, row.Namespace)
				if err != nil {
					m.err = fmt.Errorf("failed to create kubectl command: %w", err)
					return m, nil
				}

				result, err := m.client.Run(command)
				if err != nil {
					m.err = fmt.Errorf("failed to get reource with command '%s': %w", command, err)
					return m, nil
				}

				m.viewport.SetContent(highlightYAML(result))
				m.viewport.GotoTop()
				m.showViewport = true
			}
		case key.Matches(msg, constants.Keymap.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, constants.Keymap.Down):
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
		case key.Matches(msg, constants.Keymap.Top):
			if m.showViewport {
				m.viewport.GotoTop()
			} else {
				m.cursor = 0
			}
		case key.Matches(msg, constants.Keymap.Bottom):
			if m.showViewport {
				m.viewport.GotoBottom()
			} else {
				m.cursor = len(m.rows) - 1
			}
		case key.Matches(msg, constants.Keymap.Help):
			m.help.ShowAll = !m.help.ShowAll
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
