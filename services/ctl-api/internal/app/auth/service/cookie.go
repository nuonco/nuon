package service

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//	helpers concerned with the cross-domain nuon auth cookie

// parentDomain returns the parent domain by stripping the first label.
// e.g. "app.stage.nuon.co" -> "stage.nuon.co", "app.nuon.co" -> "nuon.co".
// Returns empty string if there is no parent (e.g. "localhost").
func parentDomain(domain string) string {
	idx := strings.IndexByte(domain, '.')
	if idx < 0 || idx == len(domain)-1 {
		return ""
	}
	return domain[idx+1:]
}

func (s *service) clearCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     NuonAuthCookieName,
		Value:    "",
		Path:     "/",
		Domain:   s.cfg.RootDomain,
		MaxAge:   -1,
		Expires:  time.Now().Add(-time.Hour),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	if pd := parentDomain(s.cfg.RootDomain); pd != "" {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     NuonAuthCookieName,
			Value:    "",
			Path:     "/",
			Domain:   pd,
			MaxAge:   -1,
			Expires:  time.Now().Add(-time.Hour),
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}
}

func (s *service) setCookie(c *gin.Context, token string) {
	s.l.Debug("setting cookie", zap.String("service", "auth"), zap.String("domain", s.cfg.RootDomain))
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     NuonAuthCookieName,
		Value:    token,
		Path:     "/",
		Domain:   s.cfg.RootDomain,
		MaxAge:   86400,
		Expires:  time.Now().Add(time.Duration(s.cfg.NuonAuthSessionTTL) * time.Minute),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	if pd := parentDomain(s.cfg.RootDomain); pd != "" {
		s.l.Debug("setting parent domain cookie", zap.String("service", "auth"), zap.String("domain", pd))
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     NuonAuthCookieName,
			Value:    token,
			Path:     "/",
			Domain:   pd,
			MaxAge:   86400,
			Expires:  time.Now().Add(time.Duration(s.cfg.NuonAuthSessionTTL) * time.Minute),
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}
}

func (s *service) getCookie(c *gin.Context) (string, error) {
	cookie, err := c.Request.Cookie(NuonAuthCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
