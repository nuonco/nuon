package envelope

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
)

// Raw decodes an unwrapped JSON event. The event ID defaults to the body
// digest so retried deliveries dedupe; selectors may override it.
type Raw struct{}

func (Raw) Decode(headers http.Header, body []byte) (*Event, error) {
	if !json.Valid(body) {
		return nil, errors.New("invalid JSON event")
	}
	sum := sha256.Sum256(body)
	return &Event{ID: hex.EncodeToString(sum[:]), Payload: json.RawMessage(body), ContentType: headers.Get("Content-Type")}, nil
}
