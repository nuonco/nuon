package client

import (
	"context"
	"time"
)

// EventAlertType mirrors DD's accepted alert_type values. Stored as a
// string so we never reject something DD actually accepts; the model
// validates against a known set at create time.
type EventAlertType string

const (
	EventAlertTypeInfo    EventAlertType = "info"
	EventAlertTypeWarning EventAlertType = "warning"
	EventAlertTypeError   EventAlertType = "error"
	EventAlertTypeSuccess EventAlertType = "success"
)

// EventPriority mirrors DD's accepted priority values.
type EventPriority string

const (
	EventPriorityNormal EventPriority = "normal"
	EventPriorityLow    EventPriority = "low"
)

// PostEventRequest is the subset of DD's /api/v1/events payload we use.
// See https://docs.datadoghq.com/api/latest/events/#post-an-event.
//
// SourceTypeName is always "nuon" for events we emit — DD treats this as
// the event's `source:` tag, which is what monitor queries filter on
// (`events("source:nuon ...")`).
//
// AggregationKey is set to the Nuon workflow ID so DD groups all step
// events under their parent workflow in the event stream UI. This is the
// closest analog DD has to Slack's threaded posts.
//
// DateHappened is the Unix epoch seconds of the event; omitted to default
// to "now" on DD's side, which matches the wall-clock of when the hook
// fired.
type PostEventRequest struct {
	Title          string         `json:"title"`
	Text           string         `json:"text"`
	Tags           []string       `json:"tags,omitempty"`
	AlertType      EventAlertType `json:"alert_type,omitempty"`
	Priority       EventPriority  `json:"priority,omitempty"`
	AggregationKey string         `json:"aggregation_key,omitempty"`
	SourceTypeName string         `json:"source_type_name,omitempty"`
	DateHappened   int64          `json:"date_happened,omitempty"`
}

// PostEventResponse captures the fields we care about after a successful
// emit. DD returns a much larger payload; we keep only the bits useful
// for logging / deep-linking.
type PostEventResponse struct {
	Status string `json:"status"`
	Event  struct {
		ID  int64  `json:"id"`
		URL string `json:"url"`
	} `json:"event"`
}

// PostEvent emits an event into the DD event stream for the tenant
// identified by apiKey. SourceTypeName defaults to "nuon" if unset by the
// caller — we only ever post under that source so monitor queries can
// reliably filter on `source:nuon`.
//
// The Events API only requires DD-API-KEY; DD-APPLICATION-KEY is not
// passed.
func (c *Client) PostEvent(ctx context.Context, baseURL, apiKey string, req PostEventRequest) (*PostEventResponse, error) {
	if req.SourceTypeName == "" {
		req.SourceTypeName = "nuon"
	}
	if req.DateHappened == 0 {
		req.DateHappened = time.Now().Unix()
	}
	var resp PostEventResponse
	if err := c.doJSON(ctx, "POST", baseURL, "/api/v1/events", apiKey, "", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
