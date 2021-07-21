package cli

import (
	"context"
	"encoding/json"
	"fmt"
)

type Cli struct {
	Verbose int `opt:"" short:"v" type:"counter" help:"Enable debug mode."`

	Partner PartnerCmd `cmd:"" help:"Partner command."`
	Match   MatchCmd   `cmd:"" help:"Match command."`
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
