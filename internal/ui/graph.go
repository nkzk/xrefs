package ui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/nkzk/xrefs/internal/models"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Graph struct {
	help help.Model
	root *models.Resource
}

var _ Tab = (*Graph)(nil)

func NewGraph(root *models.Resource, parent *Root) *Graph {
	return &Graph{
		help: help.New(),
		root: root,
	}
}

func (m Graph) Init() tea.Cmd { return nil }

func (m Graph) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case UpdateResourceMsg:
		m.root = msg.Resource
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Graph) View() tea.View {
	return tea.NewView(docStyle.Render(renderUsageGraph(m.root)))
}

func (m Graph) Title() string {
	return "Usage"
}

func (m Graph) FullScreen() bool {
	return false
}

func (m Graph) Status() string {
	return ""
}

func (m Graph) ShortHelpView() string {
	return m.help.ShortHelpView([]key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	})
}

func (m Graph) Msg() string {
	return ""
}

type usageLink struct {
	By  string
	Via string
}

func renderUsageGraph(root *models.Resource) string {
	groups := usageGroups(root)
	if len(groups) == 0 {
		return "No usage relationships found"
	}

	var targets []string
	for target := range groups {
		targets = append(targets, target)
	}
	sort.Strings(targets)

	var b strings.Builder

	for _, target := range targets {
		links := groups[target]
		sort.Slice(links, func(i, j int) bool {
			if links[i].By == links[j].By {
				return links[i].Via < links[j].Via
			}
			return links[i].By < links[j].By
		})

		fmt.Fprintf(&b, "%s\n", target)
		for i, link := range links {
			connector := "├─"
			if i == len(links)-1 {
				connector = "└─"
			}
			fmt.Fprintf(&b, "  %s %s\n", connector, link.By)
			fmt.Fprintf(&b, "     via %s\n", link.Via)
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func usageGroups(root *models.Resource) map[string][]usageLink {
	groups := map[string][]usageLink{}

	for _, r := range collectResources(root) {
		if r.Unstructured == nil || r.Unstructured.GetKind() != "Usage" {
			continue
		}

		of, ok := usageTarget(r.Unstructured, "of")
		if !ok {
			continue
		}

		by, ok := usageTarget(r.Unstructured, "by")
		if !ok {
			by = "unknown"
		}

		groups[of] = append(groups[of], usageLink{
			By:  by,
			Via: fmt.Sprintf("Usage/%s", r.Unstructured.GetName()),
		})
	}

	return groups
}

func collectResources(root *models.Resource) []*models.Resource {
	if root == nil {
		return nil
	}

	resources := []*models.Resource{root}
	for i := range root.Children {
		resources = append(resources, collectResources(&root.Children[i])...)
	}
	return resources
}

func usageTarget(u *unstructured.Unstructured, field string) (string, bool) {
	kind, _, _ := unstructured.NestedString(u.Object, "spec", field, "kind")
	name, _, _ := unstructured.NestedString(u.Object, "spec", field, "resourceRef", "name")
	namespace, _, _ := unstructured.NestedString(u.Object, "spec", field, "resourceRef", "namespace")

	if kind == "" || name == "" {
		return "", false
	}

	if namespace != "" {
		return fmt.Sprintf("%s/%s/%s", namespace, kind, name), true
	}

	return fmt.Sprintf("%s/%s", kind, name), true
}
