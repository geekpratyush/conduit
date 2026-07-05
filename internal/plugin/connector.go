// Package plugin defines the protocol-connector SPI that every protocol in
// Conduit implements: a small, dependency-free contract that the shell, vault,
// history, and env systems reuse so that adding a protocol is a well-defined
// unit of work (a package implementing Connector + a UI view).
package plugin

import (
	"context"
	"time"
)

// AuthMethod enumerates the ways a connection can authenticate. Individual
// protocols use the subset that applies to them.
type AuthMethod string

const (
	AuthNone     AuthMethod = "none"
	AuthBasic    AuthMethod = "basic"
	AuthBearer   AuthMethod = "bearer"
	AuthAPIKey   AuthMethod = "apikey"
	AuthOAuth2   AuthMethod = "oauth2"
	AuthAWSSigV4 AuthMethod = "aws-sigv4"
	AuthDigest   AuthMethod = "digest"
	AuthHMAC     AuthMethod = "hmac"
	AuthMTLS     AuthMethod = "mtls"
)

// ConnectionConfig is the protocol-agnostic description of a connection target.
// Protocol-specific settings live in Extras (free-form) so the shared shell,
// profile store, and vault can treat every connection uniformly. Secret values
// are never stored here in plaintext once saved — the profile store replaces
// them with vault references (see internal/security).
type ConnectionConfig struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Protocol string            `json:"protocol"` // e.g. "rest", "kafka", "sftp"
	Host     string            `json:"host,omitempty"`
	Port     int               `json:"port,omitempty"`
	URL      string            `json:"url,omitempty"`
	Username string            `json:"username,omitempty"`
	Secret   string            `json:"secret,omitempty"` // vault ref or resolved secret
	Auth     AuthMethod        `json:"auth,omitempty"`
	UseTLS   bool              `json:"useTLS,omitempty"`
	Extras   map[string]string `json:"extras,omitempty"`
}

// Extra returns the Extras value for key, or the empty string if absent.
func (c ConnectionConfig) Extra(key string) string {
	if c.Extras == nil {
		return ""
	}
	return c.Extras[key]
}

// TestResult reports the outcome of a connectivity probe.
type TestResult struct {
	OK      bool
	Message string
	Latency time.Duration
}

// Connector is the SPI every protocol backend implements. It intentionally
// stays minimal: connectivity plus metadata. Protocol-specific operations
// (send a request, produce a message, run a query, list a bucket) live on the
// concrete connector types and are surfaced by that protocol's view — the shell
// only depends on this common surface.
type Connector interface {
	// Protocol is the stable identifier for this connector kind (e.g. "rest").
	Protocol() string
	// DisplayName is the human-facing label (e.g. "REST / HTTP").
	DisplayName() string
	// Connect establishes/validates a session for cfg. Implementations should
	// honour ctx cancellation and deadlines.
	Connect(ctx context.Context, cfg ConnectionConfig) error
	// Test performs a lightweight connectivity probe without full setup.
	Test(ctx context.Context, cfg ConnectionConfig) TestResult
	// Close releases any resources held by a prior Connect.
	Close() error
}
