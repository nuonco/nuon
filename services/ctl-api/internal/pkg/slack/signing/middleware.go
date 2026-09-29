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
	SignatureHeader = "X-Slack-Signature"
	TimestampHeader = "X-Slack-Request-Timestamp"

	signatureVersion = "v0"

	maxClockSkew = 5 * time.Minute
)

// why: Middleware returns a gin middleware that rejects unsigned or invalidly
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
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		if !Verify(signingSecret, ts, body, sig) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
	return handler, nil
}

func Verify(signingSecret, timestamp string, body []byte, slackSig string) bool {
	base := fmt.Sprintf("%s:%s:%s", signatureVersion, timestamp, body)
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte(base))
	expected := signatureVersion + "=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(slackSig))
}
