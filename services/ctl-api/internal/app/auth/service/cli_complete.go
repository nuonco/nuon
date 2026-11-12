package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/nuonauth"
)

// CLIComplete handles GET /cli/complete, where the CLI's loopback listener
// sends the browser after a login. It never echoes request input.
func (s *service) CLIComplete(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Referrer-Policy", "no-referrer")

	switch c.Query("result") {
	case nuonauth.CLIResultSuccess:
		email := ""
		if token := s.findToken(c); token != "" {
			if info, err := s.validateToken(token); err == nil {
				email = info.Email
			}
		}
		c.HTML(http.StatusOK, "auth/device_success.tmpl", gin.H{"Email": email})
	case nuonauth.CLIResultDenied:
		c.HTML(http.StatusOK, "auth/device_denied.tmpl", gin.H{})
	default:
		c.HTML(http.StatusBadRequest, "auth/error.tmpl", gin.H{
			"Error":  fmt.Sprintf("CLI login did not finish. Return to your terminal and run %q again.", "nuon auth login"),
			"Status": http.StatusBadRequest,
		})
	}
}
