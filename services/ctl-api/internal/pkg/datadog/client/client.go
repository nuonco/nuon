// Package client is a thin Datadog REST API client scoped to the surface
// area the Nuon DD integration needs:
//
//   - POST /api/v1/events      — emit an event into the DD event stream
//   - GET  /api/v1/validate    — probe an API key (used by the Test button)
//   - POST /api/v1/monitor     — create a monitor (managed-monitor feature)
//   - DELETE /api/v1/monitor/{id} — delete a monitor
//   - GET  /api/v1/monitor/{id} — fetch a monitor (used to verify state)
//
// The surface is small enough that pulling in datadog-api-client-go (which
// is generated, large, and has its own retry semantics) is overkill —
// hand-rolling keeps the contract explicit and matches the pattern set by
// internal/pkg/slack/client.
//
// Auth: every method takes the API key (and where required, the app key)
// as parameters because the credentials live in app.DatadogConnection rows
// rather than on a process-wide singleton. The HTTP client itself is
// stateless and shared across all connections.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// defaultTimeout caps individual DD API requests. DD typically responds
// in well under a second; cap conservatively to keep activity execution
// from blocking on a misbehaving tenant.
const defaultTimeout = 10 * time.Second

// Client is a Datadog REST client. Stateless — construct once and share.
type Client struct {
	httpClient *http.Client
}

// Option mutates Client during construction.
type Option func(*Client)

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

// New constructs a DD client with sensible defaults.
func New(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// doJSON is the shared request path. It marshals `body` (if non-nil),
// sets the standard DD auth + content headers, executes the request, and
// — on a 2xx — decodes the response body into `out` if non-nil.
//
// `appKey` is optional; pass "" to skip the DD-APPLICATION-KEY header
// (Events API doesn't need it).
func (c *Client) doJSON(
	ctx context.Context,
	method, baseURL, path, apiKey, appKey string,
	body, out any,
) error {
	url := baseURL + path

	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal datadog request: %w", err)
		}
		reqBody = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("build datadog request: %w", err)
	}
	req.Header.Set("DD-API-KEY", apiKey)
	if appKey != "" {
		req.Header.Set("DD-APPLICATION-KEY", appKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("datadog request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the body up-front; we need it for either decoding or
	// surfacing an error message.
	respBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("read datadog response: %w", readErr)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// DD error responses are JSON-shaped {"errors": ["msg"]} on most
		// endpoints. We surface the raw bytes (truncated) so callers can
		// log them; structured decoding is overkill for an unhappy path.
		snippet := string(respBytes)
		const maxSnippet = 512
		if len(snippet) > maxSnippet {
			snippet = snippet[:maxSnippet] + "...(truncated)"
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Body:       snippet,
		}
	}

	if out == nil || len(respBytes) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBytes, out); err != nil {
		return fmt.Errorf("decode datadog response: %w", err)
	}
	return nil
}

// APIError is returned for non-2xx responses. Callers can type-assert to
// branch on StatusCode (e.g., 403 → invalid credentials, 404 → missing
// monitor).
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("datadog api error: status=%d body=%s", e.StatusCode, e.Body)
}
