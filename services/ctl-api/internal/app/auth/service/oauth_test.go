package service

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func s256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
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
	cases := []struct {
		uri  string
		want bool
	}{
		{"https://claude.ai/callback", true},
		{"https://example.com/oauth/cb", true},
		{"http://localhost:8765/callback", true},
		{"http://127.0.0.1:53210/cb", true},
		{"http://[::1]:5000/cb", true},
		{"http://evil.com/callback", false}, // http non-loopback
		{"ftp://localhost/cb", false},       // unsupported scheme
		{"not a url", false},
		{"https://", false}, // no host
		{"", false},
	}
	for _, tc := range cases {
		assert.Equalf(t, tc.want, isAllowedRedirectURI(tc.uri), "uri=%q", tc.uri)
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
}
