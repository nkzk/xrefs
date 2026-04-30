package ui

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/nkzk/xrefs/internal/models"
	"go.yaml.in/yaml/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type resourceViewModel struct {
	viewport viewport.Model
	ready    bool
	width    int
	height   int

	rawYAML string
	status  string

	keyMap ResourceViewKeymap
	help   help.Model
}

type ResourceViewKeymap struct {
	Back   key.Binding
	Copy   key.Binding
	Top    key.Binding
	Bottom key.Binding
}

func newResourceViewModel() resourceViewModel {
	return resourceViewModel{
		viewport: viewport.New(),
		keyMap: ResourceViewKeymap{
			Back:   key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "back")),
			Copy:   key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy to clipboard")),
			Top:    key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "top")),
			Bottom: key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "bottom")),
		},
		help: help.New(),
	}
}

func (m resourceViewModel) ShortHelpView() string {
	return m.help.ShortHelpView([]key.Binding{
		m.keyMap.Top,
		m.keyMap.Bottom,
		m.keyMap.Back,
		m.keyMap.Copy,
	})
}

type clearStatusMsg struct{}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

func (m resourceViewModel) Init() tea.Cmd {
	return nil
}

func (m resourceViewModel) Update(msg tea.Msg) (resourceViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case clearStatusMsg:
		m.status = ""
		return m, nil

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v
		m.viewport.SetWidth(m.width)
		m.viewport.SetHeight(m.height)
		m.ready = true

	case tea.KeyPressMsg:
		switch msg.String() {
		case "g":
			m.viewport.GotoTop()
			return m, nil

		case "G":
			m.viewport.GotoBottom()
			return m, nil
		case "c":
			if m.rawYAML == "" {
				m.status = "nothing to copy"
				return m, clearStatusAfter(1500 * time.Millisecond)
			}

			m.status = "copied to clipboard"

			return m, tea.Batch(
				tea.SetClipboard(m.rawYAML),
				clearStatusAfter(1500*time.Millisecond),
			)
		}

	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m resourceViewModel) View() tea.View {
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6f6f6f")).
		Render(m.ShortHelpView())

	body := strings.Join([]string{
		m.viewport.View(),
		footer,
	}, "\n")

	return tea.NewView(docStyle.Render(body))
}
func (m *resourceViewModel) SetResource(r *models.Resource) {
	y, err := toYAML(r.Unstructured)
	if err != nil {
		m.rawYAML = ""
		m.viewport.SetContent(fmt.Sprintf("error rendering yaml: %v", err))
		return
	}

	m.rawYAML = y
	m.viewport.SetContent(highlightYAML(y))
	m.viewport.GotoTop()
}

func toYAML(u *unstructured.Unstructured) (string, error) {
	if u == nil {
		return "The resource was not found", nil
	}

	b, err := yaml.Marshal(u.Object)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func highlightYAML(s string) string {
	lexer := lexers.Get("yaml")
	if lexer == nil {
		return s
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		return s
	}

	style := styles.Get("friendly")
	if style == nil {
		style = styles.Fallback
	}

	iterator, err := lexer.Tokenise(nil, s)
	if err != nil {
		return s
	}

	var b bytes.Buffer
	if err := formatter.Format(&b, style, iterator); err != nil {
		return s
	}

	return b.String()
}
