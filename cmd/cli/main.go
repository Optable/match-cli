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
			Compact: false,
			// Ensure that sub-commands and their children are not shown by
			// default. This removes a lot of the noise in the top-level help
			// where the total sub-commands is quite high.
			NoExpandSubcommands: true,
			WrapUpperBound:      80,
		},
	)

	cliCtx, err := c.NewContext()
	kongCtx.FatalIfErrorf(err)

	kongCtx.FatalIfErrorf(kongCtx.Run(cliCtx))
}
