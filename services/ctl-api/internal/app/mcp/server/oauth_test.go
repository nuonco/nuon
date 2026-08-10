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

func timeNowForTest() time.Time { return time.Now() }

func newTestServer() *Server {
	return &Server{sessions: make(map[string]*authResult)}
}

func TestResolveOrg(t *testing.T) {
	multi := &app.Account{ID: "acc1", OrgIDs: []string{"org_a", "org_b"}}
	single := &app.Account{ID: "acc2", OrgIDs: []string{"org_only"}}

	t.Run("sole org auto-selected", func(t *testing.T) {
		s := newTestServer()
		assert.Equal(t, "org_only", s.resolveOrg(single, "sess1", ""))
		// persisted to session
		assert.Equal(t, "org_only", s.sessions["sess1"].OrgID)
	})

	t.Run("multi org, none selected -> empty", func(t *testing.T) {
		s := newTestServer()
		assert.Equal(t, "", s.resolveOrg(multi, "sess1", ""))
	})

	t.Run("valid header org used and persisted", func(t *testing.T) {
		s := newTestServer()
		assert.Equal(t, "org_b", s.resolveOrg(multi, "sess1", "org_b"))
		assert.Equal(t, "org_b", s.sessions["sess1"].OrgID)
	})

	t.Run("inaccessible header org ignored", func(t *testing.T) {
		s := newTestServer()
		assert.Equal(t, "", s.resolveOrg(multi, "sess1", "org_x"))
	})

	t.Run("session selection wins over header", func(t *testing.T) {
		s := newTestServer()
		s.setSessionOrg("sess1", "acc1", "org_a")
		assert.Equal(t, "org_a", s.resolveOrg(multi, "sess1", "org_b"))
	})

	t.Run("stale session org no longer accessible falls back", func(t *testing.T) {
		s := newTestServer()
		s.setSessionOrg("sess1", "acc1", "org_gone")
		// account only has org_a/org_b now; header picks org_b
		assert.Equal(t, "org_b", s.resolveOrg(multi, "sess1", "org_b"))
	})
}

func TestEvictStaleSessions(t *testing.T) {
	s := newTestServer()
	s.sessions["fresh"] = &authResult{OrgID: "org_a", AccountID: "acc", lastSeen: timeNowForTest()}
	s.sessions["stale"] = &authResult{OrgID: "org_b", AccountID: "acc", lastSeen: timeNowForTest().Add(-2 * mcpSessionTTL)}

	s.evictStaleSessions()

	_, freshOK := s.sessions["fresh"]
	_, staleOK := s.sessions["stale"]
	assert.True(t, freshOK, "fresh session should survive")
	assert.False(t, staleOK, "stale session should be evicted")
}

func TestTouchSessionKeepsAlive(t *testing.T) {
	s := newTestServer()
	s.sessions["s1"] = &authResult{OrgID: "org_a", AccountID: "acc", lastSeen: timeNowForTest().Add(-2 * mcpSessionTTL)}
	s.touchSession("s1")
	s.evictStaleSessions()
	_, ok := s.sessions["s1"]
	assert.True(t, ok, "touched session should not be evicted")
}
