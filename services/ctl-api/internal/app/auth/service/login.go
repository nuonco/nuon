package service

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
)

// Login handles the /login endpoint.
// It initiates the OAuth flow by:
// 1. Clearing any existing auth cookie
// 2. Looking up the requested identity provider
// 3. Generating a state nonce for CSRF protection
// 4. Storing the requested URL, provider ID, and state in the session
// 5. Redirecting to the OAuth provider's authorization URL
func (s *service) Login(c *gin.Context) {
	// Clear any existing auth cookie
	s.clearCookie(c)

	// Get the provider type from query params (required)
	providerType := c.Query("provider")
	if providerType == "" {
		s.l.Warn("login attempt without provider type")
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("provider type is required"))
		return
	}

	// Look up the identity provider by type
	identityProvider, err := s.getIdentityProviderByType(c.Request.Context(), app.ProviderType(providerType))
	if err != nil {
		s.l.Error("failed to get identity provider",
			zap.String("service", "auth"),
			zap.String("provider_type", providerType),
			zap.Error(err))
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("invalid provider: %s", providerType))
		return
	}

	// Create the OAuth provider from the identity provider
	provider, err := s.createProviderFromIdentityProvider(identityProvider)
	if err != nil {
		s.l.Error("failed to create provider",
			zap.String("service", "auth"),
			zap.String("provider_id", providerType),
			zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to initialize provider"))
		return
	}

	// Check for existing session to get fail count
	var failCount int
	if existingSession, err := s.getSession(c); err == nil {
		s.l.Debug("increasing failCount", zap.String("service", "auth"), zap.Int("failCount", failCount))
		failCount = existingSession.FailCount
	}

	// Generate a state nonce for CSRF protection
	state, err := generateStateNonce()
	if err != nil {
		s.l.Error("failed to generate state nonce", zap.String("service", "auth"), zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to generate state: %w", err))
		return
	}

	// Get and validate the requested URL from query params
	// URL may be encoded, so decode it first before validation
	requestedURL := c.Query("url")
	if requestedURL != "" {
		// Decode URL-encoded value (handles double-encoding edge cases)
		decodedURL, err := url.QueryUnescape(requestedURL)
		if err != nil {
			s.l.Warn("failed to decode requested URL",
				zap.String("service", "auth"),
				zap.String("url", requestedURL),
				zap.Error(err))
			s.respondError(c, http.StatusBadRequest, fmt.Errorf("invalid URL encoding"))
			return
		}
		requestedURL = decodedURL

		// Validate the decoded URL (must have http:// or https:// prefix)
		validURL, err := s.validateRequestedURL(requestedURL)
		if err != nil {
			s.l.Warn("invalid requested URL",
				zap.String("service", "auth"),
				zap.String("url", requestedURL),
				zap.Error(err))
			s.respondError(c, http.StatusBadRequest, err)
			return
		}
		requestedURL = validURL
	}

	// Increment fail count
	failCount++

	// Check for too many failed attempts before saving session
	if failCount > failCountLimit {
		errorMsg := c.Query("error")
		s.l.Warn("too many redirect attempts",
			zap.String("service", "auth"),
			zap.String("url", requestedURL),
			zap.String("error", errorMsg),
			zap.Int("failCount", failCount))
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("%w for %s", errTooManyRedirects, requestedURL))
		return
	}

	// Create and save the session with provider ID
	sessionData := &SessionData{
		State:        state,
		ProviderID:   providerType,
		RequestedURL: requestedURL,
		FailCount:    failCount,
	}

	if err := s.setSession(c, sessionData); err != nil {
		s.l.Error("failed to save session", zap.String("service", "auth"), zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to save session: %w", err))
		return
	}

	s.l.Debug("login session state",
		zap.String("service", "auth"),
		zap.String("state", state),
		zap.String("provider_id", providerType),
		zap.String("requestedURL", requestedURL),
		zap.Int("failCount", failCount))

	// Build the OAuth authorization URL
	authURL, err := s.buildOAuthURL(provider, state)
	if err != nil {
		s.l.Error("failed to build OAuth URL",
			zap.String("service", "auth"),
			zap.String("provider_id", providerType),
			zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("provider configuration error"))
		return
	}

	s.l.Debug("redirecting to OAuth provider",
		zap.String("service", "auth"),
		zap.String("authURL", authURL))

	// Redirect to the OAuth provider
	s.redirect302(c, authURL)
}

// buildOAuthURL constructs the OAuth authorization URL with the given state.
func (s *service) buildOAuthURL(provider providers.Provider, state string) (string, error) {
	// Get the OAuth2 config from the provider via GetOAuth2Config()
	// This avoids signature mismatch issues with variadic AuthCodeURL methods
	if bp, ok := provider.(interface {
		GetOAuth2Config() *oauth2.Config
	}); ok {
		cfg := bp.GetOAuth2Config()
		if cfg != nil {
			return cfg.AuthCodeURL(state), nil
		}
	}

	return "", fmt.Errorf("provider %s does not have a valid OAuth2 configuration", provider.Name())
}
