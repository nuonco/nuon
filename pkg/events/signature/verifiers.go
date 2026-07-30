package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// SlackSignatureHeader is the HMAC signature header sent by Slack.
	SlackSignatureHeader = "X-Slack-Signature"
	// SlackTimestampHeader is the request timestamp (unix seconds) sent by Slack.
	SlackTimestampHeader = "X-Slack-Request-Timestamp"

	slackSignatureVersion    = "v0"
	slackRequestMaxClockSkew = 5 * time.Minute
)

// HMAC verifies a request body signature carried in a header.
type HMAC struct {
	Header    string
	Prefix    string
	Algorithm string
	Encoding  string
}

func (h HMAC) Verify(headers http.Header, body []byte, secrets []string, _ time.Time) (int, error) {
	signature, err := DecodeSignature(headers.Get(h.Header), h.Prefix, h.Encoding)
	if err != nil {
		return -1, errors.New("invalid signature header")
	}
	for i, secret := range secrets {
		if VerifyHMAC(secret, body, signature, h.Algorithm) {
			return i, nil
		}
	}
	return -1, errors.New("invalid signature")
}

// Slack verifies Slack's timestamp-bound v0 request signature. Reference:
// https://api.slack.com/authentication/verifying-requests-from-slack
type Slack struct{}

func (Slack) Verify(headers http.Header, body []byte, secrets []string, now time.Time) (int, error) {
	timestamp := headers.Get(SlackTimestampHeader)
	provided := headers.Get(SlackSignatureHeader)
	if !ValidSlackRequestTimestamp(timestamp, now) {
		return -1, errors.New("invalid Slack request timestamp")
	}
	for i, secret := range secrets {
		if VerifySlack(secret, timestamp, body, provided) {
			return i, nil
		}
	}
	return -1, errors.New("invalid Slack signature")
}

// ValidSlackRequestTimestamp reports whether value is a unix timestamp within
// Slack's recommended 5-minute replay window of now.
func ValidSlackRequestTimestamp(value string, now time.Time) bool {
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false
	}
	drift := now.Sub(time.Unix(seconds, 0))
	return drift <= slackRequestMaxClockSkew && drift >= -slackRequestMaxClockSkew
}

// VerifySlack reports whether slackSig matches the HMAC-SHA256 of
// "v0:{timestamp}:{body}" computed with signingSecret.
func VerifySlack(signingSecret, timestamp string, body []byte, slackSig string) bool {
	base := fmt.Sprintf("%s:%s:%s", slackSignatureVersion, timestamp, body)
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte(base))
	expected := slackSignatureVersion + "=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(slackSig))
}

// APIKey verifies a shared key carried in a header, optionally behind a
// prefix such as "Bearer ".
type APIKey struct {
	Header string
	Prefix string
}

func (k APIKey) Verify(headers http.Header, _ []byte, secrets []string, _ time.Time) (int, error) {
	value := headers.Get(k.Header)
	if !strings.HasPrefix(value, k.Prefix) {
		return -1, errors.New("invalid API key")
	}
	value = strings.TrimPrefix(value, k.Prefix)
	for i, secret := range secrets {
		if subtle.ConstantTimeCompare([]byte(value), []byte(secret)) == 1 {
			return i, nil
		}
	}
	return -1, errors.New("invalid API key")
}

// Basic verifies HTTP basic authentication with a fixed username and the
// password matched against candidate secrets.
type Basic struct {
	Username string
}

func (b Basic) Verify(headers http.Header, _ []byte, secrets []string, _ time.Time) (int, error) {
	request := http.Request{Header: headers}
	username, password, ok := request.BasicAuth()
	if !ok || subtle.ConstantTimeCompare([]byte(username), []byte(b.Username)) != 1 {
		return -1, errors.New("invalid basic authentication")
	}
	for i, secret := range secrets {
		if subtle.ConstantTimeCompare([]byte(password), []byte(secret)) == 1 {
			return i, nil
		}
	}
	return -1, errors.New("invalid basic authentication")
}
