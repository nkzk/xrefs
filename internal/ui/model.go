package ui

import (
	"fmt"
	"os"
	"time"

	viewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"gopkg.in/yaml.v3"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func NewModel() *Model {
	r := row{}
	headers := headersFromRow(r)
	rows := [][]string{}

	re := lipgloss.NewRenderer(os.Stdout)
	baseStyle := re.NewStyle().Padding(0, 1)
	selectedStyle := baseStyle.Foreground(lipgloss.Color("#01BE85")).Background(lipgloss.Color("#00E2C7"))

	m := &Model{
		viewport: viewport.New(0, 0),
		client:   mock{},
	}

	t := table.New().
		Headers(headers...).
		Rows(rows...).
		Border(lipgloss.NormalBorder()).
		BorderStyle(re.NewStyle().Foreground(lipgloss.Color("238"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == m.cursor {
				return selectedStyle
			}
			return baseStyle
		}).
		Border(lipgloss.ThickBorder())

	m.table = t
	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		getRefs(m.client.GetXRD()),
		tick(),
	)
}

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
		return m, tea.Batch(
			getRefs(m.client.GetXRD()),
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
		case "enter":
			r := m.rows[m.cursor]
			row, err := toRow(r)
			if err != nil {
				m.err = fmt.Errorf("failed to convert row string to row: %w", err)
			}
			result := m.client.Get(row.Kind, row.ApiVersion, row.Name, row.Namespace)
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
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) View() string {
	s := "\n"

	if m.err != nil {
		return "could not render view cause of error:\n" + m.err.Error()
	}

	if m.table == nil {
		return "\nloading…\n"
	}

	if m.showViewport {
		s += m.viewport.View()
	} else {
		s += m.table.String() + "\n"
	}
	return s
}

func (m *Model) applyData(newRows []row) {
	if len(newRows) == 0 {
		return
	}

	m.table.ClearRows()
	m.rows = [][]string{}

	rows := make([][]string, 0, len(newRows))
	for _, r := range newRows {
		rows = append(rows, toStringRow(r))
	}

	m.rows = rows
	m.table.Rows(rows...)
}

func getRows(yamlString string) ([]row, error) {
	xr := &XR{}
	err := yaml.Unmarshal([]byte(yamlString), xr)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal xr: %w", err)
	}

	var result []row
	for _, resourceRef := range xr.Spec.Crossplane.ResourceRefs {
		result = append(result, row{
			Namespace:    xr.Metadata.Namespace,
			Kind:         resourceRef.Kind,
			ApiVersion:   resourceRef.ApiVersion,
			Name:         resourceRef.Name,
			Synced:       "True",
			SyncedReason: "Yes",
			Ready:        "True",
			ReadyReason:  "Yes",
		})
	}

	return result, nil
}

func getRefs(yamlString string) tea.Cmd {
	return func() tea.Msg {
		refs, err := getRows(yamlString)
		if err != nil {
			return errMsg{err: err}
		}
		return refs
	}
}
