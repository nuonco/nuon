package posthog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

const (
	Lib = "ctl-api"

	defaultTimeout = time.Second * 5
)

type Event struct {
	Event      string         `json:"event"`
	DistinctID string         `json:"distinct_id"`
	Properties map[string]any `json:"properties,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

type Client interface {
	Enabled() bool
	Capture(ctx context.Context, events ...Event) error
}

type client struct {
	key  string
	host string
}

var _ Client = (*client)(nil)

func New(cfg *internal.Config) Client {
	return &client{
		key:  cfg.PostHogKey,
		host: strings.TrimSuffix(cfg.PostHogHost, "/"),
	}
}

func (c *client) Enabled() bool {
	return c.key != "" && c.host != ""
}

func (c *client) Capture(ctx context.Context, events ...Event) error {
	if !c.Enabled() || len(events) == 0 {
		return nil
	}

	for i := range events {
		if events[i].Properties == nil {
			events[i].Properties = map[string]any{}
		}
		events[i].Properties["$lib"] = Lib
	}

	byts, err := json.Marshal(map[string]any{
		"api_key": c.key,
		"batch":   events,
	})
	if err != nil {
		return fmt.Errorf("unable to marshal events: %w", err)
	}

	timeoutCtx, cancelFn := context.WithTimeout(ctx, defaultTimeout)
	defer cancelFn()

	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodPost, c.host+"/batch/", bytes.NewReader(byts))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send events: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("posthog request failed with status %d: %s", res.StatusCode, string(body))
	}

	return nil
}
