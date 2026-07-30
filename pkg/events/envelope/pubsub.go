package envelope

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// PubSubPush decodes a Google Pub/Sub push delivery envelope.
type PubSubPush struct{}

type pubSubPushEnvelope struct {
	Message struct {
		Data        string `json:"data"`
		MessageID   string `json:"messageId"`
		PublishTime string `json:"publishTime"`
	} `json:"message"`
}

func (PubSubPush) Decode(_ http.Header, body []byte) (*Event, error) {
	var push pubSubPushEnvelope
	if err := json.Unmarshal(body, &push); err != nil || push.Message.Data == "" || push.Message.MessageID == "" {
		return nil, errors.New("invalid Pub/Sub push envelope")
	}
	payload, err := base64.StdEncoding.DecodeString(push.Message.Data)
	if err != nil || !json.Valid(payload) {
		return nil, errors.New("invalid Pub/Sub message data")
	}
	var occurredAt *time.Time
	if push.Message.PublishTime != "" {
		parsed, err := time.Parse(time.RFC3339Nano, push.Message.PublishTime)
		if err != nil {
			return nil, errors.New("invalid Pub/Sub publish time")
		}
		occurredAt = &parsed
	}
	return &Event{ID: push.Message.MessageID, OccurredAt: occurredAt, Payload: payload, ContentType: "application/json"}, nil
}
