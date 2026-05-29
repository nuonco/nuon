package client

import (
	"context"
	"fmt"
)

// MonitorFiredMetric is the DD custom metric Nuon submits when a
// metric-mode managed monitor matches a lifecycle event. The single tag
// `nuon_monitor_id:<row_id>` is what the DD-side monitor query keys
// off, so this constant is the contract between two layers: the
// dashboard's monitor-create flow (uses it inside the metric-alert
// query template) and the signal lifecycle hook (uses it as the
// submitted metric name). Anchored here in the wire layer so neither
// layer has to import the other.
const MonitorFiredMetric = "nuon.monitor.fired"

// MetricType matches DD's metric type strings used in the v1 series API.
// Nuon's monitor-fired metric is a count, which DD expects as "count" so
// the monitor's sum(last_5m) aggregation behaves the way users intuit it
// (each submitted point increments the firing count by one).
type MetricType string

const (
	MetricTypeCount MetricType = "count"
)

// MetricPoint is one [timestamp, value] sample in DD's series shape.
// Timestamp is a unix epoch in seconds — DD rejects sub-second precision
// on the v1 series API.
type MetricPoint struct {
	Timestamp int64
	Value     float64
}

// MetricSeries is the subset of DD's series body we submit. Type=count
// is currently the only callsite, but the field is parameterized so a
// future gauge submission (e.g. backlog depth) can reuse the same
// path.
//
// Tags here are the metric's tags, NOT monitor tags — DD multiplexes
// the metric into one series per unique tag set. We keep this list
// tiny (just `nuon_monitor_id`) for the monitor-fired metric so the
// custom-metric cardinality footprint stays at one series per
// managed-monitor row regardless of how many installs / labels match.
type MetricSeries struct {
	Metric string
	Type   MetricType
	Tags   []string
	Points []MetricPoint
}

// postSeriesRequest is DD's POST /api/v1/series body shape. Wrapped here
// rather than exposed publicly because the v1 series API uses a tuple-
// shaped points field ([[ts, val], ...]) that we serialize from the
// nicer MetricPoint struct above.
type postSeriesRequest struct {
	Series []postSeriesRequestSeries `json:"series"`
}

type postSeriesRequestSeries struct {
	Metric string       `json:"metric"`
	Type   MetricType   `json:"type"`
	Tags   []string     `json:"tags,omitempty"`
	Points [][2]float64 `json:"points"`
}

// PostSeries submits one or more metric series to DD's v1 series API.
// Unlike monitor calls, the series API only requires the API key — the
// application key isn't part of the auth contract.
//
// Submissions are best-effort from the caller's perspective: DD returns
// 202 on accepted, and rejected points are surfaced via APIError so the
// caller can log + drop without retrying (DD already absorbs short bursts
// internally; aggressive retries from us would amplify noise during
// outages).
func (c *Client) PostSeries(ctx context.Context, baseURL, apiKey string, series []MetricSeries) error {
	if len(series) == 0 {
		return nil
	}
	body := postSeriesRequest{
		Series: make([]postSeriesRequestSeries, 0, len(series)),
	}
	for _, s := range series {
		if s.Metric == "" {
			return fmt.Errorf("post datadog series: metric name is required")
		}
		if len(s.Points) == 0 {
			continue
		}
		points := make([][2]float64, 0, len(s.Points))
		for _, p := range s.Points {
			points = append(points, [2]float64{float64(p.Timestamp), p.Value})
		}
		body.Series = append(body.Series, postSeriesRequestSeries{
			Metric: s.Metric,
			Type:   s.Type,
			Tags:   s.Tags,
			Points: points,
		})
	}
	if len(body.Series) == 0 {
		return nil
	}
	return c.doJSON(ctx, "POST", baseURL, "/api/v1/series", apiKey, "", body, nil)
}
