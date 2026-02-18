package cmd

import (
	"context"
	"strings"

	"github.com/ra/scrape_do_cli/internal/client"
)

type PluginCmd struct {
	Run PluginRunCmd `cmd:"" help:"Run a plugin endpoint"`
}

type PluginRunCmd struct {
	PluginPath string   `arg:"" name:"plugin_path" help:"Plugin endpoint path (e.g. amazon/pdp)"`
	URL        string   `name:"url" help:"Target URL for plugins that require URL input"`
	Param      []string `name:"param" help:"Additional plugin query param (key=value)"`
	Header     []string `name:"header" help:"Header passed to Scrape.do request, format: 'Key: Value'"`
	SDHeader   []string `name:"sd-header" help:"Forwarded website header, format: 'Key: Value' (auto-prefixed with sd-)"`
}

func (c *PluginRunCmd) Run(ctx context.Context) error {
	if err := requireToken(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(c.PluginPath) == "" {
		return usage("missing plugin_path")
	}
	params, err := parseParams(c.Param)
	if err != nil {
		return err
	}
	headers, err := parseHeaders(c.Header)
	if err != nil {
		return err
	}
	sdHeaders, err := parseHeaders(c.SDHeader)
	if err != nil {
		return err
	}

	resp, err := newClientFromContext(ctx).PluginRun(ctx, client.PluginRequest{
		PluginPath: c.PluginPath,
		URL:        c.URL,
		Params:     params,
		Headers:    headers,
		SDHeaders:  sdHeaders,
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, map[string]any{"plugin_path": c.PluginPath, "url": c.URL})
}
