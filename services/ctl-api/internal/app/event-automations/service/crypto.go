package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"
)

const signatureSkew = 5 * time.Minute

var errStaleTimestamp = errors.New("stale timestamp")

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

func parseTimestamp(value string, now time.Time) (string, error) {
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil || value == "" {
		return "", errors.New("invalid timestamp")
	}
	timestamp := time.Unix(seconds, 0)
	if now.Sub(timestamp).Abs() > signatureSkew {
		return "", errStaleTimestamp
	}
	return value, nil
}

func parseSignature(value string) ([]byte, error) {
	if !strings.HasPrefix(value, "v1=") || len(value) != 67 {
		return nil, errors.New("invalid signature")
	}
	hexValue := strings.TrimPrefix(value, "v1=")
	if hexValue != strings.ToLower(hexValue) {
		return nil, errors.New("invalid signature")
	}
	decoded, err := hex.DecodeString(hexValue)
	if err != nil || len(decoded) != sha256.Size {
		return nil, errors.New("invalid signature")
	}
	return decoded, nil
}

func verifySignature(secret, timestamp string, body, signature []byte) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "."))
	_, _ = mac.Write(body)
	return hmac.Equal(mac.Sum(nil), signature)
}
