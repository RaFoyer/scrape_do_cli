package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"
)

const (
	defaultBaseURL      = "https://api.scrape.do"
	defaultAsyncBaseURL = "https://async.scrape.do"
)

func New(opts Options) *Client {
	baseURL := strings.TrimSpace(opts.BaseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	asyncBaseURL := strings.TrimSpace(opts.AsyncBaseURL)
	if asyncBaseURL == "" {
		asyncBaseURL = defaultAsyncBaseURL
	}
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		token:        strings.TrimSpace(opts.Token),
		baseURL:      strings.TrimRight(baseURL, "/"),
		asyncBaseURL: strings.TrimRight(asyncBaseURL, "/"),
		httpClient:   httpClient,
	}
}

func (c *Client) Scrape(ctx context.Context, req ScrapeRequest) (*APIResponse, error) {
	params := map[string]string{
		"token": c.token,
		"url":   req.URL,
	}
	addOptionalBool(params, "render", req.Render)
	addOptionalBool(params, "super", req.Super)
	addOptionalString(params, "geoCode", req.Geo)
	addOptionalString(params, "regionalGeoCode", req.RegionalGeo)
	addOptionalString(params, "sessionId", req.SessionID)
	addOptionalString(params, "device", req.Device)
	addOptionalString(params, "output", req.Output)
	addOptionalString(params, "callback", req.Callback)
	mergeParams(params, req.Params)

	headers := map[string]string{}
	mergeParams(headers, req.Headers)
	if len(req.SDHeaders) > 0 {
		params["extraHeaders"] = "true"
		for k, v := range req.SDHeaders {
			headers[normalizeSDHeaderKey(k)] = v
		}
	}

	return c.get(ctx, c.baseURL+"/", params, headers)
}

func (c *Client) Info(ctx context.Context) (*APIResponse, error) {
	params := map[string]string{"token": c.token}
	return c.get(ctx, c.baseURL+"/info", params, nil)
}

func (c *Client) PluginRun(ctx context.Context, req PluginRequest) (*APIResponse, error) {
	pluginPath := strings.TrimSpace(req.PluginPath)
	pluginPath = strings.TrimPrefix(pluginPath, "/")
	if strings.HasPrefix(pluginPath, "plugin/") {
		pluginPath = strings.TrimPrefix(pluginPath, "plugin/")
	}
	if pluginPath == "" {
		return nil, fmt.Errorf("empty plugin path")
	}
	params := map[string]string{
		"token": c.token,
		"url":   req.URL,
	}
	mergeParams(params, req.Params)

	headers := map[string]string{}
	mergeParams(headers, req.Headers)
	if len(req.SDHeaders) > 0 {
		params["extraHeaders"] = "true"
		for k, v := range req.SDHeaders {
			headers[normalizeSDHeaderKey(k)] = v
		}
	}

	endpoint := c.baseURL + "/" + path.Join("plugin", pluginPath)
	return c.get(ctx, endpoint, params, headers)
}

func (c *Client) AsyncSubmit(ctx context.Context, req AsyncSubmitRequest) (*APIResponse, error) {
	payload := map[string]any{"url": req.URL}
	if req.Render {
		payload["render"] = true
	}
	if req.Super {
		payload["super"] = true
	}
	addOptionalPayloadString(payload, "geoCode", req.Geo)
	addOptionalPayloadString(payload, "regionalGeoCode", req.RegionalGeo)
	addOptionalPayloadString(payload, "sessionId", req.SessionID)
	addOptionalPayloadString(payload, "device", req.Device)
	addOptionalPayloadString(payload, "output", req.Output)
	addOptionalPayloadString(payload, "callback", req.Callback)
	for k, v := range req.Params {
		payload[k] = v
	}

	return c.postJSON(ctx, c.asyncBaseURL+"/job", payload, map[string]string{
		"x-api-key": c.token,
	})
}

func (c *Client) AsyncStatus(ctx context.Context, jobID string) (*APIResponse, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil, fmt.Errorf("empty job id")
	}
	endpoint := c.asyncBaseURL + "/job/" + url.PathEscape(jobID)
	return c.get(ctx, endpoint, nil, map[string]string{"x-api-key": c.token})
}

func (c *Client) get(ctx context.Context, endpoint string, params map[string]string, headers map[string]string) (*APIResponse, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}
	q := u.Query()
	for _, key := range sortedKeys(params) {
		q.Set(key, params[key])
	}
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	for _, key := range sortedKeys(headers) {
		req.Header.Set(key, headers[key])
	}
	return c.do(req)
}

func (c *Client) postJSON(ctx context.Context, endpoint string, payload map[string]any, headers map[string]string) (*APIResponse, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	for _, key := range sortedKeys(headers) {
		req.Header.Set(key, headers[key])
	}
	return c.do(req)
}

func (c *Client) do(req *http.Request) (*APIResponse, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	body := parseBody(b)
	apiResp := &APIResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    headersToMap(resp.Header),
		SDOHeaders: extractSDOHeaders(resp.Header),
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: body, RawBody: strings.TrimSpace(string(b))}
	}
	return apiResp, nil
}

func parseBody(b []byte) any {
	trimmed := bytes.TrimSpace(b)
	if len(trimmed) == 0 {
		return ""
	}
	var v any
	if err := json.Unmarshal(trimmed, &v); err == nil {
		return v
	}
	return string(trimmed)
}

func headersToMap(headers http.Header) map[string]string {
	out := make(map[string]string, len(headers))
	for k, vals := range headers {
		out[strings.ToLower(k)] = strings.Join(vals, ",")
	}
	return out
}

func extractSDOHeaders(headers http.Header) map[string]string {
	out := map[string]string{}
	for k, vals := range headers {
		lk := strings.ToLower(k)
		if strings.HasPrefix(lk, "scrape.do-") {
			out[lk] = strings.Join(vals, ",")
		}
	}
	return out
}

func normalizeSDHeaderKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return key
	}
	if strings.HasPrefix(strings.ToLower(key), "sd-") {
		return key
	}
	return "sd-" + key
}

func addOptionalBool(m map[string]string, key string, value bool) {
	if value {
		m[key] = "true"
	}
}

func addOptionalString(m map[string]string, key string, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		m[key] = value
	}
}

func addOptionalPayloadString(m map[string]any, key string, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		m[key] = value
	}
}

func mergeParams(dst map[string]string, src map[string]string) {
	for k, v := range src {
		if strings.TrimSpace(k) == "" {
			continue
		}
		dst[k] = v
	}
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func messageFromBody(body any) string {
	switch v := body.(type) {
	case map[string]any:
		if msg := readStringSlice(v["Message"]); msg != "" {
			return msg
		}
		for _, k := range []string{"message", "error", "detail", "status"} {
			if s, ok := v[k].(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	case string:
		return strings.TrimSpace(v)
	}
	return ""
}

func readStringSlice(v any) string {
	switch vv := v.(type) {
	case []any:
		parts := make([]string, 0, len(vv))
		for _, item := range vv {
			s, ok := item.(string)
			if !ok {
				continue
			}
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			parts = append(parts, s)
		}
		return strings.Join(parts, "; ")
	case []string:
		return strings.Join(vv, "; ")
	case string:
		return strings.TrimSpace(vv)
	default:
		return ""
	}
}
