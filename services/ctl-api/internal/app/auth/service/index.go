package service

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ProviderOption represents a login option to display in the UI.
type ProviderOption struct {
	ID           string
	Name         string
	ProviderType string
}

// Index handles the root / endpoint.
// It displays a simple landing page with login options for each provider.
// If a `url` query param is provided, it will be passed to the login handler.
func (s *service) Index(c *gin.Context) {
	// Get the redirect URL from query params (to pass along to login)
	// URL-encode it for safe inclusion in query strings
	redirectURL := c.Query("url")
	redirectURLEncoded := ""
	if redirectURL != "" {
		redirectURLEncoded = url.QueryEscape(redirectURL)
	}

	// Check if user is already authenticated
	isAuthenticated := false
	var email string

	if token := s.findToken(c); token != "" {
		if tokenInfo, err := s.validateToken(token); err == nil {
			isAuthenticated = true
			email = tokenInfo.Email

			// If already authenticated and a redirect URL is provided, redirect there
			if redirectURL != "" {
				validatedURL, err := s.validateRequestedURL(redirectURL)
				if err != nil {
					s.l.Warn("invalid redirect URL for authenticated user",
						zap.String("url", redirectURL),
						zap.Error(err))
				} else {
					s.l.Info("redirecting authenticated user to requested URL",
						zap.String("email", email),
						zap.String("url", validatedURL))
					s.redirect302(c, validatedURL)
					return
				}
			}
		}
	}

	// Get available identity providers
	providers, err := s.getIdentityProviders(c.Request.Context())
	if err != nil {
		s.l.Error("failed to get identity providers", zap.String("service", "auth"), zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, err)
		return
	}

	// Convert to template-friendly format
	options := make([]ProviderOption, 0, len(providers))
	for _, p := range providers {
		options = append(options, ProviderOption{
			ID:           p.ID,
			Name:         providerDisplayName(p),
			ProviderType: string(p.ProviderType),
		})
	}

	c.HTML(http.StatusOK, "auth/index.tmpl", gin.H{
		"IsAuthenticated": isAuthenticated,
		"Email":           email,
		"Providers":       options,
		"RedirectURL":     redirectURLEncoded,
	})
}

// providerDisplayName returns a human-readable name for the provider.
func providerDisplayName(p *app.IdentityProvider) string {
	switch p.ProviderType {
	case app.ProviderTypeGoogle:
		return "Google"
	case app.ProviderTypeGitHub:
		return "GitHub"
	case app.ProviderTypeOIDC:
		return "Single Sign-On"
	default:
		return string(p.ProviderType)
	}
}
