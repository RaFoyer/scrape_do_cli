package cmd

import (
	"context"
	"strings"

	"github.com/ra/scrape_do_cli/internal/client"
)

type ScrapeCmd struct {
	URL         string   `arg:"" name:"url" help:"Target URL to scrape"`
	Render      bool     `name:"render" help:"Execute JavaScript rendering"`
	Super       bool     `name:"super" help:"Use super (residential/mobile) proxy"`
	Geo         string   `name:"geo" aliases:"geocode" help:"Country geocode (e.g. us)"`
	RegionalGeo string   `name:"regional-geo" help:"Regional geocode (europe, asia, ...)"`
	SessionID   string   `name:"session-id" help:"Sticky proxy session ID"`
	Device      string   `name:"device" help:"Device profile (Desktop|Mobile)"`
	APIOutput   string   `name:"output" help:"API output mode (raw|markdown)"`
	Callback    string   `name:"callback" help:"Callback URL"`
	Param       []string `name:"param" help:"Additional query param (key=value)"`
	Header      []string `name:"header" help:"Header passed to Scrape.do request, format: 'Key: Value'"`
	SDHeader    []string `name:"sd-header" help:"Forwarded website header, format: 'Key: Value' (auto-prefixed with sd-)"`
}

func (c *ScrapeCmd) Run(ctx context.Context) error {
	if err := requireToken(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(c.URL) == "" {
		return usage("missing url")
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

	resp, err := newClientFromContext(ctx).Scrape(ctx, client.ScrapeRequest{
		URL:         c.URL,
		Render:      c.Render,
		Super:       c.Super,
		Geo:         c.Geo,
		RegionalGeo: c.RegionalGeo,
		SessionID:   c.SessionID,
		Device:      c.Device,
		Output:      c.APIOutput,
		Callback:    c.Callback,
		Params:      params,
		Headers:     headers,
		SDHeaders:   sdHeaders,
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, map[string]any{"url": c.URL})
}
