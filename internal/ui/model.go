package ui

import (
	"fmt"
	"os"
	"sync"

	viewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/nkzk/xrefs/internal/config"
	"gopkg.in/yaml.v3"
)

type Model struct {
	config        config.Config
	table         *table.Table
	rows          [][]string
	cursor        int
	err           error
	client        Client
	viewport      viewport.Model
	viewportReady bool
	showViewport  bool
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
	return m
}

func (m *Model) Init() tea.Cmd {
	command, err := createGetYamlCommand(m.config.ResourceName, m.config.ResourceGroup, m.config.ResourceVersion, m.config.Name, m.config.Namespace)
	if err != nil {
		return func() tea.Msg {
			return errMsg{err: fmt.Errorf("failed to get generate kubectl command %s, %w", command, err)}
		}
	}

	xr, err := m.client.GetXR(command)
	if err != nil {
		return func() tea.Msg {
			return errMsg{err: fmt.Errorf("failed to get XR %s, %w", command, err)}
		}
	}
	return tea.Batch(
		extractResourceRefs(xr, m.client),
		tick(),
	)
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

func (m *Model) getSelectedRow() (row, error) {
	r := m.rows[m.cursor]

	return toRow(r)
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
			Synced:       "",
			SyncedReason: "",
			Ready:        "",
			ReadyReason:  "",
		})
	}

	return result, nil
}

func extractResourceRefs(yamlString string, client Client) tea.Cmd {
	return func() tea.Msg {
		resourceRows, err := getRows(yamlString)
		if err != nil {
			return errMsg{err: err}
		}

		var wg sync.WaitGroup
		for i, row := range resourceRows {
			wg.Add(1)
			go func() {
				defer wg.Done()

				resourceRows[i], _ = client.UpdateRowStatus(row)
			}()
		}

		wg.Wait()

		return resourceRows
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
