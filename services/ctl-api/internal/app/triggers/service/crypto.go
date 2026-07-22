package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func generateCredential() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashIngressKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func decodeHMACSignature(value string, config app.TriggerAuthConfig) ([]byte, error) {
	if !strings.HasPrefix(value, config.Prefix) {
		return nil, errors.New("invalid signature prefix")
	}
	value = strings.TrimPrefix(value, config.Prefix)
	switch config.Encoding {
	case "hex":
		return hex.DecodeString(value)
	case "base64":
		return base64.StdEncoding.DecodeString(value)
	default:
		return nil, errors.New("unsupported signature encoding")
	}
}

func verifyGenericHMAC(secret string, payload, signature []byte, algorithm string) bool {
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
