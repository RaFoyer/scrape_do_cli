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
	Method                string
	URL                   string
	Body                  string
	ContentType           string
	Render                bool
	Super                 bool
	Geo                   string
	RegionalGeo           string
	SessionID             string
	Device                string
	Output                string
	Callback              string
	CustomHeaders         bool
	ForwardHeaders        bool
	DisableRedirection    bool
	DisableRetry          bool
	TransparentResponse   bool
	PureCookies           bool
	RequestTimeoutMS      int
	RetryTimeoutMS        int
	WaitUntil             string
	CustomWait            int
	WaitSelector          string
	Width                 int
	Height                int
	BlockResources        bool
	Screenshot            bool
	FullScreenshot        bool
	ParticularScreenshot  string
	PlayWithBrowser       string
	ReturnJSON            bool
	ShowWebsocketRequests bool
	ShowFrames            bool
	SetCookies            map[string]string
	Params                map[string]string
	Headers               map[string]string
	SDHeaders             map[string]string
}

type PluginRequest struct {
	PluginPath string
	URL        string
	Params     map[string]string
	Headers    map[string]string
	SDHeaders  map[string]string
}

type AsyncCreateJobRequest struct {
	Targets               []string
	Method                string
	Body                  string
	GeoCode               string
	RegionalGeoCode       string
	Super                 bool
	Headers               map[string]string
	ForwardHeaders        bool
	SessionID             string
	Device                string
	SetCookies            map[string]string
	Timeout               int
	RetryTimeout          int
	DisableRetry          bool
	TransparentResponse   bool
	DisableRedirection    bool
	Output                string
	Render                bool
	WaitUntil             string
	CustomWait            int
	WaitSelector          string
	Width                 int
	Height                int
	BlockResources        bool
	ReturnJSON            bool
	ShowWebsocketRequests bool
	ShowFrames            bool
	Screenshot            bool
	FullScreenshot        bool
	ParticularScreenshot  string
	PlayWithBrowser       any
	WebhookURL            string
	WebhookHeaders        map[string]string
	Params                map[string]string
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
