package ui

import (
	"fmt"

	"github.com/nkzk/xrefs/internal/models"
	"github.com/Digital-Shane/treeview"
)

type ResourceBuilderProvider struct{}

func (r *ResourceBuilderProvider) ID(d models.Resource) string {
	return d.ID
}
func (r *ResourceBuilderProvider) Name(d models.Resource) string {
	return fmt.Sprintf("%s ▸ %s", d.Name, *d.Namespace)
}

func (r *ResourceBuilderProvider) Children(d models.Resource) []models.Resource {
	return d.Children
}


func createResourceProvider() *treeview.DefaultNodeProvider[models.Resource] {
	// Icon rules based on project data kind and language
	projectIconRule := treeview.WithIconRule(projectHasKind("project"), "🏗")
	moduleIconRule := treeview.WithIconRule(projectHasKind("module"), "📂")
	packageIconRule := treeview.WithIconRule(projectHasKind("package"), "📁")
	goFileIconRule := treeview.WithIconRule(projectIsFileWithLang("go"), "🐹")
	jsFileIconRule := treeview.WithIconRule(projectIsFileWithLang("javascript"), "🟨")
	tsFileIconRule := treeview.WithIconRule(projectIsFileWithLang("typescript"), "🔷")
	pyFileIconRule := treeview.WithIconRule(projectIsFileWithLang("python"), "🐍")
	fileIconRule := treeview.WithIconRule(projectHasKind("file"), "📄")
	defaultIconRule := treeview.WithDefaultIcon[ProjectData]("❓")

	// Style rules based on project kind and importance
	projectStyleRule := treeview.WithStyleRule(
		projectHasKind("project"),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("57")).
			Bold(true).
			PaddingLeft(2).
			PaddingRight(2),
		lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			PaddingLeft(2).
			PaddingRight(2),
	)
	moduleStyleRule := treeview.WithStyleRule(
		projectHasKind("module"),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("228")).
			Background(lipgloss.Color("234")).
			PaddingLeft(1),
		lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			PaddingLeft(1),
	)
	packageStyleRule := treeview.WithStyleRule(
		projectHasKind("package"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("156")),
		lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true),
	)
	fileStyleRule := treeview.WithStyleRule(
		projectHasKind("file"),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("251")).
			Faint(true),
		lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true),
	)
	importantStyleRule := treeview.WithStyleRule(
		projectIsImportant(),
		lipgloss.NewStyle().Underline(true), // Add underline to existing style for important items
		lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			Underline(true),
	)
	defaultProjectStyleRule := treeview.WithStyleRule(
		func(n *treeview.Node[ProjectData]) bool { return true },
		lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true),
	)

	// Formatter rule
	formatterRule := treeview.WithFormatter[ProjectData](projectFormatter)

	return treeview.NewDefaultNodeProvider(
		// Icon rules (most specific first)
		goFileIconRule, jsFileIconRule, tsFileIconRule, pyFileIconRule, fileIconRule,
		projectIconRule, moduleIconRule, packageIconRule, defaultIconRule,
		// Style rules (important rule before others to ensure it applies)
		importantStyleRule, projectStyleRule, moduleStyleRule, packageStyleRule, fileStyleRule, defaultProjectStyleRule,
		// Formatter
		formatterRule,
	)