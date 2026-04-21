package install

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/nkzk/xrefs/internal/k9s"
)

type Cmd struct {
	Shortcut string
}

func (c *Cmd) Help() string {
	return `
		This command will install this cli as a k9s plugin.
		It will make sure to read existing config and only append the plugin configuration, and save a backup before making any changes.
		`
}

func (c *Cmd) Run(k *kong.Context) error {
	fmt.Fprintf(k.Stdout, "Installing k9s plugin\n")
	if err := k9s.Install(c.Shortcut); err != nil {
		fmt.Fprintf(os.Stderr, "install failed: %v\n", err)
		os.Exit(1)
	}
	return nil
}
