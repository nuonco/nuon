package envelope

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

// SlackEventType for URL verification handshake requests.
const SlackURLVerificationType = "url_verification"

// SlackEvents decodes a Slack Events API envelope: either a url_verification
// handshake or an event_callback wrapping the inner event.
type SlackEvents struct{}

type slackEventEnvelope struct {
	Type      string          `json:"type"`
	Challenge string          `json:"challenge,omitempty"`
	EventID   string          `json:"event_id,omitempty"`
	EventTime int64           `json:"event_time,omitempty"`
	Event     json.RawMessage `json:"event,omitempty"`
}

func (SlackEvents) Decode(_ http.Header, body []byte) (*Event, error) {
	var envelope slackEventEnvelope
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&envelope); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return nil, errors.New("invalid Slack event envelope")
	}
	if envelope.Type == SlackURLVerificationType {
		if strings.TrimSpace(envelope.Challenge) == "" {
			return nil, errors.New("Slack URL verification request is missing challenge")
		}
		return &Event{Type: envelope.Type, Payload: json.RawMessage(body), ContentType: "application/json"}, nil
	}
	if envelope.Type != "event_callback" || strings.TrimSpace(envelope.EventID) == "" || envelope.EventTime <= 0 || len(envelope.Event) == 0 || !json.Valid(envelope.Event) {
		return nil, errors.New("invalid Slack event callback")
	}
	var inner struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(envelope.Event, &inner); err != nil || strings.TrimSpace(inner.Type) == "" {
		return nil, errors.New("Slack event callback is missing event type")
	}
	occurredAt := time.Unix(envelope.EventTime, 0).UTC()
	return &Event{ID: envelope.EventID, Type: inner.Type, OccurredAt: &occurredAt, Payload: json.RawMessage(body), ContentType: "application/json"}, nil
}

// SlackChallenge extracts the url_verification challenge from a decoded Slack
// event. It returns "" for regular event callbacks.
func SlackChallenge(event *Event) (string, error) {
	if event == nil || event.Type != SlackURLVerificationType {
		return "", nil
	}
	var envelope slackEventEnvelope
	if err := json.Unmarshal(event.Payload, &envelope); err != nil || envelope.Challenge == "" {
		return "", errors.New("Slack URL verification request is missing challenge")
	}
	return envelope.Challenge, nil
}
