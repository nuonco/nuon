// Package signature verifies inbound trigger event authentication material:
// HMAC request signatures, Slack timestamped signatures, API keys, and basic
// auth credentials.
package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"
)

// Verifier authenticates a request against candidate shared secrets. It
// returns the index of the first candidate that authenticates the request, or
// an error describing why the request is rejected. Candidates must already be
// filtered to active secrets; expiry windows are the caller's concern.
type Verifier interface {
	Verify(headers http.Header, body []byte, secrets []string, now time.Time) (int, error)
}

// DecodeSignature strips prefix from a signature header value and decodes it
// with the named encoding ("hex" or "base64").
func DecodeSignature(value, prefix, encoding string) ([]byte, error) {
	if !strings.HasPrefix(value, prefix) {
		return nil, errors.New("invalid signature prefix")
	}
	value = strings.TrimPrefix(value, prefix)
	switch encoding {
	case "hex":
		return hex.DecodeString(value)
	case "base64":
		return base64.StdEncoding.DecodeString(value)
	default:
		return nil, errors.New("unsupported signature encoding")
	}
}

// VerifyHMAC reports whether signature is the HMAC of payload under secret
// using the named algorithm ("sha256" or "sha512").
func VerifyHMAC(secret string, payload, signature []byte, algorithm string) bool {
	var expected []byte
	switch algorithm {
	case "sha256":
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write(payload)
		expected = mac.Sum(nil)
	case "sha512":
		mac := hmac.New(sha512.New, []byte(secret))
		_, _ = mac.Write(payload)
		expected = mac.Sum(nil)
	default:
		return false
	}
	return hmac.Equal(expected, signature)
}
