package ui

import (
	"fmt"
	"time"

	viewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
		command, err := createGetYamlCommand(m.config.ResourceName, m.config.ResourceGroup, m.config.ResourceVersion, m.config.Name, m.config.Namespace)
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

		headerHeight := lipgloss.Height(m.viewportHeaderView())
		footerHeight := lipgloss.Height(m.viewportFooterView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.viewportReady {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
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

				result, err := m.client.Get(command)
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

				result, err := m.client.Get(command)
				if err != nil {
					m.err = fmt.Errorf("failed to get resource with command '%s': %w", command, err)
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
