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

	// Validate URL domain is within root domain to prevent open redirects
	if !s.isURLDomainAllowed(parsed.Host) {
		return "", fmt.Errorf("%w: %s", errURLDomainNotAllowed, parsed.Host)
	}

	return parsed.String(), nil
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

// isURLDomainAllowed checks if the URL's host is the root domain or a subdomain of it.
// This prevents open redirect attacks by ensuring redirects stay within the same domain.
func (s *service) isURLDomainAllowed(host string) bool {
	// Handle localhost for development
	if s.cfg.RootDomain == "localhost" {
		// Allow localhost with any port
		hostWithoutPort := strings.Split(host, ":")[0]
		return hostWithoutPort == "localhost" || hostWithoutPort == "127.0.0.1"
	}

	rootDomain := strings.ToLower(s.cfg.RootDomain)
	host = strings.ToLower(strings.Split(host, ":")[0]) // Remove port, lowercase

	// Exact match: host == rootDomain
	if host == rootDomain {
		return true
	}

	// Subdomain match: host ends with ".{rootDomain}"
	return strings.HasSuffix(host, "."+rootDomain)
}

// isEmailDomainAllowed checks if the email's domain is in the allowed domains list.
// Returns true if no allowed domains are configured (allow all) or if the domain matches.
func (s *service) isEmailDomainAllowed(email string) bool {
	// If no allowed domains configured, allow all
	if len(s.allowedDomains) == 0 {
		return true
	}

	// Extract domain from email
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	emailDomain := strings.ToLower(parts[1])

	// Check if domain is in allowed list
	return slices.Contains(s.allowedDomains, emailDomain)
}
