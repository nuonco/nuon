package sns

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/nuonco/nuon/pkg/events/envelope"
)

type Decoder struct{}

func (Decoder) Decode(_ http.Header, body []byte) (*envelope.Event, error) {
	msg, err := ParseMessage(body)
	if err != nil {
		return nil, err
	}
	if msg.Type != "Notification" {
		return nil, nil
	}
	payload := json.RawMessage(msg.Message)
	if !json.Valid(payload) {
		return nil, errors.New("SNS Notification Message must contain JSON")
	}
	occurredAt, err := time.Parse(time.RFC3339Nano, msg.Timestamp)
	if err != nil {
		return nil, errors.New("invalid SNS timestamp")
	}
	return &envelope.Event{ID: msg.MessageID, OccurredAt: &occurredAt, Payload: payload, ContentType: "application/json"}, nil
}
