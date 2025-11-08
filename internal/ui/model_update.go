package ui

import (
	"fmt"
	"time"

	viewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
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
		var cmds []tea.Cmd
		cmds = append(cmds, tick())

		if !m.updating {
			m.updating = true
			command, err := createGetYamlCommand(m.config.ResourceName, m.config.ResourceGroup, m.config.ResourceVersion, m.config.Name, m.config.Namespace)
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
		var cmds []tea.Cmd
		_ = m.saveRowsToModel([]row(msg))

		if !m.updating {
			m.updating = true
			cmds = append(cmds, m.updateStatusCmd(m.rowStatus, []row(msg), m.client))
		}
		return m, tea.Batch(cmds...)

	case statusMsg:
		return m, m.saveRowsToModel([]row(msg))
	case tea.WindowSizeMsg:
		if m.table != nil {
			m.table = m.table.Width(msg.Width).Height(msg.Height)
		}

		headerHeight := lipgloss.Height(m.viewportHeaderView())
		footerHeight := lipgloss.Height(m.viewportFooterView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.viewportReady {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewportReady = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			m.showViewport = false
		case "q", "ctrl+c":
			return m, tea.Quit
		case "d":
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
		case "enter", "y":
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
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
		case "g":
			if m.showViewport {
				m.viewport.GotoTop()
			} else {
				m.cursor = 0
			}
		case "G":
			if m.showViewport {
				m.viewport.GotoBottom()
			} else {
				m.cursor = len(m.rows) - 1
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
