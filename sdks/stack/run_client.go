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

// runClientConfig configures the run client. RunnerAPIURL and APIToken are required.
type runClientConfig struct {
	RunnerAPIURL string // runner API base URL, e.g. https://runner.nuon.co
	InstallID    string // install whose stack config is being read
	APIToken     string // bearer token; identifies the caller to the runner API
	HTTPClient   *http.Client
}

// runClient is the HTTP client for the runner API stack-run endpoints.
type runClient struct {
	cfg runClientConfig
	hc  *http.Client
}

func newRunClient(cfg runClientConfig) *runClient {
	hc := cfg.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	return &runClient{cfg: cfg, hc: hc}
}

// configResponse mirrors the runner API GET /config response
type configResponse struct {
	Config *Config `json:"config"`
}

// fetchConfig reads the rendered install-stack config for an install.
//
// Keyed on the install ID rather than a phone_home_id: the caller is authenticated
// now, so the identifier in the path is just an identifier and no longer the secret.
func (c *runClient) fetchConfig(ctx context.Context) (*Config, error) {
	url := fmt.Sprintf(
		"%s/v1/stacks/%s/config",
		strings.TrimSuffix(c.cfg.RunnerAPIURL, "/"),
		c.cfg.InstallID,
	)
	var out configResponse
	if err := c.doWithRetry(ctx, http.MethodGet, url, nil, &out); err != nil {
		return nil, err
	}
	return out.Config, nil
}

// doWithRetry retries up to 5 times with capped exponential backoff (max 8s) on
// transient failures (network errors and 5xx). 4xx errors return immediately.
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
		if c.cfg.APIToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.cfg.APIToken)
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
		// 4xx is returned immediately: a rejected credential will be rejected
		// identically on every retry, and retrying only delays the error.
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return fmt.Errorf("runner api %d: %s", resp.StatusCode, string(respBody))
		}
		lastErr = fmt.Errorf("runner api %d: %s", resp.StatusCode, string(respBody))
	}
	return fmt.Errorf("could not reach runner api at %s after %d attempts: %w", c.cfg.RunnerAPIURL, maxAttempts, lastErr)
}
