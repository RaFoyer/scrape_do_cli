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
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s", r.Method)
		}
		q := r.URL.Query()
		if q.Get("token") != "tok" || q.Get("url") != "https://example.com" {
			t.Fatalf("query=%v", q)
		}
		if q.Get("render") != "true" || q.Get("super") != "true" {
			t.Fatalf("query=%v", q)
		}
		if q.Get("customHeaders") != "true" || q.Get("extraHeaders") != "true" {
			t.Fatalf("query=%v", q)
		}
		if q.Get("setCookies") != "session=abc;" {
			t.Fatalf("query=%v", q)
		}
		if r.Header.Get("X-Test") != "1" {
			t.Fatalf("header=%v", r.Header)
		}
		if r.Header.Get("sd-A") != "B" {
			t.Fatalf("header=%v", r.Header)
		}
		if r.Header.Get("content-type") != "application/json" {
			t.Fatalf("header=%v", r.Header)
		}
		b := make([]byte, 64)
		n, _ := r.Body.Read(b)
		if !strings.Contains(string(b[:n]), `"hello":"world"`) {
			t.Fatalf("body=%q", string(b[:n]))
		}
		w.Header().Set("scrape.do-request-cost", "1")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	c := New(Options{Token: "tok", BaseURL: ts.URL, AsyncBaseURL: ts.URL})
	resp, err := c.Scrape(context.Background(), ScrapeRequest{
		Method:        "POST",
		URL:           "https://example.com",
		Body:          `{"hello":"world"}`,
		ContentType:   "application/json",
		Render:        true,
		Super:         true,
		CustomHeaders: true,
		SetCookies:    map[string]string{"session": "abc"},
		Params:        map[string]string{"foo": "bar"},
		Headers:       map[string]string{"X-Test": "1"},
		SDHeaders:     map[string]string{"A": "B"},
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

func TestAsyncEndpoints(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Token") != "tok" {
			t.Fatalf("header=%v", r.Header)
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/jobs":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if _, ok := payload["Targets"].([]any); !ok {
				t.Fatalf("payload=%v", payload)
			}
			_, _ = w.Write([]byte(`{"JobID":"1","Status":"pending"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/jobs":
			_, _ = w.Write([]byte(`{"Jobs":[]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/jobs/1":
			_, _ = w.Write([]byte(`{"JobID":"1","Status":"completed"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/jobs/1/2":
			_, _ = w.Write([]byte(`{"TaskID":"2","Status":"completed"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/jobs/1":
			_, _ = w.Write([]byte(`{"Status":"canceled"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/me":
			_, _ = w.Write([]byte(`{"TotalConcurrency":1}`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()
	c := New(Options{Token: "tok", BaseURL: ts.URL, AsyncBaseURL: ts.URL})

	if _, err := c.AsyncCreateJob(context.Background(), AsyncCreateJobRequest{Targets: []string{"https://example.com"}, Render: true}); err != nil {
		t.Fatalf("AsyncCreateJob: %v", err)
	}
	if _, err := c.AsyncListJobs(context.Background(), 1, 10); err != nil {
		t.Fatalf("AsyncListJobs: %v", err)
	}
	if _, err := c.AsyncGetJob(context.Background(), "1"); err != nil {
		t.Fatalf("AsyncGetJob: %v", err)
	}
	if _, err := c.AsyncGetTask(context.Background(), "1", "2"); err != nil {
		t.Fatalf("AsyncGetTask: %v", err)
	}
	if _, err := c.AsyncCancelJob(context.Background(), "1"); err != nil {
		t.Fatalf("AsyncCancelJob: %v", err)
	}
	if _, err := c.AsyncMe(context.Background()); err != nil {
		t.Fatalf("AsyncMe: %v", err)
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
