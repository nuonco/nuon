// Package stackrun is a small HTTP client for the ctl-api stack-run
// endpoints used by the AWS-native SDK provisioner.
//
// The endpoints are public and mirror the phone-home pattern: the per-
// stack-version phone_home_id sits in the URL path as the secret. No
// Authorization header, no Nuon API token.
package stackrun

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-installer-go/stack"
)

// Config configures the client. All fields are required.
type Config struct {
	CtlAPIURL   string // base URL, e.g. https://api.nuon.co
	PhoneHomeID string // per-stack-version secret, in the URL path
	HTTPClient  *http.Client
}

type Client struct {
	cfg Config
	hc  *http.Client
}

func New(cfg Config) *Client {
	hc := cfg.HTTPClient
	if hc == nil {
		// Per-attempt timeout is intentionally short. doWithRetry retries up
		// to 5 times with capped backoff, so total wall time on a hard outage
		// is ~60s — long enough to ride out a brief network blip, short
		// enough to fail the CLI fast when ctl-api is genuinely unreachable.
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{cfg: cfg, hc: hc}
}

// LogStream mirrors the transient app.LogStream fields ctl-api returns on
// the create-run response. The SDK uses these to push OTLP logs back.
type LogStream struct {
	ID           string `json:"id"`
	WriteToken   string `json:"write_token"`
	RunnerAPIURL string `json:"runner_api_url"`
}

// CreateRunResponse mirrors the ctl-api response shape. Config is the
// rendered tfvars-equivalent that the SDK threads into the stack package;
// ctl-api must populate it on every successful create-run, otherwise the
// SDK has no idea what permissions/secrets/roles to provision.
type CreateRunResponse struct {
	ID                    string        `json:"id"`
	InstallStackVersionID string        `json:"install_stack_version_id"`
	LogStream             *LogStream    `json:"log_stream,omitempty"`
	Config                *stack.Config `json:"config,omitempty"`
}

// RunKind identifies the operation a stack run represents. Mirrors the
// ctl-api `app.InstallStackVersionRunKind` enum.
type RunKind string

const (
	RunKindProvision   RunKind = "provision"
	RunKindReprovision RunKind = "reprovision"
	RunKindDeprovision RunKind = "deprovision"
)

// CreateRun starts a new stack version run of the given kind. Returns the
// full response so the caller can read the run ID + log-stream credentials.
//
// kind is appended to the URL as a /kind/{kind} segment so it can't shadow
// the PATCH terminal-update route (which uses /:run_id at the same depth).
// An empty kind defaults to provision.
func (c *Client) CreateRun(ctx context.Context, kind RunKind) (*CreateRunResponse, error) {
	if kind == "" {
		kind = RunKindProvision
	}
	url := fmt.Sprintf(
		"%s/v1/stack-runs/%s/kind/%s",
		strings.TrimSuffix(c.cfg.CtlAPIURL, "/"),
		c.cfg.PhoneHomeID,
		kind,
	)
	var out CreateRunResponse
	if err := c.doWithRetry(ctx, http.MethodPost, url, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateRunRequest is the PATCH body. Status must be "succeeded" or "failed".
type UpdateRunRequest struct {
	Status            string         `json:"status"`
	StatusDescription string         `json:"status_description,omitempty"`
	Data              map[string]any `json:"data,omitempty"`
}

// UpdateRun marks a run as succeeded or failed and includes terminal data.
func (c *Client) UpdateRun(ctx context.Context, runID string, req UpdateRunRequest) error {
	url := fmt.Sprintf(
		"%s/v1/stack-runs/%s/%s",
		strings.TrimSuffix(c.cfg.CtlAPIURL, "/"),
		c.cfg.PhoneHomeID,
		runID,
	)
	return c.doWithRetry(ctx, http.MethodPatch, url, req, nil)
}

// doWithRetry retries up to 5 times with capped exponential backoff on
// transient failures (network errors and 5xx). 4xx errors are returned
// immediately. The backoff caps at 8s so total wall time on a hard outage
// stays bounded.
func (c *Client) doWithRetry(ctx context.Context, method, url string, body, out any) error {
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
