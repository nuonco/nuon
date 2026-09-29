package service

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
)

func (s *service) Login(c *gin.Context) {
	s.clearCookie(c)

	providerRef := c.Query("provider")
	if providerRef == "" {
		s.l.Warn("login attempt without provider")
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("provider is required"))
		return
	}

	identityProvider, err := s.getIdentityProvider(c.Request.Context(), providerRef)
	if err != nil {
		s.l.Error("failed to get identity provider",
			zap.String("service", "auth"),
			zap.String("provider", providerRef),
			zap.Error(err))
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("invalid provider: %s", providerRef))
		return
	}

	provider, err := s.createProviderFromIdentityProvider(identityProvider)
	if err != nil {
		s.l.Error("failed to create provider",
			zap.String("service", "auth"),
			zap.String("provider_id", identityProvider.ID),
			zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to initialize provider"))
		return
	}

	var failCount int
	if existingSession, err := s.getSession(c); err == nil {
		s.l.Debug("increasing failCount", zap.String("service", "auth"), zap.Int("failCount", failCount))
		failCount = existingSession.FailCount
	}

	state, err := generateStateNonce()
	if err != nil {
		s.l.Error("failed to generate state nonce", zap.String("service", "auth"), zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to generate state: %w", err))
		return
	}

	requestedURL := c.Query("url")
	if requestedURL != "" {
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

	failCount++

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

	sessionData := &SessionData{
		State:        state,
		ProviderID:   identityProvider.ID,
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
		zap.String("provider_id", identityProvider.ID),
		zap.String("requestedURL", requestedURL),
		zap.Int("failCount", failCount))

	authURL, err := s.buildOAuthURL(provider, state)
	if err != nil {
		s.l.Error("failed to build OAuth URL",
			zap.String("service", "auth"),
			zap.String("provider_id", identityProvider.ID),
			zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("provider configuration error"))
		return
	}

	s.l.Debug("redirecting to OAuth provider",
		zap.String("service", "auth"),
		zap.String("authURL", authURL))

	s.redirect302(c, authURL)
}

func (s *service) buildOAuthURL(provider providers.Provider, state string) (string, error) {
	// why: Get the OAuth2 config from the provider via GetOAuth2Config()
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
