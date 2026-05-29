package client

import (
	"context"
	"fmt"
)

// MonitorType matches DD's monitor type strings. Nuon's one-click
// alerts use two monitor types depending on the managed-monitor mode:
//
//   - event-v2 alert: queries DD's event stream. Used by event-mode
//     managed monitors, which require an active DD event subscription
//     so events are actually flowing into DD.
//
//   - metric alert: queries a Nuon-emitted custom metric. Used by
//     metric-mode managed monitors — Nuon evaluates the match on its
//     own side and submits `nuon.monitor.fired{nuon_monitor_id:<id>}`
//     to DD. The monitor then fires on `sum(last_5m) > 0` of that one
//     tag value. Total custom-metric cardinality stays constant at one
//     series per managed-monitor row.
type MonitorType string

const (
	MonitorTypeEventV2Alert MonitorType = "event-v2 alert"
	MonitorTypeMetricAlert  MonitorType = "metric alert"
)

// MonitorOptions matches the subset of DD's monitor `options` block we
// set. NotifyAudit / IncludeTags / RenotifyInterval defaults are picked
// to match common ops conventions; callers can override per call.
type MonitorOptions struct {
	// NotifyNoData controls whether DD pages when the stream goes silent.
	// Always false for our event-driven monitors — silence means "no
	// failures", which is the desired state.
	NotifyNoData bool `json:"notify_no_data"`

	// NewGroupDelay is the seconds to wait before alerting on a newly-
	// seen group. Zero (default) is fine for our use case.
	NewGroupDelay int `json:"new_group_delay,omitempty"`

	// IncludeTags pins the monitor's tag set so it doesn't auto-expand
	// when DD discovers new tags on the matched events. We want
	// deterministic behavior.
	IncludeTags bool `json:"include_tags"`

	// RenotifyInterval (minutes) is how often DD re-pages an unresolved
	// monitor. 0 disables renotification, which matches what we want
	// for "fired once, ack manually" workflows. Callers can override.
	RenotifyInterval int `json:"renotify_interval,omitempty"`
}

// CreateMonitorRequest is the subset of DD's POST /api/v1/monitor body
// we set. See https://docs.datadoghq.com/api/latest/monitors/#create-a-monitor.
//
// Message is plain text (with @-handles spliced in by the caller) — DD
// renders it in the alert payload.
//
// Tags here are the monitor's own tags, NOT the tags it filters events by
// (those are part of the Query string). Useful for DD UI grouping.
type CreateMonitorRequest struct {
	Name    string         `json:"name"`
	Type    MonitorType    `json:"type"`
	Query   string         `json:"query"`
	Message string         `json:"message"`
	Tags    []string       `json:"tags,omitempty"`
	Options MonitorOptions `json:"options"`
}

// Monitor is the subset of DD's monitor response we care about.
type Monitor struct {
	ID      int64       `json:"id"`
	Name    string      `json:"name"`
	Type    MonitorType `json:"type"`
	Query   string      `json:"query"`
	Message string      `json:"message"`
	Tags    []string    `json:"tags,omitempty"`
}

// CreateMonitor creates a new DD monitor under the tenant identified by
// (apiKey, appKey). The Monitors API requires both keys — apiKey alone
// returns 403.
func (c *Client) CreateMonitor(ctx context.Context, baseURL, apiKey, appKey string, req CreateMonitorRequest) (*Monitor, error) {
	if appKey == "" {
		return nil, fmt.Errorf("create datadog monitor: application key is required")
	}
	var resp Monitor
	if err := c.doJSON(ctx, "POST", baseURL, "/api/v1/monitor", apiKey, appKey, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMonitor fetches an existing monitor by ID. Used by the dashboard to
// verify that a managed monitor still exists in DD (i.e. it wasn't
// deleted from the DD UI behind our back).
func (c *Client) GetMonitor(ctx context.Context, baseURL, apiKey, appKey string, id int64) (*Monitor, error) {
	if appKey == "" {
		return nil, fmt.Errorf("get datadog monitor: application key is required")
	}
	var resp Monitor
	path := fmt.Sprintf("/api/v1/monitor/%d", id)
	if err := c.doJSON(ctx, "GET", baseURL, path, apiKey, appKey, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteMonitor removes a monitor from DD. A 404 from DD is rolled into
// the returned APIError so callers can distinguish "already gone" from
// other failures via the StatusCode field.
func (c *Client) DeleteMonitor(ctx context.Context, baseURL, apiKey, appKey string, id int64) error {
	if appKey == "" {
		return fmt.Errorf("delete datadog monitor: application key is required")
	}
	path := fmt.Sprintf("/api/v1/monitor/%d", id)
	return c.doJSON(ctx, "DELETE", baseURL, path, apiKey, appKey, nil, nil)
}
