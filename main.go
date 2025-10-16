package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nkzk/xtree/internal/config"
	"github.com/nkzk/xtree/internal/ui"
)

func main() {
	client := ui.NewKubectlClient()
	config := config.Create()
	if !config.IsValid() {
		fmt.Printf(
			"invalid configuration, make sure NAME, NAMESPACE, RESOURCE_NAME, RESOURCE_VERSION environment variables are set\n" +
				"this is usually done by k9s automatically\n" +
				"if you try to run this program outside of k9s you must supply these environment variables manually\n",
		)
	}
	if _, err := tea.NewProgram(ui.NewModel(*client, *config), tea.WithAltScreen()).Run(); err != nil {
		fmt.Printf("failed to start: %v", err)
		os.Exit(1)
	}
}
