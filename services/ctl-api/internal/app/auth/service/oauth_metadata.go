package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// oauthIssuer returns the base URL the auth service is served at, used as the
// OAuth 2.0 issuer identifier and to build absolute endpoint URLs.
func (s *service) oauthIssuer() string {
	if s.cfg.RootDomain == "localhost" {
		return "http://localhost:8084"
	}
	return fmt.Sprintf("https://%s", s.domain)
}

// oauthScopesSupported lists the scopes clients may request. They map to the
// org role granted to the issued token.
func oauthScopesSupported() []string {
	return []string{"org_read_only", "org_admin"}
}

// OAuthAuthorizationServerMetadata handles GET /.well-known/oauth-authorization-server
// (RFC 8414). It lets MCP clients discover the authorization/token/registration
// endpoints and supported capabilities with no manual configuration.
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
