package server

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestAccountHasOrgAccess(t *testing.T) {
	acct := &app.Account{OrgIDs: []string{"org_a", "org_b"}}
	assert.True(t, accountHasOrgAccess(acct, "org_a"))
	assert.True(t, accountHasOrgAccess(acct, "org_b"))
	assert.False(t, accountHasOrgAccess(acct, "org_c"))
	assert.False(t, accountHasOrgAccess(&app.Account{}, "org_a"))
}

func newTestServer() *Server {
	return &Server{orgSelections: make(map[string]*orgSelection)}
}

func TestResolveOrg(t *testing.T) {
	multi := &app.Account{ID: "acc1", OrgIDs: []string{"org_a", "org_b"}}
	single := &app.Account{ID: "acc2", OrgIDs: []string{"org_only"}}

	t.Run("sole org auto-selected", func(t *testing.T) {
		s := newTestServer()
		assert.Equal(t, "org_only", s.resolveOrg(single, "tok1", ""))
		assert.Equal(t, "org_only", s.orgSelections["tok1"].orgID)
	})

	t.Run("multi org, none selected -> empty", func(t *testing.T) {
		s := newTestServer()
		assert.Equal(t, "", s.resolveOrg(multi, "tok1", ""))
	})

	t.Run("valid header org used and remembered", func(t *testing.T) {
		s := newTestServer()
		assert.Equal(t, "org_b", s.resolveOrg(multi, "tok1", "org_b"))
		assert.Equal(t, "org_b", s.orgSelections["tok1"].orgID)
	})

	t.Run("inaccessible header org ignored", func(t *testing.T) {
		s := newTestServer()
		assert.Equal(t, "", s.resolveOrg(multi, "tok1", "org_x"))
	})

	t.Run("prior selection wins over header", func(t *testing.T) {
		s := newTestServer()
		s.setOrgSelection("tok1", "org_a")
		assert.Equal(t, "org_a", s.resolveOrg(multi, "tok1", "org_b"))
	})

	t.Run("stale selection no longer accessible falls back", func(t *testing.T) {
		s := newTestServer()
		s.setOrgSelection("tok1", "org_gone")
		assert.Equal(t, "org_b", s.resolveOrg(multi, "tok1", "org_b"))
	})

	t.Run("selections are per token", func(t *testing.T) {
		s := newTestServer()
		s.setOrgSelection("tok1", "org_a")
		assert.Equal(t, "org_a", s.resolveOrg(multi, "tok1", ""))
		assert.Equal(t, "", s.resolveOrg(multi, "tok2", ""))
	})
}

func TestEvictStaleOrgSelections(t *testing.T) {
	s := newTestServer()
	s.orgSelections["fresh"] = &orgSelection{orgID: "org_a", lastSeen: time.Now()}
	s.orgSelections["stale"] = &orgSelection{orgID: "org_b", lastSeen: time.Now().Add(-2 * orgSelectionTTL)}

	s.evictStaleOrgSelections()

	_, freshOK := s.orgSelections["fresh"]
	_, staleOK := s.orgSelections["stale"]
	assert.True(t, freshOK, "fresh selection should survive")
	assert.False(t, staleOK, "stale selection should be evicted")
}

func TestTouchOrgSelectionKeepsAlive(t *testing.T) {
	s := newTestServer()
	s.orgSelections["tok1"] = &orgSelection{orgID: "org_a", lastSeen: time.Now().Add(-2 * orgSelectionTTL)}
	s.touchOrgSelection("tok1")
	s.evictStaleOrgSelections()
	_, ok := s.orgSelections["tok1"]
	assert.True(t, ok, "touched selection should not be evicted")
}
