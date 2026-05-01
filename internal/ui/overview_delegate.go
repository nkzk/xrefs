package ui

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

type ResourceListDelegate struct {
	selected lipgloss.Style
	normal   lipgloss.Style
	notFound lipgloss.Style
	err      lipgloss.Style
}

func NewResourceListDelegate() ResourceListDelegate {
	return ResourceListDelegate{
		selected: lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true),
		normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("#9b9b9b")),
		notFound: lipgloss.NewStyle().Foreground(lipgloss.Color("#ff9898")),
		err:      lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f5f")),
	}
}

func (d ResourceListDelegate) Height() int  { return 1 }
func (d ResourceListDelegate) Spacing() int { return 0 }
func (d ResourceListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d ResourceListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	resourceItem, ok := item.(ResourceItem)
	if !ok {
		return
	}

	r := resourceItem.Resource

	ready := condStatus(r, "Ready")
	synced := condStatus(r, "Synced")
	reason := shorten(condReason(r), 40)

	if r != nil && r.Error != nil {
		ready = "-"
		synced = "-"
		reason = shorten(r.Error.Error(), 40)
	}

	if r != nil && r.NotFound {
		ready = "-"
		synced = "-"
		reason = "Resource was not found"
	}

	row := fmt.Sprintf(
		"%-64s  %-15s %-13s %-14s %s",
		renderResourceName(resourceItem),
		resourceNamespace(r),
		ready,
		synced,
		reason,
	)

	isSelected := index == m.Index()

	switch {
	case r != nil && r.NotFound && isSelected:
		fmt.Fprint(w, d.notFound.Bold(true).Render(row))
	case r != nil && r.NotFound:
		fmt.Fprint(w, d.notFound.Render(row))
	case r != nil && r.Error != nil && isSelected:
		fmt.Fprint(w, d.err.Bold(true).Render(row))
	case r != nil && r.Error != nil:
		fmt.Fprint(w, d.err.Render(row))
	case isSelected:
		fmt.Fprint(w, d.selected.Render(row))
	default:
		fmt.Fprint(w, d.normal.Render(row))
	}
}

func shorten(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}

	const marker = "(...)"
	if max <= len(marker) {
		return s[:max]
	}

	head := (max - len(marker)) / 2
	tail := max - len(marker) - head
	return s[:head] + marker + s[len(s)-tail:]
}
