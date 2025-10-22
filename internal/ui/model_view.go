package ui

import "fmt"

func (m *Model) View() string {
	s := "\n" + fmt.Sprintf("%s.%s.%s/%s -n %s | %s | %s\n", m.config.ResourceName, m.config.ResourceVersion, m.config.ResourceGroup, m.config.Name, m.config.Namespace, m.config.ColComposition, m.config.ColCompositionRevision)

	if m.err != nil {
		return "could not render view cause of error:\n" + m.err.Error()
	}

	if m.table == nil {
		return "\nloading…\n"
	}

	if m.showViewport {
		s += m.viewport.View()
	} else {
		s += m.table.String() + "\n"
	}
	return s
}
