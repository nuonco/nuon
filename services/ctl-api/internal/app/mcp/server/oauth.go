package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// authServerURL returns the OAuth 2.0 authorization server (Nuon Auth) base URL,
// matching how the auth service derives its own issuer.
func (s *Server) authServerURL() string {
	if s.cfg.RootDomain == "localhost" {
		return "http://localhost:8084"
	}
	return fmt.Sprintf("https://auth.%s", s.cfg.RootDomain)
}

// requestBaseURL reconstructs the externally-visible base URL of this MCP server
// from the incoming request, honoring reverse-proxy forwarding headers.
func requestBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	host := r.Host
	if fwd := r.Header.Get("X-Forwarded-Host"); fwd != "" {
		host = fwd
	}
	return fmt.Sprintf("%s://%s", scheme, host)
}

// resourceMetadataURL is the RFC 9728 protected-resource metadata document URL.
func resourceMetadataURL(r *http.Request) string {
	return requestBaseURL(r) + "/.well-known/oauth-protected-resource"
}

// protectedResourceMetadataHandler serves GET /.well-known/oauth-protected-resource
// (RFC 9728), telling MCP clients which authorization server to use.
func (s *Server) protectedResourceMetadataHandler(w http.ResponseWriter, r *http.Request) {
	meta := map[string]any{
		"resource":                 requestBaseURL(r) + "/mcp",
		"authorization_servers":    []string{s.authServerURL()},
		"scopes_supported":         []string{"org_read_only", "org_support", "org_admin"},
		"bearer_methods_supported": []string{"header"},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(meta)
}

// writeUnauthorized responds with 401 and a WWW-Authenticate header pointing at
// the protected-resource metadata, which bootstraps OAuth discovery in the client
// (MCP authorization spec / RFC 9728 §5.1).
func (s *Server) writeUnauthorized(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate",
		fmt.Sprintf(`Bearer resource_metadata=%q`, resourceMetadataURL(r)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":             "invalid_token",
		"error_description": "missing or invalid access token",
	})
}
