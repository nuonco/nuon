package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var dangerousPatterns = []string{
	"javascript:",
	"data:",
	"vbscript:",
	"file://",
}

var regExAlphaNum = regexp.MustCompile("[^a-zA-Z0-9]+")

// TODO: write some tests for this
// generateStateNonce creates a cryptographically secure random state string.
func generateStateNonce() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state nonce: %w", err)
	}
	state := base64.URLEncoding.EncodeToString(b)
	state = regExAlphaNum.ReplaceAllString(state, "")
	return state, nil
}

func (s *service) validateRequestedURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", errNoURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errInvalidURL, err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errURLNotHTTP
	}

	lowerURL := strings.ToLower(rawURL)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerURL, pattern) {
			return "", fmt.Errorf("%w: contains %s", errDangerousQS, pattern)
		}
	}

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

	if !s.isURLDomainAllowed(parsed.Host) {
		return "", fmt.Errorf("%w: %s", errURLDomainNotAllowed, parsed.Host)
	}

	return parsed.String(), nil
}

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

func (s *service) redirect302(c *gin.Context, url string) {
	s.l.Debug("redirecting",
		zap.String("url", url),
	)
	c.Redirect(http.StatusFound, url)
}

func (s *service) isURLDomainAllowed(host string) bool {
	if s.cfg.RootDomain == "localhost" {
		hostWithoutPort := strings.Split(host, ":")[0]
		return hostWithoutPort == "localhost" || hostWithoutPort == "127.0.0.1"
	}

	rootDomain := strings.ToLower(s.cfg.RootDomain)
	host = strings.ToLower(strings.Split(host, ":")[0])

	if host == rootDomain {
		return true
	}

	return strings.HasSuffix(host, "."+rootDomain)
}

func (s *service) isEmailDomainAllowed(email string) bool {
	if len(s.allowedDomains) == 0 {
		return true
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	emailDomain := strings.ToLower(parts[1])

	return slices.Contains(s.allowedDomains, emailDomain)
}
