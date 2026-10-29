package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/origin"
)

const (
	connectStateCookie = "nuon_github_connect"
	connectStateTTL    = 15 * time.Minute
)

type connectState struct {
	Nonce      string `json:"n"`
	OrgID      string `json:"o"`
	Onboarding bool   `json:"b,omitempty"`
	TokenHash  string `json:"t"`
	ExpiresAt  int64  `json:"e"`
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (h *ConnectHandler) setConnectState(c *gin.Context, orgID, token string, onboarding bool) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	nonce := base64.RawURLEncoding.EncodeToString(buf)
	value, err := sealConnectState(connectState{
		Nonce:      nonce,
		OrgID:      orgID,
		Onboarding: onboarding,
		TokenHash:  tokenHash(token),
		ExpiresAt:  time.Now().Add(connectStateTTL).Unix(),
	})
	if err != nil {
		return "", err
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     connectStateCookie,
		Value:    value,
		Path:     "/connect",
		MaxAge:   int(connectStateTTL.Seconds()),
		Expires:  time.Now().Add(connectStateTTL),
		Secure:   connectCookieSecure(h.cfg.AppUrl),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nonce, nil
}

func (h *ConnectHandler) openConnectState(c *gin.Context) (connectState, error) {
	cookie, err := c.Cookie(connectStateCookie)
	if err != nil {
		return connectState{}, err
	}
	state, err := openConnectState(cookie)
	if err != nil {
		return connectState{}, err
	}
	if time.Now().Unix() > state.ExpiresAt || !orgIDPattern.MatchString(state.OrgID) || state.Nonce == "" {
		return state, errors.New("github connect state is not usable")
	}
	return state, nil
}

func (h *ConnectHandler) clearConnectState(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     connectStateCookie,
		Value:    "",
		Path:     "/connect",
		MaxAge:   -1,
		Expires:  time.Now().Add(-time.Hour),
		Secure:   connectCookieSecure(h.cfg.AppUrl),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func connectCookieSecure(appURL string) bool {
	return !origin.IsHTTPLoopback(appURL)
}

func sealConnectState(state connectState) (string, error) {
	raw, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func openConnectState(value string) (connectState, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return connectState{}, err
	}
	var state connectState
	if err := json.Unmarshal(raw, &state); err != nil {
		return connectState{}, err
	}
	return state, nil
}

func hmacEqual(a, b string) bool {
	return a != "" && hmac.Equal([]byte(a), []byte(b))
}
