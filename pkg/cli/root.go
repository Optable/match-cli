package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alecthomas/kong"
)

// version will be set to be the latest git tag through build flag"
var version string

type Cli struct {
	Verbose  int        `opt:"" short:"v" type:"counter" help:"Enable debug mode."`
	Insecure bool       `opt:"" hidden:"true"`
	Version  VersionCmd `cmd:"" help:"Show match-cli version."`
	Partner  PartnerCmd `cmd:"" help:"Partner command."`
	Match    MatchCmd   `cmd:"" help:"Match command."`
}

type VersionCmd struct{}

func (v *VersionCmd) Run(ctx *kong.Context) error {
	ctx.Printf(version)
	return nil
}

func (c *Cli) NewContext() (*CliContext, error) {
	cliCtx := &CliContext{
		ctx: NewLogger("match-cli", c.Verbose).WithContext(context.Background()),
	}

	var err error
	cliCtx.configPath, err = ensureConfigPath()
	if err != nil {
		return nil, err
	}

	err = cliCtx.LoadConfig()
	if err != nil {
		return nil, err
	}

	cliCtx.insecure = c.Insecure
	return cliCtx, nil
}

func printJson(v interface{}) error {
	result, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	fmt.Println(string(result))
	return nil
}
