package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *service) oauthIssuer() string {
	if s.cfg.RootDomain == "localhost" {
		return "http://localhost:8084"
	}
	return fmt.Sprintf("https://%s", s.domain)
}

func oauthScopesSupported() []string {
	return []string{"org_read_only", "org_admin"}
}

func (s *service) OAuthAuthorizationServerMetadata(c *gin.Context) {
	issuer := s.oauthIssuer()

	meta := gin.H{
		"issuer":                                issuer,
		"authorization_endpoint":                issuer + "/oauth/authorize",
		"token_endpoint":                        issuer + "/oauth/token",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"none"},
		"scopes_supported":                      oauthScopesSupported(),
	}

	if s.cfg.OAuthDCREnabled {
		meta["registration_endpoint"] = issuer + "/oauth/register"
	}

	c.JSON(http.StatusOK, meta)
}
