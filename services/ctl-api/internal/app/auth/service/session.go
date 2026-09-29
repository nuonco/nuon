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

const sessionCookieMaxAge = 5 * 60

func (s *service) getSession(c *gin.Context) (*SessionData, error) {
	cookie, err := c.Request.Cookie(NuonAuthSessionName)
	if err != nil {
		return nil, err
	}

	return s.decodeSession(cookie.Value)
}

func (s *service) setSession(c *gin.Context, data *SessionData) error {
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

func (s *service) encodeSession(data *SessionData) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}

	payload := base64.RawURLEncoding.EncodeToString(jsonData)

	signature := s.signPayload(payload)

	return payload + "." + signature, nil
}

func (s *service) decodeSession(encoded string) (*SessionData, error) {
	parts := strings.SplitN(encoded, ".", 2)
	if len(parts) != 2 {
		return nil, errInvalidSessionFormat
	}

	payload, signature := parts[0], parts[1]

	expectedSig := s.signPayload(payload)
	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return nil, errInvalidSessionSignature
	}

	jsonData, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode session payload: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	if time.Now().Unix()-data.CreatedAt > sessionCookieMaxAge {
		return nil, errSessionExpired
	}

	return &data, nil
}

func (s *service) signPayload(payload string) string {
	h := hmac.New(sha256.New, []byte(s.cfg.NuonAuthSessionKey))
	h.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
