package httpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/geekpratyush/conduit/internal/plugin"
)

// RESTConnector executes HTTP requests. It implements plugin.Connector so the
// shell can list it and manage its lifecycle, and adds Send for the actual
// request/response work the REST view drives.
type RESTConnector struct {
	client *http.Client
}

// compile-time check that RESTConnector satisfies the SPI.
var _ plugin.Connector = (*RESTConnector)(nil)

// New returns a REST connector with sensible transport defaults (HTTP/2 enabled
// for https via the default transport, redirects followed).
func New() *RESTConnector {
	return &RESTConnector{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *RESTConnector) Protocol() string    { return "rest" }
func (c *RESTConnector) DisplayName() string { return "REST / HTTP" }

// Connect is a no-op for REST — each request is independent — but validates that
// the target URL parses so the shell's "Connect" action gives immediate feedback.
func (c *RESTConnector) Connect(ctx context.Context, cfg plugin.ConnectionConfig) error {
	if cfg.URL == "" {
		return nil
	}
	if _, err := url.ParseRequestURI(cfg.URL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	return nil
}

// Test issues a lightweight GET against cfg.URL and reports reachability + latency.
func (c *RESTConnector) Test(ctx context.Context, cfg plugin.ConnectionConfig) plugin.TestResult {
	if cfg.URL == "" {
		return plugin.TestResult{OK: false, Message: "no URL configured"}
	}
	start := time.Now()
	resp, err := c.Send(ctx, Request{Method: http.MethodGet, URL: cfg.URL})
	if err != nil {
		return plugin.TestResult{OK: false, Message: err.Error(), Latency: time.Since(start)}
	}
	return plugin.TestResult{
		OK:      resp.Status < 500,
		Message: fmt.Sprintf("%d %s", resp.Status, resp.StatusText),
		Latency: resp.Elapsed,
	}
}

// Close releases idle connections.
func (c *RESTConnector) Close() error {
	c.client.CloseIdleConnections()
	return nil
}

// Send executes req and returns the response with timing and size populated.
func (c *RESTConnector) Send(ctx context.Context, req Request) (*Response, error) {
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}

	u, err := url.Parse(strings.TrimSpace(req.URL))
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("url must be absolute (scheme + host): %q", req.URL)
	}

	// Merge enabled query params into the URL's existing query.
	q := u.Query()
	for _, p := range req.Params {
		if p.Enabled && p.Key != "" {
			q.Add(p.Key, p.Value)
		}
	}

	body, contentType := encodeBody(req)

	// API key placed in the query string is added before we serialise the query.
	if req.Auth.Method == plugin.AuthAPIKey && strings.EqualFold(req.Auth.APIKeyIn, "query") && req.Auth.APIKeyName != "" {
		q.Set(req.Auth.APIKeyName, req.Auth.APIKeyValue)
	}
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	for _, h := range req.Headers {
		if h.Enabled && h.Key != "" {
			httpReq.Header.Add(h.Key, h.Value)
		}
	}
	if contentType != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", contentType)
	}
	applyAuth(httpReq, req.Auth)

	start := time.Now()
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	data, err := io.ReadAll(httpResp.Body)
	elapsed := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return &Response{
		Status:     httpResp.StatusCode,
		StatusText: strings.TrimPrefix(httpResp.Status, fmt.Sprintf("%d ", httpResp.StatusCode)),
		Proto:      httpResp.Proto,
		Headers:    httpResp.Header,
		Body:       data,
		Elapsed:    elapsed,
		Size:       int64(len(data)),
	}, nil
}

// encodeBody produces the request body reader and the content type implied by
// the body type (which the caller only applies if no explicit header was set).
func encodeBody(req Request) (io.Reader, string) {
	switch req.BodyType {
	case BodyJSON:
		ct := req.ContentType
		if ct == "" {
			ct = "application/json"
		}
		return strings.NewReader(req.Body), ct
	case BodyForm:
		form := url.Values{}
		for _, f := range req.FormFields {
			if f.Enabled && f.Key != "" {
				form.Add(f.Key, f.Value)
			}
		}
		return strings.NewReader(form.Encode()), "application/x-www-form-urlencoded"
	case BodyRaw:
		return strings.NewReader(req.Body), req.ContentType
	default: // BodyNone or unset
		if req.Body != "" {
			return bytes.NewReader([]byte(req.Body)), req.ContentType
		}
		return nil, ""
	}
}

// applyAuth attaches the configured authentication to the outgoing request.
// Query-string API keys are handled in Send (before the query is serialised).
func applyAuth(httpReq *http.Request, a Auth) {
	switch a.Method {
	case plugin.AuthBasic:
		httpReq.SetBasicAuth(a.Username, a.Password)
	case plugin.AuthBearer:
		if a.Token != "" {
			httpReq.Header.Set("Authorization", "Bearer "+a.Token)
		}
	case plugin.AuthAPIKey:
		if a.APIKeyName != "" && !strings.EqualFold(a.APIKeyIn, "query") {
			httpReq.Header.Set(a.APIKeyName, a.APIKeyValue)
		}
	}
}

// Descriptor is the registry record that wires REST into the shell.
func Descriptor() plugin.Descriptor {
	return plugin.Descriptor{
		Protocol:    "rest",
		DisplayName: "REST / HTTP",
		Category:    "HTTP & Web",
		New:         func() plugin.Connector { return New() },
	}
}
