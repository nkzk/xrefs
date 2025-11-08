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

func topText(m *Model) string {
	return fmt.Sprintf("%s\n", m.config.ColCompositionRevision)
}

func (m *Model) View() string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Faint(true).PaddingLeft(2)
	tableHelp := helpStyle.Render("\nd: describe\ny/enter: yaml\narrow-up/k: up\narrow-down/j: down\nG: bottom\ng: top")

	topTextStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A1D6FE")).Faint(true).PaddingLeft(2)
	s := "\n" + topTextStyle.Render(topText(m))

	if m.err != nil {
		return "could not render view cause of error:\n" + m.err.Error()
	}

	if !m.loaded {
		return "\nloading…\n"
	}

	if len(m.rows) == 0 {
	m.loaded = false
		return "\nThis one doesnt have any resource references ¯\\_(ツ)_/¯\nTry another one\n\n(press q to go back)\n"
	}

	if m.showViewport {
		return m.viewportView()
	} else {
		s += m.table.String() + highlightText(tableHelp) + "\n"
	}
	return s
}

func (m *Model) viewportView() string {
	// m.viewport.GotoTop()
	return fmt.Sprintf("%s\n%s\n%s", m.viewportHeaderView(), m.viewport.View(), m.viewportFooterView())
}

func (m *Model) viewportHeaderView() string {
	title := titleStyle.Render("ツ")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m *Model) viewportFooterView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func highlightText(src string) string {
	var buf bytes.Buffer

	err := quick.Highlight(&buf, src, "go", "", "github")
	if err != nil {
		return src
	}

	return buf.String()
}

func highlightYAML(src string) string {
	var buf bytes.Buffer

	// styles: try "dracula", "github", "monokai", "solarized-dark", etc.
	if err := quick.Highlight(&buf, src, "yaml", "terminal", "github"); err != nil {
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
