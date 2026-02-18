package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ra/scrape_do_cli/internal/client"
)

type ScrapeCmd struct {
	URL                  string   `arg:"" name:"url" help:"Target URL to scrape"`
	Method               string   `name:"method" help:"HTTP method for target request" default:"GET" enum:"GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS"`
	Body                 string   `name:"body" help:"Raw request body for non-GET methods"`
	BodyFile             string   `name:"body-file" help:"Read request body from file for non-GET methods"`
	ContentType          string   `name:"content-type" help:"Content-Type for request body"`
	Render               bool     `name:"render" help:"Execute JavaScript rendering"`
	WaitUntil            string   `name:"wait-until" help:"Render waitUntil mode (load, domcontentloaded, networkidle0, networkidle2)"`
	CustomWait           int      `name:"custom-wait" help:"Render custom wait in milliseconds"`
	WaitSelector         string   `name:"wait-selector" help:"Render wait for CSS selector"`
	Width                int      `name:"width" help:"Render viewport width in pixels"`
	Height               int      `name:"height" help:"Render viewport height in pixels"`
	BlockResources       bool     `name:"block-resources" help:"Block non-essential resources during render"`
	Screenshot           bool     `name:"screenshot" help:"Capture viewport screenshot"`
	FullScreenshot       bool     `name:"full-screenshot" help:"Capture full page screenshot"`
	ParticularScreenshot string   `name:"particular-screenshot" help:"Capture screenshot for a CSS selector"`
	PlayWithBrowser      string   `name:"play-with-browser" help:"PlayWithBrowser actions as JSON array/object"`
	PlayWithBrowserFile  string   `name:"play-with-browser-file" help:"Read PlayWithBrowser JSON from file"`
	ReturnJSON           bool     `name:"return-json" help:"Return JSON render payload (actions, screenshots, etc.)"`
	ShowWebsocket        bool     `name:"show-websocket-requests" help:"Include websocket request details in render output"`
	ShowFrames           bool     `name:"show-frames" help:"Include frame details in render output"`
	Super                bool     `name:"super" help:"Use super (residential/mobile) proxy"`
	Geo                  string   `name:"geo" aliases:"geocode" help:"Country geocode (e.g. us)"`
	RegionalGeo          string   `name:"regional-geo" help:"Regional geocode (europe, asia, ...)"`
	SessionID            string   `name:"session-id" help:"Sticky proxy session ID"`
	Device               string   `name:"device" help:"Device profile (Desktop|Mobile)"`
	APIOutput            string   `name:"output" help:"API output mode (raw|markdown)"`
	Callback             string   `name:"callback" help:"Callback URL"`
	DisableRedirection   bool     `name:"disable-redirection" help:"Disable following redirects"`
	DisableRetry         bool     `name:"disable-retry" help:"Disable automatic retries"`
	TransparentResponse  bool     `name:"transparent-response" help:"Return full target response without status validation"`
	PureCookies          bool     `name:"pure-cookies" help:"Return only target cookies in response body"`
	RequestTimeoutMS     int      `name:"request-timeout-ms" help:"Scrape.do target timeout in milliseconds"`
	RetryTimeoutMS       int      `name:"retry-timeout-ms" help:"Retry timeout in milliseconds"`
	CustomHeaders        bool     `name:"custom-headers" help:"Treat --header entries as complete custom target headers"`
	ForwardHeaders       bool     `name:"forward-headers" help:"Forward --header entries directly to target"`
	Param                []string `name:"param" help:"Additional query param (key=value)"`
	Header               []string `name:"header" help:"Header for target request, format: 'Key: Value'"`
	SDHeader             []string `name:"sd-header" help:"Extra header for Scrape.do (format: 'Key: Value', auto-prefixed with sd-)"`
	SetCookie            []string `name:"set-cookie" help:"Cookie for target request (key=value); repeatable"`
}

func (c *ScrapeCmd) Run(ctx context.Context) error {
	if err := requireToken(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(c.URL) == "" {
		return usage("missing url")
	}
	body, err := readBodyInput(c.Body, c.BodyFile)
	if err != nil {
		return err
	}
	method := strings.ToUpper(strings.TrimSpace(c.Method))
	if method == "GET" && strings.TrimSpace(body) != "" {
		return usage("GET method does not support body; use --method POST/PUT/PATCH/DELETE")
	}
	playWithBrowser, err := readJSONInput(c.PlayWithBrowser, c.PlayWithBrowserFile, "--play-with-browser", "--play-with-browser-file")
	if err != nil {
		return err
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
	cookies, err := parseCookies(c.SetCookie)
	if err != nil {
		return err
	}

	if len(cookies) > 0 {
		if len(headers) > 0 || c.CustomHeaders || c.ForwardHeaders || len(sdHeaders) > 0 {
			return usage("--set-cookie cannot be combined with --header/--custom-headers/--forward-headers/--sd-header")
		}
	}

	if c.RequestTimeoutMS < 0 || c.RetryTimeoutMS < 0 || c.CustomWait < 0 {
		return usage("timeout values must be non-negative")
	}
	if c.Width < 0 || c.Height < 0 {
		return usage("--width and --height must be non-negative")
	}
	contentType := strings.TrimSpace(c.ContentType)
	if strings.TrimSpace(body) != "" && contentType == "" {
		contentType = "application/json"
	}

	customHeaders := c.CustomHeaders
	forwardHeaders := c.ForwardHeaders
	if len(headers) > 0 && !customHeaders && !forwardHeaders {
		customHeaders = true
	}

	resp, err := newClientFromContext(ctx).Scrape(ctx, client.ScrapeRequest{
		Method:                method,
		URL:                   c.URL,
		Body:                  body,
		ContentType:           contentType,
		Render:                c.Render,
		WaitUntil:             c.WaitUntil,
		CustomWait:            c.CustomWait,
		WaitSelector:          c.WaitSelector,
		Width:                 c.Width,
		Height:                c.Height,
		BlockResources:        c.BlockResources,
		Screenshot:            c.Screenshot,
		FullScreenshot:        c.FullScreenshot,
		ParticularScreenshot:  c.ParticularScreenshot,
		PlayWithBrowser:       playWithBrowser,
		ReturnJSON:            c.ReturnJSON,
		ShowWebsocketRequests: c.ShowWebsocket,
		ShowFrames:            c.ShowFrames,
		Super:                 c.Super,
		Geo:                   c.Geo,
		RegionalGeo:           c.RegionalGeo,
		SessionID:             c.SessionID,
		Device:                c.Device,
		Output:                c.APIOutput,
		Callback:              c.Callback,
		DisableRedirection:    c.DisableRedirection,
		DisableRetry:          c.DisableRetry,
		TransparentResponse:   c.TransparentResponse,
		PureCookies:           c.PureCookies,
		RequestTimeoutMS:      c.RequestTimeoutMS,
		RetryTimeoutMS:        c.RetryTimeoutMS,
		CustomHeaders:         customHeaders,
		ForwardHeaders:        forwardHeaders,
		SetCookies:            cookies,
		Params:                params,
		Headers:               headers,
		SDHeaders:             sdHeaders,
	})
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, map[string]any{
		"url":    c.URL,
		"method": method,
	})
}

func readBodyInput(body string, bodyFile string) (string, error) {
	body = strings.TrimSpace(body)
	bodyFile = strings.TrimSpace(bodyFile)
	if body != "" && bodyFile != "" {
		return "", usage("use either --body or --body-file, not both")
	}
	if bodyFile == "" {
		return body, nil
	}
	b, err := os.ReadFile(bodyFile)
	if err != nil {
		return "", fmt.Errorf("read --body-file: %w", err)
	}
	return string(b), nil
}

func readJSONInput(value string, file string, valueFlag string, fileFlag string) (string, error) {
	value = strings.TrimSpace(value)
	file = strings.TrimSpace(file)
	if value != "" && file != "" {
		return "", usagef("use either %s or %s, not both", valueFlag, fileFlag)
	}
	if file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", fileFlag, err)
		}
		value = strings.TrimSpace(string(b))
	}
	if value == "" {
		return "", nil
	}
	var tmp any
	if err := json.Unmarshal([]byte(value), &tmp); err != nil {
		return "", usagef("invalid JSON for %s: %v", valueFlag, err)
	}
	return value, nil
}
