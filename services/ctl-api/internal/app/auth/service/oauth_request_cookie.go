package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const oauthRequestCookieName = "nuon-oauth-rid"

func (s *service) oauthRequestCookieValue(requestID string) string {
	mac := hmac.New(sha256.New, []byte(s.cfg.NuonAuthSessionKey))
	mac.Write([]byte(requestID))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *service) setOAuthRequestCookie(c *gin.Context, requestID string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oauthRequestCookieName,
		Value:    s.oauthRequestCookieValue(requestID),
		Path:     "/oauth",
		MaxAge:   int(oauthAuthorizationCodeExpiry.Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *service) clearOAuthRequestCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oauthRequestCookieName,
		Value:    "",
		Path:     "/oauth",
		MaxAge:   -1,
		Expires:  time.Now().Add(-time.Hour),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *service) requireOAuthRequestCookie(c *gin.Context, requestID string) bool {
	cookie, err := c.Request.Cookie(oauthRequestCookieName)
	expected := s.oauthRequestCookieValue(requestID)
	if err != nil || !hmac.Equal([]byte(cookie.Value), []byte(expected)) {
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("authorization request is not bound to this browser"))
		return false
	}
	return true
}
