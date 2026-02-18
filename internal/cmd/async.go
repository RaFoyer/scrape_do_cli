package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ra/scrape_do_cli/internal/client"
	"github.com/ra/scrape_do_cli/internal/outfmt"
)

type AsyncCmd struct {
	Submit AsyncSubmitCmd `cmd:"" help:"Submit asynchronous scrape job"`
	Status AsyncStatusCmd `cmd:"" help:"Fetch asynchronous job status"`
	Wait   AsyncWaitCmd   `cmd:"" help:"Wait until asynchronous job reaches terminal state"`
}

type AsyncSubmitCmd struct {
	URL         string   `arg:"" name:"url" help:"Target URL to scrape"`
	Render      bool     `name:"render" help:"Execute JavaScript rendering"`
	Super       bool     `name:"super" help:"Use super (residential/mobile) proxy"`
	Geo         string   `name:"geo" aliases:"geocode" help:"Country geocode (e.g. us)"`
	RegionalGeo string   `name:"regional-geo" help:"Regional geocode (europe, asia, ...)"`
	SessionID   string   `name:"session-id" help:"Sticky proxy session ID"`
	Device      string   `name:"device" help:"Device profile (Desktop|Mobile)"`
	APIOutput   string   `name:"output" help:"API output mode (raw|markdown)"`
	Callback    string   `name:"callback" help:"Callback URL"`
	Param       []string `name:"param" help:"Additional payload field (key=value)"`
}

func (c *AsyncSubmitCmd) Run(ctx context.Context) error {
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

	resp, err := newClientFromContext(ctx).AsyncSubmit(ctx, client.AsyncSubmitRequest{
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
	})
	if err != nil {
		return addAsyncDNSHint(err)
	}
	return writeResponse(ctx, resp, map[string]any{"url": c.URL})
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
	resp, err := newClientFromContext(ctx).AsyncStatus(ctx, c.JobID)
	if err != nil {
		return addAsyncDNSHint(err)
	}
	return writeResponse(ctx, resp, map[string]any{"job_id": c.JobID})
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
		return usage("--timeout must be > 0")
	}

	deadline := time.Now().Add(c.Timeout)
	for {
		resp, err := newClientFromContext(ctx).AsyncStatus(ctx, c.JobID)
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
		return fmt.Errorf("%w (tip: set --async-base-url or SCRAPEDO_ASYNC_BASE_URL if your network cannot resolve async.scrape.do)", err)
	}
	return err
}
