package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

type CliContext struct {
	ctx        context.Context
	config     Config
	configPath string
}

func (c *CliContext) LoadConfig() error {
	file, err := os.OpenFile(c.configPath, os.O_RDONLY, 0600)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to open config file %s : %w", c.configPath, err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&c.config)
	if err != nil {
		return fmt.Errorf("failed to decode config file %s : %w", c.configPath, err)
	}
	return nil
}

func (c *CliContext) SaveConfig() error {
	file, err := os.OpenFile(c.configPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to save config file %s : %w", configFile, err)
	}

	err = json.NewEncoder(file).Encode(&c.config)
	if err != nil {
		return fmt.Errorf("failed to encode config file %s : %w", configFile, err)
	}
	return nil
}
