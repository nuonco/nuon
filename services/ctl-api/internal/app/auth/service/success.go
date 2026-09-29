package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (s *service) Success(c *gin.Context) {
	token := s.findToken(c)
	if token == "" {
		s.l.Debug("no auth cookie found, redirecting to index")
		s.redirect302(c, "/")
		return
	}

	tokenInfo, err := s.validateToken(token)
	if err != nil {
		s.l.Warn("invalid auth cookie, redirecting to index",
			zap.Error(err))
		s.clearCookie(c)
		s.redirect302(c, "/")
		return
	}

	c.HTML(http.StatusOK, "auth/success.tmpl", gin.H{
		"Email":    tokenInfo.Email,
		"Username": tokenInfo.Username,
	})
}
