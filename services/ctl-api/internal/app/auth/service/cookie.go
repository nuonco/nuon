package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//	helpers concerned with the cross-domain nuon auth cookie

func (s *service) clearCookie(c *gin.Context) {
	// Secure flag should be false for localhost (HTTP), true for production (HTTPS)
	secure := s.cfg.RootDomain != "localhost"

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     NuonAuthCookieName,
		Value:    "",
		Path:     "/",
		Domain:   s.cfg.RootDomain,
		MaxAge:   -1,
		Expires:  time.Now().Add(-time.Hour),
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *service) setCookie(c *gin.Context, token string) {
	// Secure flag should be false for localhost (HTTP), true for production (HTTPS)
	secure := s.cfg.RootDomain != "localhost"

	s.l.Debug("setting cookie",
		zap.String("service", "auth"),
		zap.String("domain", s.cfg.RootDomain),
		zap.Bool("secure", secure))

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     NuonAuthCookieName,
		Value:    token,
		Path:     "/",
		Domain:   s.cfg.RootDomain, // this should be the root domain
		MaxAge:   86400,            // 24 hours
		Expires:  time.Now().Add(time.Duration(s.cfg.NuonAuthSessionTTL) * time.Minute),
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *service) getCookie(c *gin.Context) (string, error) {
	cookie, err := c.Request.Cookie(NuonAuthCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
