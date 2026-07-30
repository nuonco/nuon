package envelope

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

// CloudEvents decodes a structured-mode CloudEvents 1.0 envelope. The
// envelope's own ID becomes the dedupe ID even when selectors override the
// event ID.
type CloudEvents struct{}

type cloudEvent struct {
	SpecVersion     string          `json:"specversion"`
	ID              string          `json:"id"`
	Source          string          `json:"source"`
	Type            string          `json:"type"`
	Subject         string          `json:"subject,omitempty"`
	Time            *time.Time      `json:"time,omitempty"`
	DataContentType string          `json:"datacontenttype,omitempty"`
	Data            json.RawMessage `json:"data"`
}

func (CloudEvents) Decode(_ http.Header, body []byte) (*Event, error) {
	event, err := parseCloudEvent(body)
	if err != nil {
		return nil, err
	}
	return &Event{ID: event.ID, DedupeID: event.ID, Source: event.Source, Type: event.Type, OccurredAt: event.Time, Payload: event.Data, ContentType: event.DataContentType}, nil
}

func parseCloudEvent(body []byte) (*cloudEvent, error) {
	var event cloudEvent
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&event); err != nil {
		return nil, err
	}
	if decoder.Decode(&struct{}{}) != io.EOF || event.SpecVersion != "1.0" || event.ID == "" || event.Source == "" || event.Type == "" || len(event.Data) == 0 || !json.Valid(event.Data) {
		return nil, errors.New("invalid structured CloudEvent 1.0")
	}
	return &event, nil
}
