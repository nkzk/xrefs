package ui

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

func (m *Model) View() string {
	s := "\n" + fmt.Sprintf("%s.%s.%s/%s -n %s | %s | %s\n", m.config.ResourceName, m.config.ResourceVersion, m.config.ResourceGroup, m.config.Name, m.config.Namespace, m.config.ColComposition, m.config.ColCompositionRevision)

	if m.err != nil {
		return "could not render view cause of error:\n" + m.err.Error()
	}

	if m.table == nil {
		return "\nloading…\n"
	}

	if m.showViewport {
		return m.viewportView()
	} else {
		s += m.table.String() + "\n"
	}
	return s
}

func (m *Model) viewportView() string {
	// m.viewport.GotoTop()
	return fmt.Sprintf("%s\n%s\n%s", m.viewportHeaderView(), m.viewport.View(), m.viewportFooterView())
}

func (m *Model) viewportHeaderView() string {
	title := titleStyle.Render("¯\\_(ツ)_/¯")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m *Model) viewportFooterView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func highlightYAML(src string) string {
	var buf bytes.Buffer

	// styles: try "dracula", "github", "monokai", "solarized-dark", etc.
	if err := quick.Highlight(&buf, src, "yaml", "terminal16m", "github"); err != nil {
		return src
	}

	return buf.String()
}

// highlightDescribe colors only the key in "key: value" lines,
// keeping indentation and spacing intact.
func highlightDescribe(src string) string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")) // cyan

	re := regexp.MustCompile(`^(\s*)([^:\n]+?)(:)(\s*)(.*)$`)

	var b strings.Builder
	sc := bufio.NewScanner(strings.NewReader(src))
	for sc.Scan() {
		line := sc.Text()
		if m := re.FindStringSubmatch(line); m != nil {
			indent, key, colon, spAfter, val := m[1], m[2], m[3], m[4], m[5]
			b.WriteString(indent)
			b.WriteString(keyStyle.Render(key))
			b.WriteString(colon)
			b.WriteString(spAfter)
			b.WriteString(val)
		} else {
			b.WriteString(line)
		}
		b.WriteByte('\n')
	}
	return b.String()
}
