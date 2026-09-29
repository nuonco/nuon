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

type Verifier interface {
	Verify(headers http.Header, body []byte, secrets []string, now time.Time) (int, error)
}

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
