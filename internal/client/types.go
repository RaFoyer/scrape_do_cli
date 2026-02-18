package client

import "net/http"

type Options struct {
	Token        string
	BaseURL      string
	AsyncBaseURL string
	HTTPClient   *http.Client
}

type Client struct {
	token        string
	baseURL      string
	asyncBaseURL string
	httpClient   *http.Client
}

type ScrapeRequest struct {
	URL         string
	Render      bool
	Super       bool
	Geo         string
	RegionalGeo string
	SessionID   string
	Device      string
	Output      string
	Callback    string
	Params      map[string]string
	Headers     map[string]string
	SDHeaders   map[string]string
}

type PluginRequest struct {
	PluginPath string
	URL        string
	Params     map[string]string
	Headers    map[string]string
	SDHeaders  map[string]string
}

type AsyncSubmitRequest struct {
	URL         string
	Render      bool
	Super       bool
	Geo         string
	RegionalGeo string
	SessionID   string
	Device      string
	Output      string
	Callback    string
	Params      map[string]string
}

type APIResponse struct {
	StatusCode int               `json:"status_code"`
	Body       any               `json:"body"`
	Headers    map[string]string `json:"headers,omitempty"`
	SDOHeaders map[string]string `json:"sdo_headers,omitempty"`
}

type APIError struct {
	StatusCode int
	Body       any
	RawBody    string
}

func (e *APIError) Error() string {
	msg := messageFromBody(e.Body)
	if msg == "" {
		msg = e.RawBody
	}
	if msg == "" {
		return "request failed"
	}
	return msg
}
