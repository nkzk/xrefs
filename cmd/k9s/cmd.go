package k9s

import (
	"github.com/alecthomas/kong"
	"github.com/nkzk/xrefs/cmd/k9s/install"
)

type Cmd struct {
	Install install.Cmd `cmd:"" help:"install"`
}

func (c *Cmd) Help() string {
	return `Subcommands for configuring this cli as a k9s plugin`
}

func (c *Cmd) Run(k *kong.Context) error {

	return nil
}
