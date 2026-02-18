package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ra/scrape_do_cli/internal/config"
	"github.com/ra/scrape_do_cli/internal/outfmt"
)

type ConfigCmd struct {
	Get   ConfigGetCmd   `cmd:"" aliases:"show" help:"Get a config value"`
	Set   ConfigSetCmd   `cmd:"" aliases:"add,update" help:"Set a config value"`
	Unset ConfigUnsetCmd `cmd:"" aliases:"rm,remove,del" help:"Unset a config value"`
	List  ConfigListCmd  `cmd:"" aliases:"ls,all" help:"List all config values"`
	Path  ConfigPathCmd  `cmd:"" aliases:"where" help:"Print config file path"`
}

type ConfigGetCmd struct {
	Key string `arg:"" help:"Config key (token, base_url, async_base_url, timeout_ms, default_output)"`
}

func (c *ConfigGetCmd) Run(ctx context.Context) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return &ExitError{Code: exitCodeConfig, Err: err}
	}
	key, err := config.ParseKey(c.Key)
	if err != nil {
		return err
	}
	value := config.GetValue(cfg, key)
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"key": c.Key, "value": value})
	}
	fmt.Fprintln(os.Stdout, value)
	return nil
}

type ConfigSetCmd struct {
	Key   string `arg:"" help:"Config key"`
	Value string `arg:"" help:"Config value"`
}

func (c *ConfigSetCmd) Run(ctx context.Context) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return &ExitError{Code: exitCodeConfig, Err: err}
	}
	key, err := config.ParseKey(c.Key)
	if err != nil {
		return err
	}
	if err := config.SetValue(&cfg, key, c.Value); err != nil {
		return err
	}
	if err := config.WriteConfig(cfg); err != nil {
		return &ExitError{Code: exitCodeConfig, Err: err}
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"key": c.Key, "value": c.Value, "saved": true})
	}
	fmt.Fprintf(os.Stdout, "Set %s = %s\n", c.Key, c.Value)
	return nil
}

type ConfigUnsetCmd struct {
	Key string `arg:"" help:"Config key"`
}

func (c *ConfigUnsetCmd) Run(ctx context.Context) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return &ExitError{Code: exitCodeConfig, Err: err}
	}
	key, err := config.ParseKey(c.Key)
	if err != nil {
		return err
	}
	if err := config.UnsetValue(&cfg, key); err != nil {
		return err
	}
	if err := config.WriteConfig(cfg); err != nil {
		return &ExitError{Code: exitCodeConfig, Err: err}
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"key": c.Key, "removed": true})
	}
	fmt.Fprintf(os.Stdout, "Unset %s\n", c.Key)
	return nil
}

type ConfigListCmd struct{}

func (c *ConfigListCmd) Run(ctx context.Context) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return &ExitError{Code: exitCodeConfig, Err: err}
	}
	cfgPath, _ := config.ConfigPath()
	payload := map[string]any{"path": cfgPath}
	for _, name := range config.KeyNames() {
		key, _ := config.ParseKey(name)
		payload[name] = config.GetValue(cfg, key)
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, payload)
	}
	fmt.Fprintf(os.Stdout, "Config file: %s\n", cfgPath)
	for _, name := range config.KeyNames() {
		fmt.Fprintf(os.Stdout, "%s\t%v\n", name, payload[name])
	}
	return nil
}

type ConfigPathCmd struct{}

func (c *ConfigPathCmd) Run(ctx context.Context) error {
	cfgPath, err := config.ConfigPath()
	if err != nil {
		return &ExitError{Code: exitCodeConfig, Err: err}
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"path": cfgPath})
	}
	fmt.Fprintln(os.Stdout, cfgPath)
	return nil
}
