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

type Event struct {
	ID          string
	DedupeID    string
	Source      string
	Type        string
	OccurredAt  *time.Time
	Payload     json.RawMessage
	ContentType string
}

type Decoder interface {
	Decode(headers http.Header, body []byte) (*Event, error)
}

type FieldSelector struct {
	Header  string `json:"header,omitempty"`
	Payload string `json:"payload,omitempty"`
}

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
