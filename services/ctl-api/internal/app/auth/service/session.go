package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SessionData holds the OAuth flow state in a signed cookie.
type SessionData struct {
	State        string `json:"state"`
	ProviderID   string `json:"pid,omitempty"`
	RequestedURL string `json:"url,omitempty"`
	FailCount    int    `json:"fc,omitempty"`
	CreatedAt    int64  `json:"iat"`
}

var (
	errInvalidSessionFormat    = errors.New("invalid session cookie format")
	errInvalidSessionSignature = errors.New("invalid session cookie signature")
	errSessionExpired          = errors.New("session cookie expired")
)

// sessionCookieMaxAge is how long the session cookie is valid (5 minutes for OAuth flow).
const sessionCookieMaxAge = 5 * 60

// getSession retrieves and validates the session data from the cookie.
func (s *service) getSession(c *gin.Context) (*SessionData, error) {
	cookie, err := c.Request.Cookie(NuonAuthSessionName)
	if err != nil {
		return nil, err
	}

	return s.decodeSession(cookie.Value)
}

// setSession creates a signed session cookie with the given data.
func (s *service) setSession(c *gin.Context, data *SessionData) error {
	// Set creation time if not already set
	if data.CreatedAt == 0 {
		data.CreatedAt = time.Now().Unix()
	}

	encoded, err := s.encodeSession(data)
	if err != nil {
		return err
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     NuonAuthSessionName,
		Value:    encoded,
		Path:     "/",
		Domain:   s.domain,
		MaxAge:   sessionCookieMaxAge,
		Expires:  time.Now().Add(sessionCookieMaxAge * time.Second),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// clearSession removes the session cookie.
func (s *service) clearSession(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     NuonAuthSessionName,
		Value:    "",
		Path:     "/",
		Domain:   s.domain,
		MaxAge:   -1,
		Expires:  time.Now().Add(-time.Hour),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// encodeSession serializes and signs the session data.
// Format: base64(json) + "." + base64(hmac-sha256)
func (s *service) encodeSession(data *SessionData) (string, error) {
	// JSON encode the data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}

	// Base64 encode the JSON
	payload := base64.RawURLEncoding.EncodeToString(jsonData)

	// Create HMAC signature
	signature := s.signPayload(payload)

	// Combine payload and signature
	return payload + "." + signature, nil
}

// decodeSession verifies and deserializes the session data.
func (s *service) decodeSession(encoded string) (*SessionData, error) {
	// Split into payload and signature
	parts := strings.SplitN(encoded, ".", 2)
	if len(parts) != 2 {
		return nil, errInvalidSessionFormat
	}

	payload, signature := parts[0], parts[1]

	// Verify signature
	expectedSig := s.signPayload(payload)
	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return nil, errInvalidSessionSignature
	}

	// Base64 decode the payload
	jsonData, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode session payload: %w", err)
	}

	// JSON decode the data
	var data SessionData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Check expiration
	if time.Now().Unix()-data.CreatedAt > sessionCookieMaxAge {
		return nil, errSessionExpired
	}

	return &data, nil
}

// signPayload creates an HMAC-SHA256 signature for the payload.
func (s *service) signPayload(payload string) string {
	h := hmac.New(sha256.New, []byte(s.cfg.NuonAuthSessionKey))
	h.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
