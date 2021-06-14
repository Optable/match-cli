package main

import (
	"github.com/optable/match-cli/pkg/cli"

	"github.com/alecthomas/kong"
)

const description = `
Optable Match CLI interface.
`

func main() {
	var c cli.Cli
	kongCtx := kong.Parse(&c, kong.Description(description))

	cliCtx, err := c.NewContext()
	kongCtx.FatalIfErrorf(err)

	err = kongCtx.Run(cliCtx)
	kongCtx.FatalIfErrorf(err)
}
