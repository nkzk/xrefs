package main

import (
	"fmt"
	"net/http"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

type model2 struct {
	status int
	err    error
}

var _ tea.Model = &model{}
var _ tea.Model = &model2{}

func initialModel() model {
	return model{
		choices: []string{
			"1",
			"2",
		},
		cursor:   0,
		selected: make(map[int]struct{}),
	}
}

type statusMsg int
type errMsg struct{ err error }

func checkServer() tea.Msg {
	return statusMsg(200)
}

func (e errMsg) Error() string { return e.err.Error() }

func (m model) Init() tea.Cmd {
	return nil
}

func (m model2) Init() tea.Cmd {
	return checkServer
}

func (m model2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statusMsg:
		m.status = int(msg)
		return m, tea.Quit
	case errMsg:
		m.err = msg
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model2) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
	}

	s := fmt.Sprintf("Checking %s...", "yo")

	if m.status > 0 {
		s += fmt.Sprintf("%d %s!", m.status, http.StatusText(m.status))
	}

	return "\n" + s + "\n\n"
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+C", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}

	}

	return m, nil
}

func (m model) View() string {
	s := "What should we do\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	s += "\nPress q to quit\n"

	return s
}

func Run() {
	if _, err := tea.NewProgram(model2{}).Run(); err != nil {
		fmt.Printf("failed to start TUI: %w", err)
		os.Exit(1)
	}
}
