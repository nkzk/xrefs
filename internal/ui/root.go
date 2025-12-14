package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nkzk/xrefs/internal/config"
	"github.com/nkzk/xrefs/internal/ui/constants"
	"github.com/nkzk/xrefs/internal/utils"
)

type Child interface {
	ID() string
	Init() tea.Cmd
	Update(tea.Msg) (Child, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type Root struct {
	cfg      config.Config
	client   Client
	help     help.Model
	spinner  spinner.Model
	width    int
	height   int
	children map[string]Child
	stack    utils.Stack[string]
	err      error
}

func NewRootModel(client Client, cfg config.Config) *Root {
	table := NewTableModel()

	children := map[string]Child{
		"table": table,
	}

	h := help.New()
	h.ShowAll = false

	s := spinner.New()
	s.Spinner = spinner.Jump
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	var stack utils.Stack[string]
	stack.Push("table")

	return &Root{
		cfg:      cfg,
		client:   client,
		help:     h,
		spinner:  s,
		children: children,
		stack:    stack,
	}
}

func (r Root) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	for _, child := range r.children {
		cmds = append(cmds, child.Init())
	}

	return tea.Batch(
		cmds...,
	)
}

func (r *Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height

		for _, child := range r.children {
			child.SetSize(msg.Width, msg.Height)
		}

		return r, nil

	case errMsg:
		r.err = msg
		return r, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keymap.Quit):
			return r, tea.Quit
		case key.Matches(msg, constants.Keymap.Back):
			if r.stack.Size() > 1 {
				r.stack.Pop()
				return r, nil
			}
			return r, tea.Quit
		case key.Matches(msg, constants.Keymap.Help):
			r.help.ShowAll = !r.help.ShowAll
			return r, nil
		}

	}

	cur := r.current()
	if cur == nil {
		return r, nil
	}

	updated, cmd := cur.Update(msg)
	if updated != cur {
		r.children[updated.ID()] = updated
	}

	return r, cmd
}

func (r *Root) View() string {
	if r.err != nil {
		return constants.ErrStyle(fmt.Sprintf("error: %v", r.err))
	}

	hdr := constants.Header(r.cfg.ColCompositionRevision, r.spinner, r.isUpdating(), r.width)
	ftr := constants.Footer(r.help.View(constants.Keymap), r.width)

	bodyH := max(0, r.height-lipgloss.Height(hdr)-lipgloss.Height(ftr))

	cur := r.current()
	body := ""
	if cur != nil {
		cur.SetSize(r.width, bodyH)
		body = cur.View()
	}

	return lipgloss.JoinVertical(lipgloss.Top, hdr, body, ftr)
}

func (r *Root) current() Child {
	id, ok := r.stack.Peek()
	if !ok {
		return nil
	}

	return r.children[id]
}

type updatingChild interface{ IsUpdating() bool }

func (r *Root) isUpdating() bool {
	if u, ok := r.current().(updatingChild); ok {
		return u.IsUpdating()
	}

	return false
}
