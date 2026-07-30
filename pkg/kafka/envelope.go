package kafka

import (
	"encoding/json"
	"fmt"
	"time"
)

const EnvelopeVersion = 1

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

// Wrap marshals payload into a versioned envelope stamped with source and type.
func Wrap(source, typ string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	return json.Marshal(Envelope{
		Version:    EnvelopeVersion,
		Type:       typ,
		ProducedAt: time.Now().UTC(),
		Source:     source,
		Payload:    raw,
	})
}

// Unwrap parses an envelope. The caller unmarshals Payload based on Type/Version.
//
// Rejects a version newer than EnvelopeVersion, not merely a different one: an
// older envelope must still pass, since a rolling deploy can have
// not-yet-restarted producers still writing it while consumers have already
// upgraded to accept a new one. Bumping EnvelopeVersion is an
// expand-then-contract change — consumers accept the new version first, then
// producers start writing it — so an exact-match check would reject the new
// version everywhere until every consumer had rolled, which defeats the
// expand step entirely.
func Unwrap(b []byte) (Envelope, error) {
	var e Envelope
	if err := json.Unmarshal(b, &e); err != nil {
		return Envelope{}, fmt.Errorf("unmarshal envelope: %w", err)
	}
	if e.Version > EnvelopeVersion {
		return Envelope{}, fmt.Errorf("envelope version %d newer than supported %d", e.Version, EnvelopeVersion)
	}
	return e, nil
}
