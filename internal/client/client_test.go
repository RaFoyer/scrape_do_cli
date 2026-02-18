package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestScrape(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("token") != "tok" || q.Get("url") != "https://example.com" {
			t.Fatalf("query=%v", q)
		}
		if q.Get("render") != "true" || q.Get("super") != "true" {
			t.Fatalf("query=%v", q)
		}
		if q.Get("foo") != "bar" {
			t.Fatalf("query=%v", q)
		}
		if q.Get("extraHeaders") != "true" {
			t.Fatalf("query=%v", q)
		}
		if r.Header.Get("X-Test") != "1" {
			t.Fatalf("header=%v", r.Header)
		}
		if r.Header.Get("sd-A") != "B" {
			t.Fatalf("header=%v", r.Header)
		}
		w.Header().Set("scrape.do-request-cost", "1")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	c := New(Options{Token: "tok", BaseURL: ts.URL, AsyncBaseURL: ts.URL})
	resp, err := c.Scrape(context.Background(), ScrapeRequest{
		URL:       "https://example.com",
		Render:    true,
		Super:     true,
		Params:    map[string]string{"foo": "bar"},
		Headers:   map[string]string{"X-Test": "1"},
		SDHeaders: map[string]string{"A": "B"},
	})
	if err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	body, ok := resp.Body.(map[string]any)
	if !ok || body["ok"] != true {
		t.Fatalf("body=%#v", resp.Body)
	}
	if resp.SDOHeaders["scrape.do-request-cost"] != "1" {
		t.Fatalf("sdo headers=%v", resp.SDOHeaders)
	}
}

func TestInfoAndPluginRun(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/info":
			if r.URL.Query().Get("token") != "tok" {
				t.Fatalf("info query=%v", r.URL.Query())
			}
			_, _ = w.Write([]byte(`{"IsActive":true}`))
		case r.URL.Path == "/plugin/amazon/pdp":
			if r.URL.Query().Get("token") != "tok" || r.URL.Query().Get("url") == "" {
				t.Fatalf("plugin query=%v", r.URL.Query())
			}
			_, _ = w.Write([]byte(`{"product":"ok"}`))
		default:
			t.Fatalf("unexpected path=%s", r.URL.Path)
		}
	}))
	defer ts.Close()
	c := New(Options{Token: "tok", BaseURL: ts.URL, AsyncBaseURL: ts.URL})

	info, err := c.Info(context.Background())
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if info.StatusCode != 200 {
		t.Fatalf("status=%d", info.StatusCode)
	}

	plugin, err := c.PluginRun(context.Background(), PluginRequest{PluginPath: "amazon/pdp", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("PluginRun: %v", err)
	}
	if plugin.StatusCode != 200 {
		t.Fatalf("status=%d", plugin.StatusCode)
	}
}

func TestAsyncSubmitAndStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "tok" {
			t.Fatalf("header=%v", r.Header)
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/job":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload["url"] != "https://example.com" {
				t.Fatalf("payload=%v", payload)
			}
			_, _ = w.Write([]byte(`{"job_id":"1","status":"queued"}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/job/"):
			_, _ = w.Write([]byte(`{"job_id":"1","status":"completed"}`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()
	c := New(Options{Token: "tok", BaseURL: ts.URL, AsyncBaseURL: ts.URL})

	submit, err := c.AsyncSubmit(context.Background(), AsyncSubmitRequest{URL: "https://example.com", Render: true})
	if err != nil {
		t.Fatalf("AsyncSubmit: %v", err)
	}
	if submit.StatusCode != 200 {
		t.Fatalf("status=%d", submit.StatusCode)
	}
	status, err := c.AsyncStatus(context.Background(), "1")
	if err != nil {
		t.Fatalf("AsyncStatus: %v", err)
	}
	if status.StatusCode != 200 {
		t.Fatalf("status=%d", status.StatusCode)
	}
}

func TestAPIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"Message":["bad token"]}`))
	}))
	defer ts.Close()
	c := New(Options{Token: "tok", BaseURL: ts.URL, AsyncBaseURL: ts.URL})
	_, err := c.Info(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err=%T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status=%d", apiErr.StatusCode)
	}
}
