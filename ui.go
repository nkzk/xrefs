package main

import (
	"fmt"
	"os"
	"reflect"
	"time"

	viewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type model struct {
	table        *table.Table
	rows         [][]string
	cursor       int
	err          error
	client       Client
	viewport     viewport.Model
	showViewport bool
}

type row struct {
	Namespace  string
	Kind       string
	ApiVersion string
	Name       string
	Synced     string
	Ready      string
}

func NewModel() *model {
	r := row{}
	headers := headersFromRow(r)
	rows := [][]string{}

	re := lipgloss.NewRenderer(os.Stdout)

	baseStyle := re.NewStyle().Padding(0, 1)

	selectedStyle := baseStyle.Foreground(lipgloss.Color("#01BE85")).Background(lipgloss.Color("#00E2C7"))

	m := &model{
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

const refreshInterval = 7 * time.Second

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		getRefs(m.client.GetXRD()),
		tick(),
	)
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
	return []string{r.Namespace, r.Kind, r.ApiVersion, r.Name, r.Synced, r.Ready}
}

func toRow(s []string) (row, error) {
	var r row
	v := reflect.ValueOf(r)

	if len(s) != v.NumField() {
		return row{}, fmt.Errorf("row has %d fields but only %d is allowed", len(s), v.NumField())
	}

	rv := reflect.ValueOf(&r).Elem()
	for i := 0; i < rv.NumField(); i++ {
		v.Type().Field(i)
		f := rv.Field(i)
		sf := v.Field(i)

		if !f.CanSet() {
			return row{}, fmt.Errorf("field %q cannot be set (make it exported)", sf.Type().Name())
		}

		if f.Kind() != reflect.String {
			return row{}, fmt.Errorf("field %q is %s, need string", sf.Type().Name, f.Kind())
		}

		f.SetString(s[i])
	}

	return r, nil
}
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

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

func (m *model) View() string {
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

func (m *model) applyData(newRows []row) {
	if len(newRows) == 0 {
		return
	}

	// just reset for now
	m.table.ClearRows()
	m.rows = [][]string{}

	rows := make([][]string, 0, len(newRows))
	for _, r := range newRows {
		rows = append(rows, toStringRow(r))
	}

	m.rows = rows
	m.table.Rows(rows...)
}
