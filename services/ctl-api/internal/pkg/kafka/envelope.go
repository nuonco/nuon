package kafka

import (
	"encoding/json"
	"fmt"
	"time"
)

// EnvelopeVersion is the current schema version stamped on outgoing messages.
const EnvelopeVersion = 1

const envelopeSource = "ctl-api"

// Message type discriminators carried in the envelope.
const (
	TypeRunnerHeartBeat = "runner_heart_beat"
	// TypeOtelLogRecord is defined for the logs stream, which is not wired yet.
	TypeOtelLogRecord = "otel_log_record"
)

// Envelope is the versioned wrapper every Kafka message uses. Payload is left as
// raw JSON so consumers decode it by Type/Version rather than the producer and
// consumer sharing a compiled payload type.
type Envelope struct {
	Version    int             `json:"version"`
	Type       string          `json:"type"`
	ProducedAt time.Time       `json:"produced_at"`
	Source     string          `json:"source"`
	Payload    json.RawMessage `json:"payload"`
}

// Wrap marshals payload and returns the JSON-encoded envelope ready to produce.
func Wrap(typ string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	return json.Marshal(Envelope{
		Version:    EnvelopeVersion,
		Type:       typ,
		ProducedAt: time.Now().UTC(),
		Source:     envelopeSource,
		Payload:    raw,
	})
}

// Unwrap parses an envelope. The caller unmarshals Payload based on Type/Version.
func Unwrap(b []byte) (Envelope, error) {
	var e Envelope
	if err := json.Unmarshal(b, &e); err != nil {
		return Envelope{}, fmt.Errorf("unmarshal envelope: %w", err)
	}
	return e, nil
}
