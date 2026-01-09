package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Dangerous URL patterns to reject
var dangerousPatterns = []string{
	"javascript:",
	"data:",
	"vbscript:",
	"file://",
}

// regExAlphaNum matches only alphanumeric characters
var regExAlphaNum = regexp.MustCompile("[^a-zA-Z0-9]+")

// TODO: write some tests for this
// generateStateNonce creates a cryptographically secure random state string.
func generateStateNonce() (string, error) {
	b := make([]byte, 32) // does this have to be configurable?
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state nonce: %w", err)
	}
	// Encode to base64 and strip non-alphanumeric chars for URL safety
	state := base64.URLEncoding.EncodeToString(b)
	state = regExAlphaNum.ReplaceAllString(state, "")
	return state, nil
}

// validateRequestedURL validates and sanitizes the requested redirect URL.
func (s *service) validateRequestedURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", errNoURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errInvalidURL, err)
	}

	// Must be http or https
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errURLNotHTTP
	}

	// Check for dangerous patterns in the URL
	lowerURL := strings.ToLower(rawURL)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerURL, pattern) {
			return "", fmt.Errorf("%w: contains %s", errDangerousQS, pattern)
		}
	}

	// Check query string values for dangerous patterns
	for _, values := range parsed.Query() {
		for _, val := range values {
			lowerVal := strings.ToLower(val)
			for _, pattern := range dangerousPatterns {
				if strings.HasPrefix(lowerVal, pattern) {
					return "", fmt.Errorf("%w: query param contains %s", errDangerousQS, pattern)
				}
			}
		}
	}

	// TODO: validate against allowed domains if configured
	// For now, we trust the URL if it passes basic validation

	return parsed.String(), nil
}

// clearCookie removes the auth cookie by setting it to empty and expired.
func (s *service) clearCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     NuonAuthCookieName,
		Value:    "",
		Path:     "/",
		Domain:   s.cfg.NuonAuthDomain,
		MaxAge:   -1,
		Expires:  time.Now().Add(-time.Hour),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
}

// setCookie sets the auth cookie with the given JWT token value.
func (s *service) setCookie(c *gin.Context, token string) {
	s.l.Debug("setting cookie", zap.String("service", "auth"), zap.String("domain", s.cfg.NuonAuthDomain))
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     NuonAuthCookieName,
		Value:    token,
		Path:     "/",
		Domain:   s.cfg.NuonAuthDomain,
		MaxAge:   86400, // 24 hours
		Expires:  time.Now().Add(24 * time.Hour),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
}

// getCookie retrieves the auth cookie value from the request.
func (s *service) getCookie(c *gin.Context) (string, error) {
	cookie, err := c.Request.Cookie(NuonAuthCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// respondError sends an error response with the appropriate status code.
func (s *service) respondError(c *gin.Context, status int, err error) {
	s.l.Error("nuon auth error",
		zap.Int("status", status),
		zap.Error(err),
		zap.String("path", c.Request.URL.Path),
	)
	c.HTML(status, "auth/error.tmpl", gin.H{
		"Error":  err.Error(),
		"Status": status,
	})
}

// redirect302 performs a 302 redirect to the given URL.
func (s *service) redirect302(c *gin.Context, url string) {
	s.l.Debug("redirecting",
		zap.String("url", url),
	)
	c.Redirect(http.StatusFound, url)
}
