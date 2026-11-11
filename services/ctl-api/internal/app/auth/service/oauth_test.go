package service

import (
	"crypto/sha256"
	"encoding/base64"
	"html/template"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/nuonauth"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func s256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func TestAuthTemplatesParse(t *testing.T) {
	sub, err := fs.Sub(tmplFS, "templates")
	assert.NoError(t, err)
	_, err = template.ParseFS(sub, "*.tmpl")
	assert.NoError(t, err)
}

func TestVerifyPKCE(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := s256(verifier)

	assert.True(t, verifyPKCE(verifier, challenge), "correct verifier should pass")
	assert.False(t, verifyPKCE("wrong-verifier", challenge), "wrong verifier should fail")
	assert.False(t, verifyPKCE(verifier, "not-the-challenge"), "wrong challenge should fail")
	assert.False(t, verifyPKCE("", challenge), "empty verifier should fail")
	assert.False(t, verifyPKCE(verifier, ""), "empty challenge should fail")
}

func TestIsAllowedRedirectURI(t *testing.T) {
	hosts := []string{
		"claude.ai", "www.claude.ai", "claude.com",
		"cursor.com", "www.cursor.com",
		"vscode.dev", "insiders.vscode.dev",
	}
	cases := []struct {
		uri  string
		want bool
	}{
		{"https://claude.ai/api/mcp/auth_callback", true},
		{"https://claude.com/api/mcp/auth_callback", true},
		{"https://www.cursor.com/agents/mcp/oauth/callback", true},
		{"https://vscode.dev/redirect", true},
		{"https://example.com/oauth/cb", false},
		{"http://localhost:8765/callback", true},
		{"http://127.0.0.1:53210/cb", true},
		{"http://127.0.0.42:9/cb", true},
		{"http://[::1]:5000/cb", true},
		{"http://evil.com/callback", false},
		{"https://localhost/cb", false},
		{"ftp://localhost/cb", false},
		{"not a url", false},
		{"https://", false},
		{"", false},
	}
	for _, tc := range cases {
		assert.Equalf(t, tc.want, isAllowedRedirectURI(tc.uri, hosts), "uri=%q", tc.uri)
	}
}

func TestOAuthScopeToRole(t *testing.T) {
	assert.Equal(t, string(app.RoleTypeOrgAdmin), oauthScopeToRole(string(app.RoleTypeOrgAdmin)))
	assert.Equal(t, string(app.RoleTypeOrgReadOnly), oauthScopeToRole(string(app.RoleTypeOrgReadOnly)))
	// only org_admin grants write; everything else (support, unknown, empty)
	// defaults to least-privileged read-only
	assert.Equal(t, string(app.RoleTypeOrgReadOnly), oauthScopeToRole(string(app.RoleTypeOrgSupport)))
	assert.Equal(t, string(app.RoleTypeOrgReadOnly), oauthScopeToRole("bogus"))
	assert.Equal(t, string(app.RoleTypeOrgReadOnly), oauthScopeToRole(""))
}

func TestOAuthScopesSupported(t *testing.T) {
	assert.Equal(t, []string{"org_read_only", "org_admin"}, oauthScopesSupported())
}

func TestOAuthClientAllowsRedirectURI(t *testing.T) {
	client := app.OAuthClient{
		RedirectURIs: []string{"https://a.example/cb", "http://localhost:9000/cb"},
	}
	assert.True(t, client.AllowsRedirectURI("https://a.example/cb"))
	assert.True(t, client.AllowsRedirectURI("http://localhost:9000/cb"))
	assert.False(t, client.AllowsRedirectURI("https://a.example/other"), "exact match only")
	assert.False(t, client.AllowsRedirectURI("https://b.example/cb"))
	assert.False(t, client.AllowsRedirectURI("http://localhost:9001/cb"), "mcp clients do not ignore loopback port")
}

func TestNuonCLIAllowsLoopbackPort(t *testing.T) {
	client := app.OAuthClient{
		ID:           nuonauth.CLIClientID,
		RedirectURIs: []string{nuonauth.CLILoopbackRedirectURI},
	}
	assert.True(t, client.AllowsRedirectURI("http://127.0.0.1:54321/callback"))
	assert.True(t, client.AllowsRedirectURI("http://localhost:9/callback"))
	assert.True(t, client.AllowsRedirectURI("http://[::1]:5000/callback"))
	assert.False(t, client.AllowsRedirectURI("http://127.0.0.1:54321/other"))
	assert.False(t, client.AllowsRedirectURI("https://127.0.0.1:54321/callback"))
	assert.False(t, client.AllowsRedirectURI("http://127.0.0.1:54321/callback?code=1"))
	assert.False(t, client.AllowsRedirectURI("http://evil.example/callback"))

	mcp := app.OAuthClient{
		ID:           "oacother00000000000000000",
		RedirectURIs: []string{"http://127.0.0.1:1111/cb"},
	}
	assert.True(t, mcp.AllowsRedirectURI("http://127.0.0.1:1111/cb"))
	assert.False(t, mcp.AllowsRedirectURI("http://127.0.0.1:2222/cb"))
}
