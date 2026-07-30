package envelope

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

// AzureEventGridValidationEventType identifies the Event Grid subscription
// validation handshake event.
const AzureEventGridValidationEventType = "Microsoft.EventGrid.SubscriptionValidationEvent"

// AzureEventGrid decodes an Azure Event Grid delivery containing exactly one
// event.
type AzureEventGrid struct{}

type azureEventGridEvent struct {
	ID        string          `json:"id"`
	EventType string          `json:"eventType"`
	EventTime *time.Time      `json:"eventTime,omitempty"`
	Data      json.RawMessage `json:"data"`
}

func (AzureEventGrid) Decode(_ http.Header, body []byte) (*Event, error) {
	var events []json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&events); err != nil || decoder.Decode(&struct{}{}) != io.EOF || len(events) != 1 {
		return nil, errors.New("Azure Event Grid request must contain exactly one event")
	}
	var event azureEventGridEvent
	if err := json.Unmarshal(events[0], &event); err != nil || event.ID == "" || event.EventType == "" || len(event.Data) == 0 || !json.Valid(event.Data) || (event.EventType != AzureEventGridValidationEventType && event.EventTime == nil) {
		return nil, errors.New("invalid Azure Event Grid event")
	}
	return &Event{ID: event.ID, Type: event.EventType, OccurredAt: event.EventTime, Payload: events[0], ContentType: "application/json"}, nil
}

// AzureEventGridValidationCode extracts the subscription validation code from
// a validation handshake event. It returns "" for regular events.
func AzureEventGridValidationCode(event *Event) (string, error) {
	if event == nil || event.Type != AzureEventGridValidationEventType {
		return "", nil
	}
	var payload struct {
		Data struct {
			ValidationCode string `json:"validationCode"`
		} `json:"data"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil || payload.Data.ValidationCode == "" {
		return "", errors.New("Azure Event Grid validation event is missing validationCode")
	}
	return payload.Data.ValidationCode, nil
}
