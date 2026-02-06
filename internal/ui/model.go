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
)

type Model struct {
	client Client
	config config.Config

	table    *table.Table
	viewport viewport.Model
	help     help.Model
	spinner  spinner.Model

	width, height int
	cursor        int

	viewportReady bool
	showViewport  bool
	loaded        bool
	updating      bool

	rows      [][]string
	rowStatus *sync.Map

	err error

	debugWriter io.Writer
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
		Border(lipgloss.RoundedBorder()).
		// BorderStyle(re.NewStyle().Foreground(lipgloss.Color("#ffffffff"))).
		BorderRow(false).
		BorderColumn(false).
		BorderTop(true).
		BorderBottom(true).
		BorderRight(false).
		BorderLeft(false).
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
		})

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
