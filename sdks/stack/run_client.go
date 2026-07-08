package stack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// runClientConfig configures the run client. CtlAPIURL + PhoneHomeID are required.
type runClientConfig struct {
	CtlAPIURL   string // base URL, e.g. https://api.nuon.co
	PhoneHomeID string // per-stack-version secret, in the URL path
	HTTPClient  *http.Client
}

// runClient is the small HTTP client for the ctl-api stack-run endpoints. The
// endpoints are public and mirror the phone-home pattern: the per-stack-version
// phone_home_id sits in the URL path as the secret. No Authorization header, no
// Nuon API token.
type runClient struct {
	cfg runClientConfig
	hc  *http.Client
}

func newRunClient(cfg runClientConfig) *runClient {
	hc := cfg.HTTPClient
	if hc == nil {
		// Per-attempt timeout is intentionally short. doWithRetry retries up
		// to 5 times with capped backoff, so total wall time on a hard outage
		// is ~60s — long enough to ride out a brief network blip, short
		// enough to fail fast when ctl-api is genuinely unreachable.
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	return &runClient{cfg: cfg, hc: hc}
}

// configResponse mirrors the ctl-api GET /config response: just the rendered
// config block, no run. Used by FetchConfig for read-only callers (the
// Terraform provider's nuon_stack data source).
type configResponse struct {
	Config *Config `json:"config"`
}

// fetchConfig reads the rendered install-stack config for the stack version
// without creating a run or mutating any state (GET, side-effect free).
func (c *runClient) fetchConfig(ctx context.Context) (*Config, error) {
	url := fmt.Sprintf(
		"%s/v1/stack-runs/%s/config",
		strings.TrimSuffix(c.cfg.CtlAPIURL, "/"),
		c.cfg.PhoneHomeID,
	)
	var out configResponse
	if err := c.doWithRetry(ctx, http.MethodGet, url, nil, &out); err != nil {
		return nil, err
	}
	return out.Config, nil
}

// doWithRetry retries up to 5 times with capped exponential backoff on
// transient failures (network errors and 5xx). 4xx errors are returned
// immediately. The backoff caps at 8s so total wall time on a hard outage
// stays bounded.
func (c *runClient) doWithRetry(ctx context.Context, method, url string, body, out any) error {
	const maxAttempts = 5
	const maxDelay = 8 * time.Second
	var lastErr error
	delay := 500 * time.Millisecond
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			if delay *= 2; delay > maxDelay {
				delay = maxDelay
			}
		}

		var bodyReader io.Reader
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("marshal body: %w", err)
			}
			bodyReader = bytes.NewReader(b)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.hc.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out != nil && len(respBody) > 0 {
				if err := json.Unmarshal(respBody, out); err != nil {
					return fmt.Errorf("decode response: %w", err)
				}
			}
			return nil
		}
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return fmt.Errorf("ctl-api %d: %s", resp.StatusCode, string(respBody))
		}
		lastErr = fmt.Errorf("ctl-api %d: %s", resp.StatusCode, string(respBody))
	}
	return fmt.Errorf("could not reach ctl-api at %s after %d attempts: %w", c.cfg.CtlAPIURL, maxAttempts, lastErr)
}
