package cmd

import "testing"

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
