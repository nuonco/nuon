package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type fakeCtlAPI struct {
	callback func(w http.ResponseWriter, r *http.Request)
	orgs     func(w http.ResponseWriter, r *http.Request)
	org      func(w http.ResponseWriter, r *http.Request)
	account  func(w http.ResponseWriter, r *http.Request)
	validate func(w http.ResponseWriter, r *http.Request)
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
		case "/v1/orgs/current":
			h = f.org
		case "/v1/general/current-user":
			h = f.account
		case "/v1/auth/validate":
			h = f.validate
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

func testConfig(apiURL string) *internal.Config {
	return &internal.Config{
		APIUrl:         apiURL,
		AppUrl:         "http://app.example.test",
		AuthServiceUrl: "http://auth.example.test",
		GithubAppName:  "Nuon Dev",
	}
}

func allowToken() func(http.ResponseWriter, *http.Request) {
	return jsonResponse(http.StatusOK, `{"valid":true}`)
}

func allowOrg() func(http.ResponseWriter, *http.Request) {
	return jsonResponse(http.StatusOK, `{"id":"orgabc123"}`)
}

func startConnect(t *testing.T, h *ConnectHandler, query string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	return serve(t, h.RegisterRoutes, "/api/connect-github?"+query, cookies...)
}

func stateCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, c := range rec.Result().Cookies() {
		if c.Name == connectStateCookie && c.Value != "" {
			return c
		}
	}
	t.Fatal("connect state cookie was not set")
	return nil
}

func TestConnectHandle(t *testing.T) {
	ok := jsonResponse(http.StatusCreated, `{"id":"vcsconn123"}`)
	fail := jsonResponse(http.StatusBadRequest, `{"error":"bad install"}`)

	tests := []struct {
		name       string
		onboarding bool
		callback   func(http.ResponseWriter, *http.Request)
		mutate     func(t *testing.T, h *ConnectHandler, state *http.Cookie, nonce *string)
		noAuth     bool
		want       string
	}{
		{name: "plain success", callback: ok, want: "/orgabc123/apps?vcs-connected=vcsconn123"},
		{name: "plain failure", callback: fail, want: "/orgabc123/apps"},
		{name: "onboarding success", onboarding: true, callback: ok, want: "/onboarding?vcs-connected=vcsconn123"},
		{name: "onboarding failure", onboarding: true, callback: fail, want: "/onboarding?vcs-error=1"},
		{name: "onboarding without auth", onboarding: true, noAuth: true, want: "/onboarding?vcs-error=1"},
		{
			name:     "tampered state is ignored",
			callback: ok,
			mutate: func(t *testing.T, h *ConnectHandler, state *http.Cookie, nonce *string) {
				*nonce = "someone-elses-nonce"
			},
			want: "/orgabc123/apps",
		},
		{
			name: "other session is ignored",
			mutate: func(t *testing.T, h *ConnectHandler, state *http.Cookie, nonce *string) {
				opened, err := openConnectState(state.Value)
				if err != nil {
					t.Fatal(err)
				}
				opened.TokenHash = tokenHash("other-token")
				sealed, err := sealConnectState(opened)
				if err != nil {
					t.Fatal(err)
				}
				state.Value = sealed
			},
			want: "/orgabc123/apps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := (&fakeCtlAPI{callback: tt.callback, org: allowOrg(), validate: allowToken()}).start(t)
			h := NewConnectHandler(testConfig(api.URL), zap.NewNop())

			query := "org_id=orgabc123"
			if tt.onboarding {
				query += "&onboarding=1"
			}
			started := startConnect(t, h, query, authed())
			if started.Code != http.StatusFound {
				t.Fatalf("start status = %d, want 302", started.Code)
			}
			location, err := url.Parse(started.Header().Get("Location"))
			if err != nil {
				t.Fatal(err)
			}
			nonce := location.Query().Get("state")
			if nonce == "" || nonce == "orgabc123" || strings.Contains(nonce, "onboarding") {
				t.Fatalf("github state = %q", nonce)
			}
			state := stateCookie(t, started)
			if tt.mutate != nil {
				tt.mutate(t, h, state, &nonce)
			}

			var cookies []*http.Cookie
			if !tt.noAuth {
				cookies = append(cookies, authed())
			}
			cookies = append(cookies, state)
			target := "/connect?installation_id=42&state=" + url.QueryEscape(nonce)
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

func TestConnectStateIsSingleUseInTheBrowser(t *testing.T) {
	api := (&fakeCtlAPI{
		callback: jsonResponse(http.StatusCreated, `{"id":"vcsconn123"}`),
		org:      allowOrg(),
		validate: allowToken(),
	}).start(t)
	h := NewConnectHandler(testConfig(api.URL), zap.NewNop())
	started := startConnect(t, h, "org_id=orgabc123", authed())
	nonce := mustState(t, started)
	state := stateCookie(t, started)

	first := serve(t, h.RegisterRoutes, "/connect?installation_id=42&state="+url.QueryEscape(nonce), authed(), state)
	if first.Header().Get("Location") != "/orgabc123/apps?vcs-connected=vcsconn123" {
		t.Fatalf("first Location = %q", first.Header().Get("Location"))
	}
	cleared := false
	for _, c := range first.Result().Cookies() {
		if c.Name == connectStateCookie && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatal("connect state cookie was not cleared")
	}

	second := serve(t, h.RegisterRoutes, "/connect?installation_id=42&state="+url.QueryEscape(nonce), authed())
	if second.Header().Get("Location") != "/" {
		t.Fatalf("replay Location = %q", second.Header().Get("Location"))
	}
}

func TestConnectStateExpires(t *testing.T) {
	h := NewConnectHandler(testConfig("http://ctl.example.test"), zap.NewNop())
	sealed, err := sealConnectState(connectState{
		Nonce:     "nonce",
		OrgID:     "orgabc123",
		TokenHash: tokenHash("token"),
		ExpiresAt: time.Now().Add(-time.Minute).Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	rec := serve(t, h.RegisterRoutes, "/connect?installation_id=42&state=nonce", authed(), &http.Cookie{
		Name:  connectStateCookie,
		Value: sealed,
	})
	if rec.Header().Get("Location") != "/orgabc123/apps" {
		t.Fatalf("Location = %q", rec.Header().Get("Location"))
	}
}

func TestStartConnectGithub(t *testing.T) {
	api := (&fakeCtlAPI{org: allowOrg(), validate: allowToken()}).start(t)
	h := NewConnectHandler(testConfig(api.URL), zap.NewNop())

	t.Run("requires a session", func(t *testing.T) {
		rec := startConnect(t, h, "org_id=orgabc123")
		if !strings.HasPrefix(rec.Header().Get("Location"), "http://auth.example.test/?url=") {
			t.Fatalf("Location = %q", rec.Header().Get("Location"))
		}
	})

	t.Run("rejects an org the session cannot access", func(t *testing.T) {
		denied := (&fakeCtlAPI{
			org:      jsonResponse(http.StatusForbidden, `{"error":"nope"}`),
			validate: allowToken(),
		}).start(t)
		deniedHandler := NewConnectHandler(testConfig(denied.URL), zap.NewNop())
		rec := startConnect(t, deniedHandler, "org_id=orgabc123", authed())
		if rec.Header().Get("Location") != "/orgabc123/apps" {
			t.Fatalf("Location = %q", rec.Header().Get("Location"))
		}
	})

	t.Run("rejects a bad org id before calling the api", func(t *testing.T) {
		rec := startConnect(t, h, "org_id=../evil", authed())
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d", rec.Code)
		}
	})
}

func oauthConfig(apiURL string) *internal.Config {
	cfg := testConfig(apiURL)
	cfg.GithubAppClientID = "Iv1.clientid"
	return cfg
}

// callbackBody is what the dashboard sends ctl-api for a connection callback.
type callbackBody struct {
	OrgID           string `json:"org_id"`
	GithubInstallID string `json:"github_install_id"`
	Code            string `json:"code"`
}

func captureCallback(t *testing.T, got *callbackBody) func(http.ResponseWriter, *http.Request) {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(got); err != nil {
			t.Errorf("decode callback body: %v", err)
		}
		jsonResponse(http.StatusCreated, `{"id":"vcsconn123"}`)(w, r)
	}
}

func TestConnectWithOAuthClient(t *testing.T) {
	var got callbackBody
	api := (&fakeCtlAPI{callback: captureCallback(t, &got), org: allowOrg(), validate: allowToken()}).start(t)
	h := NewConnectHandler(oauthConfig(api.URL), zap.NewNop())

	started := startConnect(t, h, "org_id=orgabc123", authed())
	installNonce := mustState(t, started)
	installState := stateCookie(t, started)

	// GitHub setup URL: must bounce to the authorize page, not ctl-api.
	setup := serve(t, h.RegisterRoutes, "/connect?installation_id=42&state="+url.QueryEscape(installNonce), authed(), installState)
	if setup.Code != http.StatusFound {
		t.Fatalf("setup status = %d", setup.Code)
	}
	authorize, err := url.Parse(setup.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if authorize.Host != "github.com" || authorize.Path != "/login/oauth/authorize" {
		t.Fatalf("setup Location = %q", setup.Header().Get("Location"))
	}
	if q := authorize.Query(); q.Get("client_id") != "Iv1.clientid" || q.Get("redirect_uri") != "http://app.example.test/connect/oauth" {
		t.Fatalf("authorize query = %q", authorize.RawQuery)
	}
	oauthNonce := authorize.Query().Get("state")
	if oauthNonce == "" || oauthNonce == installNonce {
		t.Fatalf("oauth state = %q (install state %q)", oauthNonce, installNonce)
	}
	oauthState := stateCookie(t, setup)
	if got.GithubInstallID != "" {
		t.Fatal("ctl-api must not be called before the user authorizes")
	}

	// GitHub callback URL: forwards code and the stashed installation id.
	done := serve(t, h.RegisterRoutes, "/connect/oauth?code=c0de&state="+url.QueryEscape(oauthNonce), authed(), oauthState)
	if loc := done.Header().Get("Location"); loc != "/orgabc123/apps?vcs-connected=vcsconn123" {
		t.Fatalf("oauth Location = %q", loc)
	}
	if got.Code != "c0de" || got.GithubInstallID != "42" || got.OrgID != "orgabc123" {
		t.Fatalf("callback body = %+v", got)
	}

	t.Run("install nonce cannot finish the oauth step", func(t *testing.T) {
		got = callbackBody{}
		rec := serve(t, h.RegisterRoutes, "/connect/oauth?code=c0de&state="+url.QueryEscape(installNonce), authed(), oauthState)
		if loc := rec.Header().Get("Location"); loc != "/orgabc123/apps" {
			t.Fatalf("Location = %q", loc)
		}
		if got.Code != "" {
			t.Fatal("ctl-api must not be called with a mismatched state")
		}
	})

	t.Run("oauth step without a code", func(t *testing.T) {
		got = callbackBody{}
		rec := serve(t, h.RegisterRoutes, "/connect/oauth?state="+url.QueryEscape(oauthNonce), authed(), oauthState)
		if loc := rec.Header().Get("Location"); loc != "/orgabc123/apps" {
			t.Fatalf("Location = %q", loc)
		}
		if got.Code != "" {
			t.Fatal("ctl-api must not be called without a code")
		}
	})

	t.Run("oauth step without a stashed installation", func(t *testing.T) {
		got = callbackBody{}
		// A cookie from the start step has no installation id yet.
		rec := serve(t, h.RegisterRoutes, "/connect/oauth?code=c0de&state="+url.QueryEscape(installNonce), authed(), installState)
		if loc := rec.Header().Get("Location"); loc != "/orgabc123/apps" {
			t.Fatalf("Location = %q", loc)
		}
		if got.Code != "" {
			t.Fatal("ctl-api must not be called without an installation id")
		}
	})

	t.Run("forged installation id is rejected before the oauth hop", func(t *testing.T) {
		started := startConnect(t, h, "org_id=orgabc123", authed())
		nonce := mustState(t, started)
		rec := serve(t, h.RegisterRoutes, "/connect?installation_id=../evil&state="+url.QueryEscape(nonce), authed(), stateCookie(t, started))
		if loc := rec.Header().Get("Location"); loc != "/orgabc123/apps" {
			t.Fatalf("Location = %q", loc)
		}
	})
}

func TestConnectWithoutOAuthClientSendsNoCode(t *testing.T) {
	var got callbackBody
	api := (&fakeCtlAPI{callback: captureCallback(t, &got), org: allowOrg(), validate: allowToken()}).start(t)
	h := NewConnectHandler(testConfig(api.URL), zap.NewNop())

	started := startConnect(t, h, "org_id=orgabc123", authed())
	nonce := mustState(t, started)
	rec := serve(t, h.RegisterRoutes, "/connect?installation_id=42&state="+url.QueryEscape(nonce), authed(), stateCookie(t, started))
	if loc := rec.Header().Get("Location"); loc != "/orgabc123/apps?vcs-connected=vcsconn123" {
		t.Fatalf("Location = %q", loc)
	}
	if got.Code != "" || got.GithubInstallID != "42" {
		t.Fatalf("callback body = %+v", got)
	}
}

func mustState(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	location, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	nonce := location.Query().Get("state")
	if nonce == "" {
		t.Fatal("missing state")
	}
	return nonce
}
