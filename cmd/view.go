package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"
)

type Cmd struct {
	ResourceName      string `required:"" name:"name" help:"resource metadata.name" group:"resource-flags"`
	ResourceNamespace string `required:"" name:"namespace" help:"resource metadata.namespace" group:"resource-flags"`
	ResourceGroup     string `required:"" name:"group" help:"resource API group" group:"resource-flags"`
	ResourceKind      string `required:"" name:"kind" help:"resource kind" group:"resource-flags"`
	ResourceVersion   string `required:"" name:"version" help:"resource API version" group:"resource-flags"`

	Mock bool `default:"false" help:"mock mode for development" group:"development"`
}

func (c *Cmd) Help() string {
	return `This command will display and let you navigate the sub-resources of the targeted kubernetes resource (resource-flags)`
}

func (c *Cmd) Run(k *kong.Context) error {
	_, err := tea.NewProgram(nil).Run()
	if err != nil {
		fmt.Fprintf(k.Stdout, "failed to start: %v\n", err)
		os.Exit(1)
	}

	// ui.NewRootModel(client, *config), tea.WithAltScreen()).Run(); err != nil {
	// fmt.Printf("failed to start: %v", err)
	// os.Exit(1)

	return nil
}
