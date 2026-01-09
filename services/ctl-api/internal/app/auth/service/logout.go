package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logout handles the /logout endpoint.
// It clears the auth cookie and session, then optionally redirects to:
// 1. The OAuth provider's logout endpoint (if configured)
// 2. A specified redirect URL (if provided and allowed)
// 3. A logout success page
func (s *service) Logout(c *gin.Context) {
	s.l.Debug("/logout")

	// TODO: For provider logout (e.g., OIDC end_session_endpoint), we would need
	// to store the ID token or retrieve it from session storage.
	// For now, we just clear our local session/cookie.

	// Soft delete the token from the database
	if tokenValue := s.findToken(c); tokenValue != "" {
		if err := s.deleteToken(tokenValue); err != nil {
			s.l.Warn("failed to delete token", zap.Error(err))
		}
	}

	// Clear the auth cookie
	s.clearCookie(c)

	// Clear the session cookie
	s.clearSession(c)

	s.l.Debug("session and cookie cleared")

	// Get the redirect URL from query params
	redirectURL := c.Query("url")

	// Validate the redirect URL if provided
	if redirectURL != "" {
		// TODO: Validate against allowed post-logout redirect URLs
		// For now, validate basic URL safety
		if _, err := s.validateRequestedURL(redirectURL); err != nil {
			s.l.Warn("invalid logout redirect URL",
				zap.String("url", redirectURL),
				zap.Error(err))
			redirectURL = "" // Clear invalid URL
		}
	}

	// TODO: If provider has a logout endpoint, redirect there
	// providerLogoutURL := s.getProviderLogoutURL()
	// if providerLogoutURL != "" {
	//     logoutURL, _ := url.Parse(providerLogoutURL)
	//     q := logoutURL.Query()
	//     if redirectURL != "" {
	//         q.Add("post_logout_redirect_uri", redirectURL)
	//     }
	//     if idToken != "" {
	//         q.Add("id_token_hint", idToken)
	//     }
	//     logoutURL.RawQuery = q.Encode()
	//     s.redirect302(c, logoutURL.String())
	//     return
	// }

	// If we have a valid redirect URL, redirect there
	if redirectURL != "" {
		s.l.Debug("redirecting after logout", zap.String("url", redirectURL))
		s.redirect302(c, redirectURL)
		return
	}

	// Otherwise, show the logout success page
	c.HTML(http.StatusOK, "auth/logout.tmpl", gin.H{
		"Message": "You have been logged out successfully.",
	})
}
