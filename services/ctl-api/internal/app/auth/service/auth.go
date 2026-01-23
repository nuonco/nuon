package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Auth handles the /auth callback endpoint from the OAuth provider.
// It validates the response and redirects to /auth/:state for final processing.
// This two-step approach ensures the session cookie path is properly scoped.
func (s *service) Auth(c *gin.Context) {
	// Check if the IdP returned an error
	if idpError := c.Query("error"); idpError != "" {
		errorDesc := c.Query("error_description")
		s.l.Error("OAuth provider returned error",
			zap.String("error", idpError),
			zap.String("description", errorDesc))
		s.respondError(c, http.StatusUnauthorized, fmt.Errorf("OAuth error: %s - %s", idpError, errorDesc))
		return
	}

	// Get the state from query params
	queryState := c.Query("state")
	if queryState == "" {
		s.l.Error("no state parameter in callback")
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("missing state parameter in OAuth callback"))
		return
	}

	// Redirect to /auth/:state with the full query string preserved
	// This allows the session cookie to be properly scoped
	authStateURL := fmt.Sprintf("/auth/%s?%s", queryState, c.Request.URL.RawQuery)

	s.l.Debug("redirecting to auth state handler",
		zap.String("state", queryState),
		zap.String("authStateURL", authStateURL))

	s.redirect302(c, authStateURL)
}
