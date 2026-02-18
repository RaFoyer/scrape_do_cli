package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ra/scrape_do_cli/internal/client"
	"github.com/ra/scrape_do_cli/internal/outfmt"
)

type AsyncCmd struct {
	Submit AsyncSubmitCmd `cmd:"" help:"Create asynchronous scrape job"`
	Status AsyncStatusCmd `cmd:"" help:"Fetch asynchronous job status"`
	Task   AsyncTaskCmd   `cmd:"" help:"Fetch a specific task in an asynchronous job"`
	List   AsyncListCmd   `cmd:"" help:"List asynchronous jobs"`
	Cancel AsyncCancelCmd `cmd:"" help:"Cancel an asynchronous job"`
	Me     AsyncMeCmd     `cmd:"" help:"Show async API account concurrency and credits"`
	Wait   AsyncWaitCmd   `cmd:"" help:"Wait until asynchronous job reaches terminal state"`
}

type AsyncSubmitCmd struct {
	URLs                  []string `arg:"" optional:"" name:"url" help:"One or more target URLs"`
	Target                []string `name:"target" help:"Additional target URL (repeatable)"`
	Method                string   `name:"method" help:"HTTP method for targets" default:"GET" enum:"GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS"`
	Body                  string   `name:"body" help:"Raw request body for non-GET methods"`
	BodyFile              string   `name:"body-file" help:"Read request body from file for non-GET methods"`
	GeoCode               string   `name:"geo" aliases:"geocode" help:"Country geocode (e.g. us)"`
	RegionalGeoCode       string   `name:"regional-geo" help:"Regional geocode (europe, asia, ...)"`
	Super                 bool     `name:"super" help:"Use super (residential/mobile) proxy"`
	Header                []string `name:"header" help:"Target header (format: 'Key: Value')"`
	ForwardHeaders        bool     `name:"forward-headers" help:"Forward target headers directly"`
	SessionID             string   `name:"session-id" help:"Sticky proxy session ID"`
	Device                string   `name:"device" help:"Device profile (desktop|mobile|tablet)"`
	SetCookie             []string `name:"set-cookie" help:"Cookie for targets (key=value); repeatable"`
	RequestTimeoutMS      int      `name:"request-timeout-ms" help:"Target timeout in milliseconds"`
	RetryTimeoutMS        int      `name:"retry-timeout-ms" help:"Retry timeout in milliseconds"`
	DisableRetry          bool     `name:"disable-retry" help:"Disable automatic retries"`
	TransparentResponse   bool     `name:"transparent-response" help:"Return full target response without status validation"`
	DisableRedirection    bool     `name:"disable-redirection" help:"Disable following redirects"`
	Output                string   `name:"output" help:"Output mode (raw|markdown)"`
	Render                bool     `name:"render" help:"Enable browser rendering"`
	WaitUntil             string   `name:"wait-until" help:"Render waitUntil mode"`
	CustomWait            int      `name:"custom-wait" help:"Render custom wait in milliseconds"`
	WaitSelector          string   `name:"wait-selector" help:"Render wait for CSS selector"`
	Width                 int      `name:"width" help:"Render viewport width in pixels"`
	Height                int      `name:"height" help:"Render viewport height in pixels"`
	BlockResources        bool     `name:"block-resources" help:"Block non-essential resources during render"`
	ReturnJSON            bool     `name:"return-json" help:"Return JSON render payload"`
	ShowWebsocketRequests bool     `name:"show-websocket-requests" help:"Include websocket request details in render output"`
	ShowFrames            bool     `name:"show-frames" help:"Include frame details in render output"`
	Screenshot            bool     `name:"screenshot" help:"Capture viewport screenshot"`
	FullScreenshot        bool     `name:"full-screenshot" help:"Capture full page screenshot"`
	ParticularScreenshot  string   `name:"particular-screenshot" help:"Capture screenshot for a CSS selector"`
	PlayWithBrowser       string   `name:"play-with-browser" help:"PlayWithBrowser actions as JSON array/object"`
	PlayWithBrowserFile   string   `name:"play-with-browser-file" help:"Read PlayWithBrowser JSON from file"`
	WebhookURL            string   `name:"webhook-url" help:"Webhook URL for async completion callback"`
	WebhookHeader         []string `name:"webhook-header" help:"Webhook header (format: 'Key: Value')"`
	Param                 []string `name:"param" help:"Additional payload field (key=value)"`
}

func (c *AsyncSubmitCmd) Run(ctx context.Context) error {
	if err := requireToken(ctx); err != nil {
		return err
	}
	targets := append([]string{}, c.URLs...)
	targets = append(targets, c.Target...)
	for i := range targets {
		targets[i] = strings.TrimSpace(targets[i])
	}
	cleanTargets := make([]string, 0, len(targets))
	for _, t := range targets {
		if t != "" {
			cleanTargets = append(cleanTargets, t)
		}
	}
	if len(cleanTargets) == 0 {
		return usage("missing target URL (provide <url> or --target)")
	}

	method := strings.ToUpper(strings.TrimSpace(c.Method))
	body, err := readBodyInput(c.Body, c.BodyFile)
	if err != nil {
		return err
	}
	if method == "GET" && strings.TrimSpace(body) != "" {
		return usage("GET method does not support body; use --method POST/PUT/PATCH/DELETE")
	}
	playWithBrowserRaw, err := readJSONInput(c.PlayWithBrowser, c.PlayWithBrowserFile, "--play-with-browser", "--play-with-browser-file")
	if err != nil {
		return err
	}
	playWithBrowser, err := parseJSONAny(playWithBrowserRaw, "--play-with-browser")
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
	cookies, err := parseCookies(c.SetCookie)
	if err != nil {
		return err
	}
	webhookHeaders, err := parseHeaders(c.WebhookHeader)
	if err != nil {
		return err
	}

	if c.RequestTimeoutMS < 0 || c.RetryTimeoutMS < 0 || c.CustomWait < 0 {
		return usage("timeout values must be non-negative")
	}
	if c.Width < 0 || c.Height < 0 {
		return usage("--width and --height must be non-negative")
	}

	resp, err := newClientFromContext(ctx).AsyncCreateJob(ctx, client.AsyncCreateJobRequest{
		Targets:               cleanTargets,
		Method:                method,
		Body:                  body,
		GeoCode:               c.GeoCode,
		RegionalGeoCode:       c.RegionalGeoCode,
		Super:                 c.Super,
		Headers:               headers,
		ForwardHeaders:        c.ForwardHeaders,
		SessionID:             c.SessionID,
		Device:                c.Device,
		SetCookies:            cookies,
		Timeout:               c.RequestTimeoutMS,
		RetryTimeout:          c.RetryTimeoutMS,
		DisableRetry:          c.DisableRetry,
		TransparentResponse:   c.TransparentResponse,
		DisableRedirection:    c.DisableRedirection,
		Output:                c.Output,
		Render:                c.Render,
		WaitUntil:             c.WaitUntil,
		CustomWait:            c.CustomWait,
		WaitSelector:          c.WaitSelector,
		Width:                 c.Width,
		Height:                c.Height,
		BlockResources:        c.BlockResources,
		ReturnJSON:            c.ReturnJSON,
		ShowWebsocketRequests: c.ShowWebsocketRequests,
		ShowFrames:            c.ShowFrames,
		Screenshot:            c.Screenshot,
		FullScreenshot:        c.FullScreenshot,
		ParticularScreenshot:  c.ParticularScreenshot,
		PlayWithBrowser:       playWithBrowser,
		WebhookURL:            c.WebhookURL,
		WebhookHeaders:        webhookHeaders,
		Params:                params,
	})
	if err != nil {
		return addAsyncDNSHint(err)
	}
	return writeResponse(ctx, resp, map[string]any{"targets": cleanTargets})
}

type AsyncStatusCmd struct {
	JobID string `arg:"" name:"job_id" help:"Async job ID"`
}

func (c *AsyncStatusCmd) Run(ctx context.Context) error {
	if err := requireToken(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(c.JobID) == "" {
		return usage("missing job_id")
	}
	resp, err := newClientFromContext(ctx).AsyncGetJob(ctx, c.JobID)
	if err != nil {
		return addAsyncDNSHint(err)
	}
	return writeResponse(ctx, resp, map[string]any{"job_id": c.JobID})
}

type AsyncTaskCmd struct {
	JobID  string `arg:"" name:"job_id" help:"Async job ID"`
	TaskID string `arg:"" name:"task_id" help:"Task ID within the job"`
}

func (c *AsyncTaskCmd) Run(ctx context.Context) error {
	if err := requireToken(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(c.JobID) == "" || strings.TrimSpace(c.TaskID) == "" {
		return usage("missing job_id or task_id")
	}
	resp, err := newClientFromContext(ctx).AsyncGetTask(ctx, c.JobID, c.TaskID)
	if err != nil {
		return addAsyncDNSHint(err)
	}
	return writeResponse(ctx, resp, map[string]any{"job_id": c.JobID, "task_id": c.TaskID})
}

type AsyncListCmd struct {
	Page     int `name:"page" help:"Page number" default:"1"`
	PageSize int `name:"page-size" help:"Page size" default:"20"`
}

func (c *AsyncListCmd) Run(ctx context.Context) error {
	if err := requireToken(ctx); err != nil {
		return err
	}
	if c.Page <= 0 || c.PageSize <= 0 {
		return usage("--page and --page-size must be > 0")
	}
	resp, err := newClientFromContext(ctx).AsyncListJobs(ctx, c.Page, c.PageSize)
	if err != nil {
		return addAsyncDNSHint(err)
	}
	return writeResponse(ctx, resp, map[string]any{"page": c.Page, "page_size": c.PageSize})
}

type AsyncCancelCmd struct {
	JobID string `arg:"" name:"job_id" help:"Async job ID"`
}

func (c *AsyncCancelCmd) Run(ctx context.Context) error {
	if err := requireToken(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(c.JobID) == "" {
		return usage("missing job_id")
	}
	resp, err := newClientFromContext(ctx).AsyncCancelJob(ctx, c.JobID)
	if err != nil {
		return addAsyncDNSHint(err)
	}
	return writeResponse(ctx, resp, map[string]any{"job_id": c.JobID})
}

type AsyncMeCmd struct{}

func (c *AsyncMeCmd) Run(ctx context.Context) error {
	if err := requireToken(ctx); err != nil {
		return err
	}
	resp, err := newClientFromContext(ctx).AsyncMe(ctx)
	if err != nil {
		return addAsyncDNSHint(err)
	}
	return writeResponse(ctx, resp, nil)
}

type AsyncWaitCmd struct {
	JobID    string        `arg:"" name:"job_id" help:"Async job ID"`
	Interval time.Duration `name:"interval" help:"Polling interval" default:"2s"`
	Timeout  time.Duration `name:"wait-timeout" help:"Total wait timeout" default:"2m"`
}

func (c *AsyncWaitCmd) Run(ctx context.Context) error {
	if err := requireToken(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(c.JobID) == "" {
		return usage("missing job_id")
	}
	if c.Interval <= 0 {
		return usage("--interval must be > 0")
	}
	if c.Timeout <= 0 {
		return usage("--wait-timeout must be > 0")
	}

	deadline := time.Now().Add(c.Timeout)
	for {
		resp, err := newClientFromContext(ctx).AsyncGetJob(ctx, c.JobID)
		if err != nil {
			return addAsyncDNSHint(err)
		}
		status := extractStatus(resp.Body)
		if outfmt.IsJSON(ctx) {
			if isTerminalStatus(status) {
				if err := writeResponse(ctx, resp, map[string]any{"job_id": c.JobID, "status": status}); err != nil {
					return err
				}
				if isFailureStatus(status) {
					return fmt.Errorf("async job ended with status %q", status)
				}
				return nil
			}
		} else {
			if status != "" {
				fmt.Fprintf(os.Stderr, "job %s status: %s\n", c.JobID, status)
			}
			if isTerminalStatus(status) {
				if err := writeResponse(ctx, resp, map[string]any{"job_id": c.JobID, "status": status}); err != nil {
					return err
				}
				if isFailureStatus(status) {
					return fmt.Errorf("async job ended with status %q", status)
				}
				return nil
			}
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for async job %s", c.JobID)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.Interval):
		}
	}
}

func addAsyncDNSHint(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "no such host") || strings.Contains(msg, "could not resolve host") {
		return fmt.Errorf("%w (tip: set --async-base-url or SCRAPEDO_ASYNC_BASE_URL if your network cannot resolve q.scrape.do)", err)
	}
	return err
}

func parseJSONAny(raw string, flag string) (any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var out any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, usagef("invalid JSON for %s: %v", flag, err)
	}
	return out, nil
}
