package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
)

// AuthState handles the /auth/:state endpoint.
// This is where we:
// 1. Validate the state parameter matches what we stored in the session
// 2. Look up the identity provider from the session
// 3. Exchange the authorization code for tokens
// 4. Fetch user information from the provider
// 5. Create and set the auth cookie with a JWT
// 6. Redirect to the originally requested URL
func (s *service) AuthState(c *gin.Context) {
	// Get the state from URL path
	pathState := c.Param("state")
	if pathState == "" {
		s.respondError(c, http.StatusBadRequest, errInvalidState)
		return
	}

	// Get the session
	sessionData, err := s.getSession(c)
	if err != nil {
		s.l.Error("failed to get session", zap.Error(err))
		s.respondError(c, http.StatusBadRequest, errSessionNotFound)
		return
	}

	// Validate the state matches what we stored
	if sessionData.State != pathState {
		s.l.Error("state mismatch",
			zap.String("stored", sessionData.State),
			zap.String("received", pathState))
		s.respondError(c, http.StatusBadRequest, errStateMismatch)
		return
	}

	// Also verify query state matches (belt and suspenders)
	queryState := c.Query("state")
	if queryState != pathState {
		s.l.Error("query state mismatch",
			zap.String("path", pathState),
			zap.String("query", queryState))
		s.respondError(c, http.StatusBadRequest, errStateMismatch)
		return
	}

	// Get the provider type from session
	providerType := sessionData.ProviderID
	if providerType == "" {
		s.l.Error("no provider type in session")
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("no provider type in session"))
		return
	}

	// Look up and create the provider by type
	identityProvider, err := s.getIdentityProviderByType(c.Request.Context(), app.ProviderType(providerType))
	if err != nil {
		s.l.Error("failed to get identity provider",
			zap.String("provider_type", providerType),
			zap.Error(err))
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("invalid provider"))
		return
	}

	provider, err := s.createProviderFromIdentityProvider(identityProvider)
	if err != nil {
		s.l.Error("failed to create provider",
			zap.String("provider_type", providerType),
			zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to initialize provider"))
		return
	}

	// Get user info from the provider (this exchanges the code for tokens)
	userInfo, _, err := provider.GetUserInfo(c.Request.Context(), c.Request)
	if err != nil {
		s.l.Error("failed to get user info from provider", zap.Error(err))
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("failed to get user info: %w", err))
		return
	}

	s.l.Info("user authenticated",
		zap.String("email", userInfo.Email),
		zap.String("username", userInfo.Username),
		zap.String("subject", userInfo.Subject))

	// TODO: Verify/authorize the user against allowed domains, whitelists, etc.
	// For now, we allow all authenticated users
	if err := s.verifyUser(userInfo); err != nil {
		s.l.Warn("user not authorized", zap.Error(err))
		s.respondError(c, http.StatusForbidden, fmt.Errorf("user not authorized: %w", err))
		return
	}

	// Create JWT token for the cookie
	jwtToken, err := s.createJWT(userInfo, string(identityProvider.ProviderType))
	if err != nil {
		s.l.Error("failed to create JWT", zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to create token: %w", err))
		return
	}

	// Set the auth cookie
	s.setCookie(c, jwtToken)

	// Clear the session now that auth is complete
	s.clearSession(c)

	// Redirect to the originally requested URL, or show success page
	if sessionData.RequestedURL != "" {
		s.l.Debug("redirecting to requested URL", zap.String("url", sessionData.RequestedURL))
		s.redirect302(c, sessionData.RequestedURL)
		return
	}

	// No requested URL - show success page
	c.HTML(http.StatusOK, "auth/success.tmpl", gin.H{
		"Email":    userInfo.Email,
		"Username": userInfo.Username,
	})
}

// verifyUser checks if the user is authorized to access the system.
// TODO: Implement domain whitelists, team checks, etc.
func (s *service) verifyUser(userInfo *providers.UserInfo) error {
	// For now, allow all authenticated users
	// TODO: Add checks for:
	// - Allowed email domains
	// - User whitelists
	// - Team memberships (for GitHub)
	if userInfo.Email == "" && userInfo.Username == "" {
		return fmt.Errorf("user has no email or username")
	}
	return nil
}
