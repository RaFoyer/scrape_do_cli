package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ra/scrape_do_cli/internal/outfmt"
)

type DocsCmd struct{}

func (c *DocsCmd) Run(ctx context.Context) error {
	payload := map[string]any{
		"sync_api": map[string]any{
			"base_url": "https://api.scrape.do",
			"endpoints": []string{
				"GET/POST /",
				"GET /info",
				"GET /plugin/{path}",
			},
		},
		"async_api": map[string]any{
			"base_url":    "https://q.scrape.do",
			"auth_header": "X-Token",
			"endpoints": []string{
				"POST /api/v1/jobs",
				"GET /api/v1/jobs",
				"GET /api/v1/jobs/{job_id}",
				"GET /api/v1/jobs/{job_id}/{task_id}",
				"DELETE /api/v1/jobs/{job_id}",
				"GET /api/v1/me",
			},
		},
		"examples": []string{
			"sdo info --json",
			"sdo scrape https://example.com --geo us --render",
			"sdo scrape https://example.com --method POST --body '{\"q\":\"x\"}' --content-type application/json",
			"sdo plugin run amazon/pdp --url https://www.amazon.com/dp/B08N5WRWNW --param asin=B08N5WRWNW --param geocode=us --param countryName='United States'",
			"sdo async me --json",
			"sdo async submit https://example.com --render",
			"sdo async list --page 1 --page-size 20",
			"sdo async status <job_id>",
			"sdo async task <job_id> <task_id>",
			"sdo async wait <job_id> --wait-timeout 5m",
			"sdo async cancel <job_id>",
		},
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, payload)
	}

	fmt.Fprintln(os.Stdout, "sdo endpoint coverage")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Sync API (https://api.scrape.do)")
	fmt.Fprintln(os.Stdout, "  - GET/POST /")
	fmt.Fprintln(os.Stdout, "  - GET /info")
	fmt.Fprintln(os.Stdout, "  - GET /plugin/{path}")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Async API (https://q.scrape.do, auth: X-Token)")
	fmt.Fprintln(os.Stdout, "  - POST /api/v1/jobs")
	fmt.Fprintln(os.Stdout, "  - GET /api/v1/jobs")
	fmt.Fprintln(os.Stdout, "  - GET /api/v1/jobs/{job_id}")
	fmt.Fprintln(os.Stdout, "  - GET /api/v1/jobs/{job_id}/{task_id}")
	fmt.Fprintln(os.Stdout, "  - DELETE /api/v1/jobs/{job_id}")
	fmt.Fprintln(os.Stdout, "  - GET /api/v1/me")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Run `sdo --help` and `sdo <command> --help` for full flags.")
	fmt.Fprintln(os.Stdout, "Run `sdo docs --json` for machine-readable examples.")

	return nil
}
