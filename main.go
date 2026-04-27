package main

import (
	"os"

	"github.com/alecthomas/kong"
	view "github.com/nkzk/xrefs/cmd"
	"github.com/nkzk/xrefs/cmd/k9s"
)

var _ = kong.Must(&cli{})

type debugFlag bool

func (v debugFlag) BeforeApply(ctx *kong.Context) error {
	return nil
}

// the top-level cli
type cli struct {
	// subcommands
	ViewCmd view.Cmd `cmd:"" name:"view" help:"display subresources"`
	K9sCmd  k9s.Cmd  `cmd:"" name:"k9s" help:""`

	// flags
	Debug debugFlag `help:"Enable debug logging"`
}

func main() {
	parser := kong.Must(&cli{},
		kong.Name("xrefs"),
		kong.Description("A command-line tool for interacting with subresources of kubernetes resources like crossplane claims or flux-kustomization"),
		kong.ConfigureHelp(kong.HelpOptions{
			FlagsLast:      true,
			Compact:        true,
			WrapUpperBound: 80,
		}),
		kong.UsageOnError())

	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	err = ctx.Run()
	ctx.FatalIfErrorf(err)
}
