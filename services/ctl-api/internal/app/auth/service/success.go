package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Success handles the /success endpoint.
// It checks for a valid auth cookie and displays the success page.
// If no valid cookie is found, redirects to the index page.
func (s *service) Success(c *gin.Context) {
	// Check for auth cookie
	token := s.findToken(c)
	if token == "" {
		s.l.Debug("no auth cookie found, redirecting to index")
		s.redirect302(c, "/")
		return
	}

	// Validate the token
	tokenInfo, err := s.validateToken(token)
	if err != nil {
		s.l.Warn("invalid auth cookie, redirecting to index",
			zap.Error(err))
		s.clearCookie(c)
		s.redirect302(c, "/")
		return
	}

	// Show success page with user info from token
	c.HTML(http.StatusOK, "auth/success.tmpl", gin.H{
		"Email":    tokenInfo.Email,
		"Username": tokenInfo.Username,
	})
}
