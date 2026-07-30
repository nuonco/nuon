// Package envelope normalizes inbound trigger event payloads. A Decoder
// understands one wire format (raw JSON, CloudEvents, Pub/Sub push, SNS,
// Azure Event Grid, Slack Events API) and produces a provider-agnostic Event.
package envelope

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nuonco/nuon/pkg/eventfilter"
)

// Event is a normalized inbound event, independent of the wire format it
// arrived in.
type Event struct {
	ID          string
	DedupeID    string
	Source      string
	Type        string
	OccurredAt  *time.Time
	Payload     json.RawMessage
	ContentType string
}

// Decoder decodes one wire format into a normalized Event. A nil Event with a
// nil error means the request was valid but carries no event to persist (for
// example an SNS subscription confirmation).
type Decoder interface {
	Decode(headers http.Header, body []byte) (*Event, error)
}

// FieldSelector extracts an event field from a request header or a JSONPath
// expression evaluated against the event payload.
type FieldSelector struct {
	Header  string `json:"header,omitempty"`
	Payload string `json:"payload,omitempty"`
}

// ValidateSelector rejects selectors that set both sources or use an invalid
// payload path.
func ValidateSelector(selector FieldSelector) error {
	if selector.Header != "" && selector.Payload != "" {
		return errors.New("exactly one of header or payload may be set")
	}
	if selector.Payload != "" {
		if _, err := eventfilter.ParsePath(selector.Payload, false); err != nil {
			return fmt.Errorf("invalid payload selector: %w", err)
		}
	}
	return nil
}

// ApplySelectors overrides the event ID and type from the configured
// selectors. Header selectors only apply when the header is present; payload
// selectors must match exactly one nonempty string.
func ApplySelectors(event *Event, headers http.Header, typeFrom, idFrom FieldSelector) error {
	if idFrom.Header != "" {
		if value := headers.Get(idFrom.Header); value != "" {
			event.ID = value
		}
	}
	if typeFrom.Header != "" {
		if value := headers.Get(typeFrom.Header); value != "" {
			event.Type = value
		}
	}
	if idFrom.Payload == "" && typeFrom.Payload == "" {
		return nil
	}
	payload, err := DecodeJSON(event.Payload)
	if err != nil {
		return err
	}
	if idFrom.Payload != "" {
		event.ID, err = selectString(payload, idFrom.Payload)
		if err != nil {
			return fmt.Errorf("extract event ID: %w", err)
		}
	}
	if typeFrom.Payload != "" {
		event.Type, err = selectString(payload, typeFrom.Payload)
		if err != nil {
			return fmt.Errorf("extract event type: %w", err)
		}
	}
	return nil
}

// DecodeJSON decodes a JSON document preserving number precision.
func DecodeJSON(body []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var payload any
	if err := decoder.Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func selectString(payload any, pathValue string) (string, error) {
	path, err := eventfilter.ParsePath(pathValue, false)
	if err != nil {
		return "", err
	}
	selected := path.Select(payload)
	if len(selected) != 1 {
		return "", fmt.Errorf("selector matched %d values", len(selected))
	}
	value, ok := selected[0].(string)
	if !ok || value == "" {
		return "", errors.New("selector must match a nonempty string")
	}
	return value, nil
}
