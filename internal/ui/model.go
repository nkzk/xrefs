package ui

import (
	"fmt"
	"os"

	viewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/nkzk/xrefs/internal/config"
	"gopkg.in/yaml.v3"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func NewModel(client Client, config config.Config) *Model {
	r := row{}
	headers := headersFromRow(r)
	rows := [][]string{}

	re := lipgloss.NewRenderer(os.Stdout)
	baseStyle := re.NewStyle().Padding(0, 1)
	selectedStyle := baseStyle.Foreground(lipgloss.Color("#01BE85")).Background(lipgloss.Color("#00E2C7"))

	m := &Model{
		viewport: viewport.New(0, 0),
		client:   client,
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

	m.config = config
	m.table = t
	return m
}

func (m *Model) Init() tea.Cmd {
	command, err := CreateKubectlCommand(m.config.ResourceName, m.config.ResourceGroup, m.config.ResourceVersion, m.config.Name, m.config.Namespace)
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
		extractResourceRefs(xr),
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
			Synced:       "True",
			SyncedReason: "Yes",
			Ready:        "True",
			ReadyReason:  "Yes",
		})
	}

	return result, nil
}

func extractResourceRefs(yamlString string) tea.Cmd {
	return func() tea.Msg {
		refs, err := getRows(yamlString)
		if err != nil {
			return errMsg{err: err}
		}
		return refs
	}
}
