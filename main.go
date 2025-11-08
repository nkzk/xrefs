package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nkzk/xrefs/internal/config"
	"github.com/nkzk/xrefs/internal/k9s"
	"github.com/nkzk/xrefs/internal/ui"
	"github.com/nkzk/xrefs/internal/utils"
)

func main() {
	var install bool
	var shortcut string
	config := &config.Config{}

	flag.BoolVar(&install, "install", false, "Install the k9s plugin")
	flag.StringVar(&shortcut, "shortcut", "Shift-G", "Shortcut for the plugin (e.g. x, Shift-G, Ctrl-G)")
	flag.BoolVar(&config.Mock, "mock", false, "Mock mode for development")

	flag.StringVar(&config.Name, "name", "unknown", "selected resource metadata.name")
	flag.StringVar(&config.Namespace, "namespace", "unknown", "selected resource namespace")
	flag.StringVar(&config.ResourceGroup, "resourceGroup", "unknown", "selected resource API Group")
	flag.StringVar(&config.ResourceName, "resourceName", "unknown", "selected resource kind")
	flag.StringVar(&config.ResourceVersion, "resourceVersion", "unknown", "selected resource apiVersion")
	flag.StringVar(&config.ColComposition, "colComposition", "unknown", "selected column composition")
	flag.StringVar(&config.ColCompositionRevision, "colCompositionRevision", "unknown", "selected column composition revision")

	flag.BoolVar(&config.Debug, "debug", false, "debug mode")
	flag.StringVar(&config.DebugPath, "debugPath", "/tmp/xrefs/debug.log", "path to debug log file")

	flag.Parse()

	if install {
		if err := k9s.Install(shortcut); err != nil {
			fmt.Fprintf(os.Stderr, "install failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("plugin installed")
		return
	}

	var client ui.Client
	if config.Mock {
		client = ui.NewMockClient()
	} else {
		client = ui.NewKubectlClient()
	}

	if config.Debug {
		file, err := utils.OpenFile(config.DebugPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open %s: %w", config.DebugPath, err)
		}

		config.DebugWriter = file
	}

	if !config.IsValid() {
		fmt.Printf(
			"invalid configuration, make sure NAME, NAMESPACE, RESOURCE_NAME, RESOURCE_VERSION is set\n" +
				"this is usually done by k9s automatically\n" +
				"if you try to run this program outside of k9s you must supply these environment variables manually\n",
		)
	}

	if _, err := tea.NewProgram(ui.NewModel(client, *config), tea.WithAltScreen()).Run(); err != nil {
		fmt.Printf("failed to start: %v", err)
		os.Exit(1)
	}
}
