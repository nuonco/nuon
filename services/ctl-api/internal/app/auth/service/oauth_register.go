package service

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type oauthRegisterRequest struct {
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	Scope                   string   `json:"scope"`
}

// oauthError writes an RFC 6749 / 7591 style error body.
func oauthError(c *gin.Context, status int, code, desc string) {
	c.JSON(status, gin.H{"error": code, "error_description": desc})
}

func isLoopbackHTTPRedirect(u *url.URL) bool {
	if u.Scheme != "http" {
		return false
	}
	host := u.Hostname()
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// isAllowedRedirectURI permits http loopback URLs (native/CLI MCP agents) and
// https URLs whose host is on the configured allowlist.
func isAllowedRedirectURI(raw string, httpsHosts []string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	if isLoopbackHTTPRedirect(u) {
		return true
	}
	if u.Scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, allowed := range httpsHosts {
		if host == strings.ToLower(strings.TrimSpace(allowed)) {
			return true
		}
	}
	return false
}

// OAuthRegister handles POST /oauth/register — OAuth 2.0 Dynamic Client
// Registration (RFC 7591). Clients self-register as public clients (no secret)
// that authenticate with PKCE.
func (s *service) OAuthRegister(c *gin.Context) {
	if !s.cfg.OAuthDCREnabled {
		oauthError(c, http.StatusForbidden, "access_denied", "dynamic client registration is disabled")
		return
	}

	var req oauthRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		oauthError(c, http.StatusBadRequest, "invalid_client_metadata", "unable to parse request body")
		return
	}

	if len(req.RedirectURIs) == 0 {
		oauthError(c, http.StatusBadRequest, "invalid_redirect_uri", "at least one redirect_uri is required")
		return
	}
	for _, uri := range req.RedirectURIs {
		if !isAllowedRedirectURI(uri, s.cfg.OAuthDCRRedirectHosts) {
			oauthError(c, http.StatusBadRequest, "invalid_redirect_uri", "redirect_uri must be an http loopback address or an allowlisted https host: "+uri)
			return
		}
	}

	// We only support public clients authenticating with PKCE.
	if req.TokenEndpointAuthMethod != "" && req.TokenEndpointAuthMethod != "none" {
		oauthError(c, http.StatusBadRequest, "invalid_client_metadata", "only token_endpoint_auth_method \"none\" is supported")
		return
	}

	clientName := strings.TrimSpace(req.ClientName)
	if clientName == "" {
		clientName = "mcp-client"
	}

	client := app.OAuthClient{
		ClientName:              clientName,
		RedirectURIs:            req.RedirectURIs,
		TokenEndpointAuthMethod: "none",
	}
	if err := s.db.WithContext(c.Request.Context()).Create(&client).Error; err != nil {
		s.l.Error("failed to register oauth client", zap.Error(err))
		oauthError(c, http.StatusInternalServerError, "server_error", "failed to register client")
		return
	}

	s.l.Info("oauth client registered", zap.String("client_id", client.ID), zap.String("client_name", client.ClientName))

	c.JSON(http.StatusCreated, gin.H{
		"client_id":                  client.ID,
		"client_id_issued_at":        client.CreatedAt.Unix(),
		"client_name":                client.ClientName,
		"redirect_uris":              []string(client.RedirectURIs),
		"token_endpoint_auth_method": "none",
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
	})
}
