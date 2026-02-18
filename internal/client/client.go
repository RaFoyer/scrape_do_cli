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
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL      = "https://api.scrape.do"
	defaultAsyncBaseURL = "https://q.scrape.do"
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
	addOptionalBool(params, "customHeaders", req.CustomHeaders)
	addOptionalBool(params, "forwardHeaders", req.ForwardHeaders)
	addOptionalBool(params, "disableRedirection", req.DisableRedirection)
	addOptionalBool(params, "disableRetry", req.DisableRetry)
	addOptionalBool(params, "transparentResponse", req.TransparentResponse)
	addOptionalBool(params, "blockResources", req.BlockResources)
	addOptionalBool(params, "screenShot", req.Screenshot)
	addOptionalBool(params, "fullScreenShot", req.FullScreenshot)
	addOptionalBool(params, "returnJSON", req.ReturnJSON)
	addOptionalBool(params, "showWebsocketRequests", req.ShowWebsocketRequests)
	addOptionalBool(params, "showFrames", req.ShowFrames)
	addOptionalString(params, "waitUntil", req.WaitUntil)
	addOptionalString(params, "waitSelector", req.WaitSelector)
	addOptionalString(params, "particularScreenShot", req.ParticularScreenshot)
	addOptionalInt(params, "customWait", req.CustomWait)
	addOptionalInt(params, "timeout", req.RequestTimeoutMS)
	addOptionalInt(params, "retryTimeout", req.RetryTimeoutMS)

	if len(req.SetCookies) > 0 {
		params["setCookies"] = cookiesString(req.SetCookies)
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

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}
	body := strings.TrimSpace(req.Body)
	if method == http.MethodGet && body != "" {
		return nil, fmt.Errorf("GET method does not support body")
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	if strings.TrimSpace(req.ContentType) != "" {
		headers["content-type"] = strings.TrimSpace(req.ContentType)
	}

	return c.request(ctx, method, c.baseURL+"/", params, headers, bodyReader)
}

func (c *Client) Info(ctx context.Context) (*APIResponse, error) {
	params := map[string]string{"token": c.token}
	return c.request(ctx, http.MethodGet, c.baseURL+"/info", params, nil, nil)
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
	return c.request(ctx, http.MethodGet, endpoint, params, headers, nil)
}

func (c *Client) AsyncCreateJob(ctx context.Context, req AsyncCreateJobRequest) (*APIResponse, error) {
	payload := map[string]any{}
	if len(req.Targets) > 0 {
		payload["Targets"] = req.Targets
	}
	if req.Method != "" {
		payload["Method"] = strings.ToUpper(strings.TrimSpace(req.Method))
	}
	addOptionalPayloadString(payload, "Body", req.Body)
	addOptionalPayloadString(payload, "GeoCode", req.GeoCode)
	addOptionalPayloadString(payload, "RegionalGeoCode", req.RegionalGeoCode)
	if req.Super {
		payload["Super"] = true
	}
	if len(req.Headers) > 0 {
		payload["Headers"] = req.Headers
	}
	if req.ForwardHeaders {
		payload["ForwardHeaders"] = true
	}
	addOptionalPayloadString(payload, "SessionID", req.SessionID)
	addOptionalPayloadString(payload, "Device", req.Device)
	if len(req.SetCookies) > 0 {
		payload["SetCookies"] = req.SetCookies
	}
	addOptionalPayloadInt(payload, "Timeout", req.Timeout)
	addOptionalPayloadInt(payload, "RetryTimeout", req.RetryTimeout)
	if req.DisableRetry {
		payload["DisableRetry"] = true
	}
	if req.TransparentResponse {
		payload["TransparentResponse"] = true
	}
	if req.DisableRedirection {
		payload["DisableRedirection"] = true
	}
	addOptionalPayloadString(payload, "Output", req.Output)

	renderPayload := map[string]any{}
	if req.Render {
		renderPayload["Enabled"] = true
	}
	addOptionalPayloadString(renderPayload, "WaitUntil", req.WaitUntil)
	addOptionalPayloadInt(renderPayload, "CustomWait", req.CustomWait)
	addOptionalPayloadString(renderPayload, "WaitSelector", req.WaitSelector)
	if req.BlockResources {
		renderPayload["BlockResources"] = true
	}
	if req.ReturnJSON {
		renderPayload["ReturnJSON"] = true
	}
	if req.ShowWebsocketRequests {
		renderPayload["ShowWebsocketRequests"] = true
	}
	if req.ShowFrames {
		renderPayload["ShowFrames"] = true
	}
	if req.Screenshot {
		renderPayload["ScreenShot"] = true
	}
	if req.FullScreenshot {
		renderPayload["FullScreenShot"] = true
	}
	addOptionalPayloadString(renderPayload, "ParticularScreenShot", req.ParticularScreenshot)
	if len(renderPayload) > 0 {
		if len(renderPayload) == 1 {
			if enabled, ok := renderPayload["Enabled"].(bool); ok {
				payload["Render"] = enabled
			} else {
				payload["Render"] = renderPayload
			}
		} else {
			delete(renderPayload, "Enabled")
			payload["Render"] = renderPayload
		}
	}

	if req.WebhookURL != "" {
		payload["WebHook"] = map[string]any{
			"URL":     req.WebhookURL,
			"Headers": req.WebhookHeaders,
		}
	}

	for k, v := range req.Params {
		payload[k] = v
	}

	endpoint := c.asyncBaseURL + "/api/v1/jobs"
	return c.postJSON(ctx, endpoint, payload, map[string]string{"X-Token": c.token})
}

func (c *Client) AsyncGetJob(ctx context.Context, jobID string) (*APIResponse, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil, fmt.Errorf("empty job id")
	}
	endpoint := c.asyncBaseURL + "/api/v1/jobs/" + url.PathEscape(jobID)
	return c.request(ctx, http.MethodGet, endpoint, nil, map[string]string{"X-Token": c.token}, nil)
}

func (c *Client) AsyncGetTask(ctx context.Context, jobID string, taskID string) (*APIResponse, error) {
	jobID = strings.TrimSpace(jobID)
	taskID = strings.TrimSpace(taskID)
	if jobID == "" || taskID == "" {
		return nil, fmt.Errorf("job id and task id are required")
	}
	endpoint := c.asyncBaseURL + "/api/v1/jobs/" + url.PathEscape(jobID) + "/" + url.PathEscape(taskID)
	return c.request(ctx, http.MethodGet, endpoint, nil, map[string]string{"X-Token": c.token}, nil)
}

func (c *Client) AsyncListJobs(ctx context.Context, page int, pageSize int) (*APIResponse, error) {
	params := map[string]string{}
	if page > 0 {
		params["page"] = strconv.Itoa(page)
	}
	if pageSize > 0 {
		params["page_size"] = strconv.Itoa(pageSize)
	}
	endpoint := c.asyncBaseURL + "/api/v1/jobs"
	return c.request(ctx, http.MethodGet, endpoint, params, map[string]string{"X-Token": c.token}, nil)
}

func (c *Client) AsyncCancelJob(ctx context.Context, jobID string) (*APIResponse, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil, fmt.Errorf("empty job id")
	}
	endpoint := c.asyncBaseURL + "/api/v1/jobs/" + url.PathEscape(jobID)
	return c.request(ctx, http.MethodDelete, endpoint, nil, map[string]string{"X-Token": c.token}, nil)
}

func (c *Client) AsyncMe(ctx context.Context) (*APIResponse, error) {
	endpoint := c.asyncBaseURL + "/api/v1/me"
	return c.request(ctx, http.MethodGet, endpoint, nil, map[string]string{"X-Token": c.token}, nil)
}

func (c *Client) request(ctx context.Context, method string, endpoint string, params map[string]string, headers map[string]string, body io.Reader) (*APIResponse, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}
	q := u.Query()
	for _, key := range sortedKeys(params) {
		q.Set(key, params[key])
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
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
	headers = copyStringMap(headers)
	headers["content-type"] = "application/json"
	return c.request(ctx, http.MethodPost, endpoint, nil, headers, bytes.NewReader(b))
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

func addOptionalInt(m map[string]string, key string, value int) {
	if value > 0 {
		m[key] = strconv.Itoa(value)
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

func addOptionalPayloadInt(m map[string]any, key string, value int) {
	if value > 0 {
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

func copyStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cookiesString(cookies map[string]string) string {
	if len(cookies) == 0 {
		return ""
	}
	keys := make([]string, 0, len(cookies))
	for k := range cookies {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, cookies[k]))
	}
	return strings.Join(pairs, ";") + ";"
}

func messageFromBody(body any) string {
	switch v := body.(type) {
	case map[string]any:
		if msg := readStringSlice(v["Message"]); msg != "" {
			return msg
		}
		if msg := readStringSlice(v["Error"]); msg != "" {
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
