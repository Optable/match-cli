package main

import (
	"github.com/optable/match-cli/pkg/cli"

	"github.com/alecthomas/kong"
)

const description = `
Optable Match CLI tool.
`

func main() {
	var c cli.Cli
	kongCtx := kong.Parse(
		&c,
		kong.Description(description),
		kong.UsageOnError(),
		&kong.HelpOptions{
			Compact:             false,
			NoExpandSubcommands: true,
			WrapUpperBound:      80,
		},
	)

	cliCtx, err := c.NewContext()
	kongCtx.FatalIfErrorf(err)

	kongCtx.FatalIfErrorf(kongCtx.Run(cliCtx))
}
