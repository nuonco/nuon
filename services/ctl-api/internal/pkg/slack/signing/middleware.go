// Package signing implements verification of incoming Slack webhooks.
//
// Slack signs every request with a v0 signature derived from the request
// timestamp + body using the workspace's signing secret. Reference:
// https://api.slack.com/authentication/verifying-requests-from-slack
//
// The middleware:
//
//   - rejects requests older than maxClockSkew (Slack recommends 5 minutes;
//     5-minute window also blunts replay attacks)
//   - reads and replaces the body so downstream handlers can read it again
//   - performs a constant-time HMAC-SHA256 compare against the supplied
//     signing secret
package signing

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// SignatureHeader is the HMAC signature header sent by Slack.
	SignatureHeader = "X-Slack-Signature"
	// TimestampHeader is the request timestamp (unix seconds) sent by Slack.
	TimestampHeader = "X-Slack-Request-Timestamp"

	// signatureVersion is the prefix Slack prepends to v0 signatures.
	signatureVersion = "v0"

	// maxClockSkew bounds how far the request timestamp may drift from now.
	maxClockSkew = 5 * time.Minute
)

// Middleware returns a gin middleware that rejects unsigned or invalidly
// signed Slack requests. It must be mounted before any handler that reads the
// request body.
//
// Returns an error at construction (rather than failing per-request) when
// signingSecret is empty so a misconfigured deploy fails fast at boot rather
// than silently 500ing every Slack webhook.
func Middleware(signingSecret string) (gin.HandlerFunc, error) {
	if signingSecret == "" {
		return nil, errors.New("signing: slack signing secret is required")
	}

	handler := func(c *gin.Context) {
		ts := c.GetHeader(TimestampHeader)
		sig := c.GetHeader(SignatureHeader)
		if ts == "" || sig == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tsInt, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if drift := time.Since(time.Unix(tsInt, 0)); drift > maxClockSkew || drift < -maxClockSkew {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		// Restore body for downstream handlers.
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		if !Verify(signingSecret, ts, body, sig) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
	return handler, nil
}

// Verify returns true if the supplied signature matches the HMAC-SHA256 of
// "v0:{timestamp}:{body}" computed with signingSecret.
func Verify(signingSecret, timestamp string, body []byte, slackSig string) bool {
	base := fmt.Sprintf("%s:%s:%s", signatureVersion, timestamp, body)
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte(base))
	expected := signatureVersion + "=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(slackSig))
}
