package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type row struct {
	namespace  string
	kind       string
	apiVersion string
	name       string
	synced     string
	ready      string
}

func (r row) Print(cursor string) string {
	return fmt.Sprintf("%s %s %s.%s/%s %s %s\n",
		cursor,
		r.namespace,
		r.kind,
		r.apiVersion,
		r.name,
		r.synced,
		r.ready,
	)
}

type model struct {
	rows   []row
	cursor int
	err    error
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func getRefs(yamlString string) tea.Cmd {
	return func() tea.Msg {
		refs, err := getRows(yamlString)
		if err != nil {
			return errMsg{err: err}
		}
		return refs
	}
}

func (m model) Init() tea.Cmd {
	mock := mock{}
	return getRefs(mock.GetXRD())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []row:
		m.rows = []row(msg)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+C", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
		}
	case errMsg:
		m.err = msg
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	s := "Resource refs:\n"

	for i, row := range m.rows {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += row.Print(cursor)
	}

	s += "\nPress q to quit\n"

	return s
}
