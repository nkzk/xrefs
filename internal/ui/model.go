package ui

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	viewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/nkzk/xrefs/internal/config"
	"gopkg.in/yaml.v3"
)

type Model struct {
	width, height int
	config        config.Config
	table         *table.Table
	loaded        bool
	rows          [][]string
	rowStatus     *sync.Map
	updating      bool
	cursor        int
	err           error
	client        Client
	viewport      viewport.Model
	viewportReady bool
	showViewport  bool

	debugWriter io.Writer

	help help.Model

	spinner spinner.Model
}

type row struct {
	Namespace    string
	Kind         string
	ApiVersion   string
	Name         string
	Synced       string
	SyncedReason string
	Ready        string
	ReadyReason  string
}

func NewModel(client Client, config config.Config) *Model {
	r := row{}
	headers := headersFromRow(r)
	rows := [][]string{}

	re := lipgloss.NewRenderer(os.Stdout)

	// colors
	red := lipgloss.Color("#FF5555")
	green := lipgloss.Color("#00FF7F")
	selHue := lipgloss.Color("#A1D6FE") // used as default unselected text AND selected bg
	baseStyle := re.NewStyle().
		Padding(0, 1).
		Bold(true).
		Foreground(selHue)
	m := &Model{
		viewport: viewport.New(0, 0),
		client:   client,
		rows:     rows,
	}

	headerStyle := baseStyle.Foreground(lipgloss.Color("#ffffff")).Bold(true)

	t := table.New().
		Headers(headers...).
		Rows(rows...).
		Border(lipgloss.NormalBorder()).
		BorderStyle(re.NewStyle().Foreground(lipgloss.Color("#2f2f2fff"))).
		BorderRow(false).
		BorderColumn(false).
		StyleFunc(func(r, c int) lipgloss.Style {
			if r == table.HeaderRow {
				return headerStyle
			}
			s := baseStyle

			if r >= 0 && r < len(m.rows) && c >= 0 && c < len(m.rows[r]) {
				switch m.rows[r][c] {
				case "False", "no":
					s = s.Foreground(red).Bold(true)
				case "True", "yes":
					s = s.Foreground(green).Bold(true)
				}
			}

			if r == m.cursor {
				s = s.Background(selHue).Foreground(lipgloss.Color("#000000")).Bold(true)
			}

			return s
		}).
		Border(lipgloss.HiddenBorder())

	m.config = config
	m.table = t
	m.rowStatus = &sync.Map{}
	m.debugWriter = config.DebugWriter

	s := spinner.New()
	s.Spinner = spinner.Jump
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m.spinner = s

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			command, err := createGetYamlCommand(
				m.config.ResourceName,
				m.config.ResourceGroup,
				m.config.ResourceVersion,
				m.config.Name,
				m.config.Namespace)
			if err != nil {
				return func() tea.Msg {
					return errMsg{
						err: fmt.Errorf("failed to get generate kubectl command %s, %w",
							command, err),
					}
				}
			}

			xr, err := m.client.GetXR(command)
			if err != nil {
				return func() tea.Msg {
					return errMsg{
						err: fmt.Errorf("failed to get XR %s, %w",
							command, err),
					}
				}
			}

			resourceRows, err := getRows(xr, m.rowStatus)
			if err != nil {
				return func() tea.Msg {
					return errMsg{err: err}
				}
			}

			return resourceRefMsg(resourceRows)

		},
		tick())
}

func (m *Model) saveRowsToModel(rows []row) tea.Cmd {
	if len(rows) == 0 {
		m.loaded = true
		return nil
	}

	newRows := make([][]string, 0, len(rows))
	for _, r := range rows {
		ready, readyReason := "-", "-"
		synced, syncedReason := "-", "-"

		if s, ok := m.rowStatus.Load(r.Name); ok {
			if rs, ok := s.(status); ok {
				ready = rs.Conditions.Get("Ready").Status
				readyReason = rs.Conditions.Get("Ready").Reason
				synced = rs.Conditions.Get("Synced").Status
				syncedReason = rs.Conditions.Get("Synced").Reason
			}
		}

		r.Ready, r.ReadyReason = ready, readyReason
		r.Synced, r.SyncedReason = synced, syncedReason

		newRows = append(newRows, toStringRow(r))
	}

	m.rows = [][]string{}
	m.table.ClearRows()
	m.rows = newRows
	m.table.Rows(newRows...)

	m.updating = false
	m.loaded = true

	return func() tea.Msg {
		return resourceRefMsg(rows)
	}
}

func (m *Model) getSelectedRow() (row, error) {
	r := m.rows[m.cursor]

	return toRow(r)
}

func getRows(yamlString string, rs *sync.Map) ([]row, error) {
	xr := &XR{}
	err := yaml.Unmarshal([]byte(yamlString), xr)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal xr: %w", err)
	}

	var result []row
	for _, resourceRef := range xr.Spec.Crossplane.ResourceRefs {
		s, exists := rs.Load(resourceRef.Name)
		if !exists {
			s = status{}
		}

		s, ok := s.(status)
		if !ok {
			s = status{}
		}

		result = append(result, row{
			Namespace:    xr.Metadata.Namespace,
			Kind:         resourceRef.Kind,
			ApiVersion:   resourceRef.ApiVersion,
			Name:         resourceRef.Name,
			Synced:       s.(status).Conditions.Get("Synced").Status,
			SyncedReason: s.(status).Conditions.Get("Synced").Reason,
			Ready:        s.(status).Conditions.Get("Ready").Status,
			ReadyReason:  s.(status).Conditions.Get("Ready").Reason,
		})
	}

	return result, nil
}

func (m *Model) updateStatusCmd(rs *sync.Map, rows []row, client Client) tea.Cmd {
	m.updating = true
	return tea.Batch(func() tea.Msg {
		var wg sync.WaitGroup
		for _, r := range rows {
			wg.Add(1)
			go func(row row) {
				defer wg.Done()
				_ = client.UpdateRowStatus(rs, row)
			}(r)
		}

		wg.Wait()
		return statusMsg(rows)
	}, m.spinner.Tick)
}

func getResourceRefs(yamlString string, rs *sync.Map) tea.Cmd {
	return func() tea.Msg {
		resourceRows, err := getRows(yamlString, rs)
		if err != nil {
			return errMsg{err: err}
		}

		return resourceRefMsg(resourceRows)
	}
}

func getStatus(cmdResult string) (*status, error) {
	type s struct {
		Status status `json:"status" yaml:"status"`
	}

	result := s{}

	err := yaml.Unmarshal([]byte(cmdResult), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal command result: %w", err)
	}

	if result.Status.Conditions == nil {
		result.Status.Conditions = []condition{}
	}

	return &result.Status, nil
}
