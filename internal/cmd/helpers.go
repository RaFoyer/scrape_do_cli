package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ra/scrape_do_cli/internal/client"
	"github.com/ra/scrape_do_cli/internal/outfmt"
)

func newClientFromContext(ctx context.Context) *client.Client {
	runtime := runtimeFromContext(ctx)
	return client.New(client.Options{
		Token:        runtime.Token,
		BaseURL:      runtime.BaseURL,
		AsyncBaseURL: runtime.AsyncBaseURL,
		HTTPClient:   nil,
	})
}

func requireToken(ctx context.Context) error {
	runtime := runtimeFromContext(ctx)
	if strings.TrimSpace(runtime.Token) == "" {
		return &ExitError{Code: exitCodeConfig, Err: fmt.Errorf("missing token; pass --token, set SCRAPEDO_TOKEN, or run 'sdo config set token ...'")}
	}
	return nil
}

func parseParams(items []string) (map[string]string, error) {
	out := map[string]string{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			return nil, usagef("invalid --param %q (expected key=value)", item)
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		if k == "" {
			return nil, usagef("invalid --param %q (empty key)", item)
		}
		out[k] = v
	}
	return out, nil
}

func parseHeaders(items []string) (map[string]string, error) {
	out := map[string]string{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, ":", 2)
		if len(parts) != 2 {
			return nil, usagef("invalid header %q (expected 'Key: Value')", item)
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		if k == "" {
			return nil, usagef("invalid header %q (empty key)", item)
		}
		out[k] = v
	}
	return out, nil
}

func writePayload(ctx context.Context, payload any) error {
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, payload)
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("encode output: %w", err)
	}
	fmt.Fprintln(os.Stdout, string(b))
	return nil
}

func writeResponse(ctx context.Context, response *client.APIResponse, extra map[string]any) error {
	payload := map[string]any{
		"status_code": response.StatusCode,
		"content":     response.Body,
	}
	if len(response.SDOHeaders) > 0 {
		payload["sdo_headers"] = response.SDOHeaders
	}
	for k, v := range extra {
		payload[k] = v
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, payload)
	}

	fmt.Fprintf(os.Stdout, "status_code\t%d\n", response.StatusCode)
	if len(response.SDOHeaders) > 0 {
		keys := make([]string, 0, len(response.SDOHeaders))
		for k := range response.SDOHeaders {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(os.Stdout, "%s\t%s\n", k, response.SDOHeaders[k])
		}
	}
	body := response.Body
	switch v := body.(type) {
	case string:
		if strings.TrimSpace(v) != "" {
			fmt.Fprintln(os.Stdout, v)
		}
	default:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stdout, "%v\n", v)
			return nil
		}
		fmt.Fprintln(os.Stdout, string(b))
	}
	return nil
}

func extractStatus(v any) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	for _, key := range []string{"status", "state", "job_status"} {
		if s, ok := m[key].(string); ok {
			s = strings.TrimSpace(strings.ToLower(s))
			if s != "" {
				return s
			}
		}
	}
	if done, ok := m["done"].(bool); ok {
		if done {
			return "done"
		}
		return "running"
	}
	if completed, ok := m["completed"].(bool); ok {
		if completed {
			return "completed"
		}
	}
	return ""
}

func isTerminalStatus(status string) bool {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "done", "completed", "success", "succeeded", "failed", "error", "cancelled", "canceled":
		return true
	default:
		return false
	}
}

func isFailureStatus(status string) bool {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "failed", "error", "cancelled", "canceled":
		return true
	default:
		return false
	}
}
