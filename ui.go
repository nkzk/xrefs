package main

import (
	"os"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type row struct {
	namespace  string
	kind       string
	apiVersion string
	name       string
	synced     string
	ready      string
}

type model struct {
	table  *table.Table
	rows   [][]string
	cursor int
	err    error
}

func newModel() *model {
	r := row{}
	headers := headersFromRow(r)
	rows := [][]string{}

	re := lipgloss.NewRenderer(os.Stdout)

	// base
	baseStyle := re.NewStyle().Padding(0, 1)

	// selected
	selectedStyle := baseStyle.Foreground(lipgloss.Color("#01BE85")).Background(lipgloss.Color("#00E2C7"))

	// header
	// headerStyle := baseStyle.Foreground(lipgloss.Color("252")).Bold(true)

	// colors
	// colors := map[bool]lipgloss.Color{
	// 	true:  lipgloss.Color("#FDFF90"),
	// 	false: lipgloss.Color("#75FBAB"),
	// }

	m := &model{}
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

func (m *model) Init() tea.Cmd {
	mock := mock{}
	return getRefs(mock.GetXRD())
}

func headersFromRow(r row) []string {
	v := reflect.ValueOf(r)
	t := v.Type()
	out := make([]string, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		out[i] = t.Field(i).Name
	}
	return out
}

func toStringRow(r row) []string {
	return []string{r.namespace, r.kind, r.apiVersion, r.name, r.synced, r.ready}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []row:
		if len(msg) == 0 {
			return m, nil
		}
		rows := make([][]string, 0, len(msg))
		for _, r := range msg {
			rows = append(rows, toStringRow(r))
		}

		m.rows = rows
		m.table.Rows(rows...)
	case tea.WindowSizeMsg:
		if m.table != nil {
			m.table = m.table.Width(msg.Width).Height(msg.Height)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			// do something
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

func (m *model) View() string {
	if m.table == nil {
		return "\nloading…\n"
	}
	return "\n" + m.table.String() + "\n"
}
