package ui

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/nkzk/xrefs/internal/config"
	"github.com/nkzk/xrefs/internal/ui/constants"
	"gopkg.in/yaml.v3"
)

type ResourcesTable struct {
	cfg    config.Config
	client Client

	debug io.Writer
	err   error

	table     *table.Table
	rows      [][]string
	rowStatus *sync.Map

	cursor, width, height int

	loaded, updating bool
}

func NewResourcesTableModel(client Client, cfg config.Config) *ResourcesTable {
	m := &ResourcesTable{
		cfg:       cfg,
		client:    client,
		debug:     cfg.DebugWriter,
		rowStatus: &sync.Map{},
		table:     table.New(),
	}

	headers := headersFromRow(row{})
	m.table = table.New().
		Headers(headers...).
		Rows(m.rows...).
		Border(lipgloss.RoundedBorder()).
		BorderRow(false).
		BorderColumn(false).
		BorderTop(true).
		BorderBottom(true).
		BorderRight(false).
		BorderLeft(false).
		StyleFunc(func(r, c int) lipgloss.Style {
			return lipgloss.NewStyle()
		})

	return m
}

func (m *ResourcesTable) Init() tea.Cmd {
	return tea.Batch(
		m.fetchXR(),
		m.tick(),
	)
}

func (m *ResourcesTable) View() string {
	if !m.loaded {
		return "\nloading…\n"
	}
	if len(m.rows) == 0 {
		return "\nNo resource references.\n"
	}
	return m.table.String()
}

func (m *ResourcesTable) SetSize(w, h int) {
	m.width, m.height = w, h
	m.table.Width(max(0, w-2))
	m.table.Height(max(0, h-2))
}

func (m *ResourcesTable) ID() string       { return "table" }
func (m *ResourcesTable) IsUpdating() bool { return m.updating }

func (m *ResourcesTable) tick() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m *ResourcesTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if m.updating {
			return m, m.tick() // schedule next tick, skip fetch
		}
		return m, tea.Batch(m.fetchXR(), m.tick())

	case resourceRefMsg:
		m.loaded = true
		m.saveRowsToModel([]row(msg))
		return m, m.updateStatus([]row(msg))

	case statusMsg:
		m.updating = false
		m.saveRowsToModel([]row(msg))
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keymap.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, constants.Keymap.Down):
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
		case key.Matches(msg, constants.Keymap.Top):
			m.cursor = 0
		case key.Matches(msg, constants.Keymap.Bottom):
			if len(m.rows) > 0 {
				m.cursor = len(m.rows) - 1
			}
		case key.Matches(msg, constants.Keymap.Enter):
			// todo
			if _, err := m.selectedRow(); err == nil {
				return m, nil
			}
		case key.Matches(msg, constants.Keymap.Describe):
			// todo
			if _, err := m.selectedRow(); err == nil {
				return m, nil
			}
		}
	}
	return m, nil
}

func (m *ResourcesTable) selectedRow() (row, error) {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return row{}, fmt.Errorf("cursor out of range")
	}
	return toRow(m.rows[m.cursor])
}

func (m *ResourcesTable) saveRowsToModel(rows []row) tea.Cmd {
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

func (m *ResourcesTable) getSelectedRow() (row, error) {
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

func (m *ResourcesTable) updateStatusCmd(rs *sync.Map, rows []row, client Client) tea.Cmd {
	m.updating = true
	return tea.Cmd(func() tea.Msg {
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
	},
	)
}

func (m *ResourcesTable) fetchXR() tea.Cmd {
	return func() tea.Msg {
		cmd, err := createGetYamlCommand(
			m.cfg.ResourceName, m.cfg.ResourceGroup, m.cfg.ResourceVersion, m.cfg.Name, m.cfg.Namespace)
		if err != nil {
			return errMsg{err}
		}
		xr, err := m.client.GetXR(cmd)
		if err != nil {
			return errMsg{err}
		}
		refs, err := getRows(xr, m.rowStatus)
		if err != nil {
			return errMsg{err}
		}
		m.updating = true
		return resourceRefMsg(refs)
	}
}

func (m *ResourcesTable) updateStatus(rows []row) tea.Cmd {
	m.updating = true
	return func() tea.Msg {
		var wg sync.WaitGroup
		for _, r := range rows {
			wg.Add(1)
			go func(r row) {
				defer wg.Done()
				_ = m.client.UpdateRowStatus(m.rowStatus, r)
			}(r)
		}
		wg.Wait()
		return statusMsg(rows)
	}
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
