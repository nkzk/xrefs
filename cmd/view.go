package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"
	"github.com/nkzk/xrefs/internal/models"
)

type Cmd struct {
	ResourceName      string `required:"" name:"name" help:"resource metadata.name" group:"resource-flags" xor:"resource-flags,development"`
	ResourceNamespace string `required:"" name:"namespace" help:"resource metadata.namespace" group:"resource-flags" xor:"resource-flags,development"`
	ResourceGroup     string `required:"" name:"group" help:"resource API group" group:"resource-flags" xor:"resource-flags,development"`
	ResourceKind      string `required:"" name:"kind" help:"resource kind" group:"resource-flags" xor:"resource-flags,development"`
	ResourceVersion   string `required:"" name:"version" help:"resource API version" group:"resource-flags" xor:"resource-flags,development"`

	Mock bool `default:"false" help:"mock mode for development" group:"development" xor:"resource-flags,development"`
}

func (c *Cmd) Help() string {
	return `This command will display and let you navigate the sub-resources of the targeted kubernetes resource`
}

func (c *Cmd) Run(k *kong.Context) error {
	m := models.NewModel()
	_, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Fprintf(k.Stdout, "failed to start: %v\n", err)
		os.Exit(1)
	}

	// ui.NewRootModel(client, *config), tea.WithAltScreen()).Run(); err != nil {
	// fmt.Printf("failed to start: %v", err)
	// os.Exit(1)

	return nil
}
