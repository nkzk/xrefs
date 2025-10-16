package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nkzk/xrefs/internal/config"
	"github.com/nkzk/xrefs/internal/k9s"
	"github.com/nkzk/xrefs/internal/ui"
)

func main() {
	var install bool
	var shortcut string

	flag.BoolVar(&install, "install", false, "Install the k9s plugin")
	flag.StringVar(&shortcut, "shortcut", "Shift-G", "Shortcut for the plugin (e.g. x, Shift-G, Ctrl-G)")
	flag.Parse()

	if install {
		if err := k9s.Install(shortcut); err != nil {
			fmt.Fprintf(os.Stderr, "install failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("plugin installed")
		return
	}

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
