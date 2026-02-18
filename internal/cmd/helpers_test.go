package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ra/scrape_do_cli/internal/client"
)

func TestParseParams(t *testing.T) {
	params, err := parseParams([]string{"a=1", "b=2"})
	if err != nil {
		t.Fatalf("parseParams: %v", err)
	}
	if params["a"] != "1" || params["b"] != "2" {
		t.Fatalf("params=%v", params)
	}
	if _, err := parseParams([]string{"broken"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseCookies(t *testing.T) {
	cookies, err := parseCookies([]string{"session=abc", "k=v"})
	if err != nil {
		t.Fatalf("parseCookies: %v", err)
	}
	if cookies["session"] != "abc" || cookies["k"] != "v" {
		t.Fatalf("cookies=%v", cookies)
	}
	if _, err := parseCookies([]string{"broken"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseHeaders(t *testing.T) {
	headers, err := parseHeaders([]string{"A: B", "X-Test: 1"})
	if err != nil {
		t.Fatalf("parseHeaders: %v", err)
	}
	if headers["A"] != "B" || headers["X-Test"] != "1" {
		t.Fatalf("headers=%v", headers)
	}
	if _, err := parseHeaders([]string{"broken"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestStatusHelpers(t *testing.T) {
	if got := extractStatus(map[string]any{"status": "Completed"}); got != "completed" {
		t.Fatalf("status=%q", got)
	}
	if !isTerminalStatus("completed") || isFailureStatus("completed") {
		t.Fatalf("unexpected status evaluation")
	}
	if !isTerminalStatus("error") || !isFailureStatus("error") {
		t.Fatalf("unexpected status evaluation")
	}
}

func TestWritePlainResponse(t *testing.T) {
	resp := &client.APIResponse{
		StatusCode: 200,
		Body:       map[string]any{"ok": true},
		SDOHeaders: map[string]string{"scrape.do-request-cost": "1"},
	}
	extra := map[string]any{"plugin_path": "amazon/pdp"}
	out := captureStdout(t, func() {
		if err := writePlainResponse(resp, extra); err != nil {
			t.Fatalf("writePlainResponse: %v", err)
		}
	})
	if !strings.Contains(out, "status_code\t200\n") {
		t.Fatalf("output=%q", out)
	}
	if !strings.Contains(out, "plugin_path\t\"amazon/pdp\"\n") {
		t.Fatalf("output=%q", out)
	}
	if !strings.Contains(out, "sdo_headers.scrape.do-request-cost\t\"1\"\n") {
		t.Fatalf("output=%q", out)
	}
	if !strings.Contains(out, "content\t{\"ok\":true}\n") {
		t.Fatalf("output=%q", out)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	return string(b)
}
