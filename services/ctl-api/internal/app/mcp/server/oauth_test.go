package server

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestRequestBaseURL(t *testing.T) {
	cases := []struct {
		name    string
		host    string
		tls     bool
		fwdSch  string
		fwdHost string
		want    string
	}{
		{name: "plain http", host: "localhost:8088", want: "http://localhost:8088"},
		{name: "tls", host: "ctl.nuon.co", tls: true, want: "https://ctl.nuon.co"},
		{name: "forwarded proto", host: "ctl.nuon.co", fwdSch: "https", want: "https://ctl.nuon.co"},
		{name: "forwarded host", host: "internal:8088", fwdSch: "https", fwdHost: "ctl.nuon.co", want: "https://ctl.nuon.co"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Host = tc.host
			if tc.tls {
				r.TLS = &tls.ConnectionState{}
			}
			if tc.fwdSch != "" {
				r.Header.Set("X-Forwarded-Proto", tc.fwdSch)
			}
			if tc.fwdHost != "" {
				r.Header.Set("X-Forwarded-Host", tc.fwdHost)
			}
			assert.Equal(t, tc.want, requestBaseURL(r))
		})
	}
}

func TestWriteUnauthorizedSetsWWWAuthenticate(t *testing.T) {
	s := &Server{}
	r := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	r.Host = "ctl.nuon.co"
	r.TLS = &tls.ConnectionState{}
	w := httptest.NewRecorder()

	s.writeUnauthorized(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t,
		`Bearer resource_metadata="https://ctl.nuon.co/.well-known/oauth-protected-resource"`,
		w.Header().Get("WWW-Authenticate"))
}

func TestWriteForbidden(t *testing.T) {
	w := httptest.NewRecorder()

	(&Server{implementationName: "nuonctl", requireEmployee: true}).writeForbidden(w, "nuonctl MCP is limited to Nuon employees")

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.JSONEq(t, `{
		"error": "access_denied",
		"error_description": "nuonctl MCP is limited to Nuon employees"
	}`, w.Body.String())
}

func TestAccountHasOrgAccess(t *testing.T) {
	acct := &app.Account{OrgIDs: []string{"org_a", "org_b"}}
	assert.True(t, accountHasOrgAccess(acct, "org_a"))
	assert.True(t, accountHasOrgAccess(acct, "org_b"))
	assert.False(t, accountHasOrgAccess(acct, "org_c"))
	assert.False(t, accountHasOrgAccess(&app.Account{}, "org_a"))
}

func TestAccountAllowed(t *testing.T) {
	employee := &app.Account{AccountType: app.AccountTypeAuth, IsEmployee: true}
	external := &app.Account{AccountType: app.AccountTypeAuth, IsEmployee: false}
	session := &app.Token{TokenType: app.TokenTypeNuon}

	assert.True(t, (&Server{}).tokenAllowed(employee, session))
	assert.True(t, (&Server{}).tokenAllowed(external, session))
	assert.True(t, (&Server{requireEmployee: true}).tokenAllowed(employee, session))
	assert.False(t, (&Server{requireEmployee: true}).tokenAllowed(external, session))
	assert.False(t, (&Server{}).tokenAllowed(employee, &app.Token{TokenType: app.TokenTypeStatic}))
}

func TestResolveOrg(t *testing.T) {
	multi := &app.Account{ID: "acc1", OrgIDs: []string{"org_a", "org_b"}}
	single := &app.Account{ID: "acc2", OrgIDs: []string{"org_only"}}
	s := &Server{}

	t.Run("header org used for this call", func(t *testing.T) {
		assert.Equal(t, "org_b", s.resolveOrg(multi, "org_b"))
	})

	t.Run("a later header is independent", func(t *testing.T) {
		assert.Equal(t, "org_a", s.resolveOrg(multi, "org_a"))
		assert.Equal(t, "org_b", s.resolveOrg(multi, "org_b"))
	})

	t.Run("empty header stays empty", func(t *testing.T) {
		assert.Equal(t, "", s.resolveOrg(multi, ""))
		assert.Equal(t, "", s.resolveOrg(single, ""))
	})

	t.Run("inaccessible header stays empty", func(t *testing.T) {
		assert.Equal(t, "", s.resolveOrg(multi, "org_x"))
		assert.Equal(t, "", s.resolveOrg(single, "org_x"))
	})
}
