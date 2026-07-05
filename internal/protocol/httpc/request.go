// Package httpc implements Conduit's HTTP-family connectors: REST (this file
// and rest.go) and, alongside it, WebSocket/SSE/GraphQL. The REST connector is
// the flagship protocol view and exercises the shared view→connector→history→env
// pipeline end to end.
package httpc

import (
	"net/http"
	"time"

	"github.com/geekpratyush/conduit/internal/plugin"
)

// KeyValue is an enable-able key/value pair used for query params and headers,
// mirroring the row model a REST view edits in its Params/Headers tabs.
type KeyValue struct {
	Key     string
	Value   string
	Enabled bool
}

// BodyType selects how a request body is encoded.
type BodyType string

const (
	BodyNone BodyType = "none"
	BodyRaw  BodyType = "raw"  // send Body verbatim with ContentType
	BodyJSON BodyType = "json" // Body is JSON; ContentType defaults to application/json
	BodyForm BodyType = "form" // Body rows are URL-encoded as x-www-form-urlencoded
)

// Auth captures a request's authentication settings. Only the fields relevant
// to Method are consulted. Additional methods (OAuth2, SigV4, Digest, HMAC) are
// layered on in later phases.
type Auth struct {
	Method plugin.AuthMethod

	// Basic
	Username string
	Password string

	// Bearer
	Token string

	// API key
	APIKeyName  string
	APIKeyValue string
	APIKeyIn    string // "header" (default) or "query"
}

// Request is a protocol-agnostic-of-transport description of one REST call,
// built by the view and executed by RESTConnector.Send. All ${VAR} expansion is
// expected to have happened before Send (the view resolves via EnvironmentService).
type Request struct {
	Method      string
	URL         string
	Params      []KeyValue
	Headers     []KeyValue
	BodyType    BodyType
	Body        string
	ContentType string
	FormFields  []KeyValue
	Auth        Auth
}

// Response is the outcome of a REST call, including timing and size so the view
// can render the status pill, latency chip, and size chip.
type Response struct {
	Status     int
	StatusText string
	Proto      string
	Headers    http.Header
	Body       []byte
	Elapsed    time.Duration
	Size       int64
}
