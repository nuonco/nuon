package service

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ProviderOption struct {
	ID           string
	Name         string
	Hint         string
	ProviderType string
}

func (s *service) Index(c *gin.Context) {
	redirectURL := c.Query("url")
	redirectURLEncoded := ""
	if redirectURL != "" {
		redirectURLEncoded = url.QueryEscape(redirectURL)
	}

	isAuthenticated := false
	var email string

	if token := s.findToken(c); token != "" {
		if tokenInfo, err := s.validateToken(token); err == nil {
			isAuthenticated = true
			email = tokenInfo.Email
		}
	}

	if isAuthenticated {
		if dest, ok := s.signedInRedirectURL(redirectURL); ok {
			s.redirect302(c, dest)
			return
		}
	}

	providers, err := s.getIdentityProviders(c.Request.Context())
	if err != nil {
		s.l.Error("failed to get identity providers", zap.String("service", "auth"), zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, err)
		return
	}

	options := make([]ProviderOption, 0, len(providers))
	for _, p := range providers {
		options = append(options, ProviderOption{
			ID:           p.ID,
			Name:         providerDisplayName(p),
			Hint:         providerHint(p),
			ProviderType: string(p.ProviderType),
		})
	}

	template := "auth/index.tmpl"
	data := gin.H{
		"IsAuthenticated": isAuthenticated,
		"Email":           email,
		"Providers":       options,
		"RedirectURL":     redirectURLEncoded,
		"DashboardURL":    s.cfg.AppURL,
	}
	if useNuonBrandedLogin(s.cfg.NuonBrandedLogin, s.cfg.AppURL) {
		template = "auth/index_nuon.tmpl"
		data["PostHogKey"] = s.cfg.PostHogKey
		data["PostHogHost"] = s.cfg.PostHogHost
	}

	c.HTML(http.StatusOK, template, data)
}

func useNuonBrandedLogin(flagEnabled bool, appURL string) bool {
	if flagEnabled {
		return true
	}

	u, err := url.Parse(appURL)
	if err != nil {
		return false
	}

	return u.Hostname() == "app.nuon.co"
}

func (s *service) signedInRedirectURL(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}

	decoded, err := url.QueryUnescape(raw)
	if err != nil || decoded == "" {
		return "", false
	}

	valid, err := s.validateRequestedURL(decoded)
	if err != nil {
		return "", false
	}

	return valid, true
}

func providerDisplayName(p *app.IdentityProvider) string {
	if p.Name != "" {
		return p.Name
	}

	switch p.ProviderType {
	case app.ProviderTypeGoogle:
		return "Google"
	case app.ProviderTypeGitHub:
		return "GitHub"
	case app.ProviderTypeOIDC:
		return "Single Sign-On"
	default:
		return string(p.ProviderType)
	}
}

func providerHint(p *app.IdentityProvider) string {
	if p.ProviderType != app.ProviderTypeOIDC {
		return ""
	}

	cfg, err := p.GetOpenIDConfig()
	if err != nil || cfg.IssuerURL == "" {
		return ""
	}

	issuer, err := url.Parse(cfg.IssuerURL)
	if err != nil {
		return ""
	}

	return issuer.Host
}
