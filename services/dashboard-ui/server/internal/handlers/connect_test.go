package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type fakeCtlAPI struct {
	callback func(w http.ResponseWriter, r *http.Request)
	orgs     func(w http.ResponseWriter, r *http.Request)
	account  func(w http.ResponseWriter, r *http.Request)
}

func (f *fakeCtlAPI) start(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var h func(http.ResponseWriter, *http.Request)
		switch r.URL.Path {
		case "/v1/vcs/connection-callback":
			h = f.callback
		case "/v1/orgs":
			h = f.orgs
		case "/v1/general/current-user":
			h = f.account
		}
		if h == nil {
			t.Errorf("unexpected ctl-api call: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func jsonResponse(status int, body string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func serve(t *testing.T, register func(e *gin.Engine) error, target string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	e := gin.New()
	if err := register(e); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, target, nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func authed() *http.Cookie { return &http.Cookie{Name: authCookie, Value: "token"} }

func TestParseConnectState(t *testing.T) {
	tests := []struct {
		state          string
		wantOrg        string
		wantOnboarding bool
	}{
		{state: "orgabc123", wantOrg: "orgabc123"},
		{state: "orgabc123:onboarding", wantOrg: "orgabc123", wantOnboarding: true},
		{state: "orgabc123:https://evil.example.com", wantOrg: "orgabc123"},
		{state: "orgabc123:onboarding:extra", wantOrg: "orgabc123"},
		{state: "", wantOrg: ""},
	}
	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			org, onboarding := parseConnectState(tt.state)
			if org != tt.wantOrg || onboarding != tt.wantOnboarding {
				t.Errorf("parseConnectState(%q) = (%q, %v), want (%q, %v)", tt.state, org, onboarding, tt.wantOrg, tt.wantOnboarding)
			}
		})
	}
}

func TestConnectHandle(t *testing.T) {
	ok := jsonResponse(http.StatusCreated, `{"id":"vcsconn123"}`)
	fail := jsonResponse(http.StatusBadRequest, `{"error":"bad install"}`)

	tests := []struct {
		name     string
		state    string
		callback func(http.ResponseWriter, *http.Request)
		noAuth   bool
		want     string
	}{
		{name: "plain success", state: "orgabc123", callback: ok, want: "/orgabc123/apps?vcs-connected=vcsconn123"},
		{name: "plain failure", state: "orgabc123", callback: fail, want: "/orgabc123/apps"},
		{name: "onboarding success", state: "orgabc123:onboarding", callback: ok, want: "/onboarding?vcs-connected=vcsconn123"},
		{name: "onboarding failure", state: "orgabc123:onboarding", callback: fail, want: "/onboarding?vcs-error=1"},
		{name: "onboarding without auth", state: "orgabc123:onboarding", noAuth: true, want: "/onboarding?vcs-error=1"},
		{name: "foreign suffix is ignored", state: "orgabc123:https://evil.example.com", callback: ok, want: "/orgabc123/apps?vcs-connected=vcsconn123"},
		{name: "foreign suffix failure", state: "orgabc123:https://evil.example.com", callback: fail, want: "/orgabc123/apps"},
		{name: "malformed org never reaches the path", state: "/evil.example.com", noAuth: true, want: "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := (&fakeCtlAPI{callback: tt.callback}).start(t)
			h := NewConnectHandler(&internal.Config{APIUrl: api.URL}, zap.NewNop())

			var cookies []*http.Cookie
			if !tt.noAuth {
				cookies = append(cookies, authed())
			}
			target := "/connect?installation_id=42&state=" + url.QueryEscape(tt.state)
			rec := serve(t, h.RegisterRoutes, target, cookies...)

			if rec.Code != http.StatusFound {
				t.Fatalf("status = %d, want 302", rec.Code)
			}
			if got := rec.Header().Get("Location"); got != tt.want {
				t.Errorf("Location = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStartConnectGithubState(t *testing.T) {
	h := NewConnectHandler(&internal.Config{GithubAppName: "Nuon Dev"}, zap.NewNop())

	tests := []struct {
		query string
		want  string
	}{
		{query: "org_id=orgabc123", want: "https://github.com/apps/nuon-dev/installations/new?state=orgabc123"},
		{query: "org_id=orgabc123&onboarding=1", want: "https://github.com/apps/nuon-dev/installations/new?state=orgabc123%3Aonboarding"},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			rec := serve(t, h.RegisterRoutes, "/api/connect-github?"+tt.query)
			if got := rec.Header().Get("Location"); got != tt.want {
				t.Errorf("Location = %q, want %q", got, tt.want)
			}
		})
	}
}
