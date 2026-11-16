package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/pkg/nuonauth"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestLoopbackLoginEnabled(t *testing.T) {
	require.True(t, (&Service{cfg: &config.Config{Interactive: true}}).loopbackLoginEnabled())
	require.False(t, (&Service{cfg: &config.Config{Interactive: false}}).loopbackLoginEnabled())
	require.False(t, (&Service{}).loopbackLoginEnabled())
}

func TestLoopbackSupportedByCtlAPI(t *testing.T) {
	require.False(t, loopbackSupportedByCtlAPI("0.19.1288"))
	require.True(t, loopbackSupportedByCtlAPI("0.19.1289"))
	require.True(t, loopbackSupportedByCtlAPI("0.19.1291"))
	require.True(t, loopbackSupportedByCtlAPI("8fdb54d"))
	require.True(t, loopbackSupportedByCtlAPI("development"))
	require.True(t, loopbackSupportedByCtlAPI(""))

	ctx := context.Background()
	require.True(t, (&Service{}).ctlAPISupportsLoopback(ctx))
	require.False(t, (&Service{api: &versionAPI{err: errors.New("dial")}}).ctlAPISupportsLoopback(ctx))
	require.False(t, (&Service{api: &versionAPI{version: &nuon.APIVersion{Version: "0.19.1109"}}}).ctlAPISupportsLoopback(ctx))
	require.True(t, (&Service{api: &versionAPI{version: &nuon.APIVersion{Version: "8fdb54d"}}}).ctlAPISupportsLoopback(ctx))
}

const testCompleteURL = "https://auth.example.test/cli/complete"

type loopbackOutcome struct {
	result *LoginResult
	err    error
}

func startLoopback(t *testing.T, state string, exchange loopbackExchange) (net.Listener, <-chan loopbackOutcome) {
	t.Helper()
	ln := listenLocal(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	done := make(chan loopbackOutcome, 1)
	go func() {
		result, err := serveLoopback(ctx, ln, state, testCompleteURL, exchange)
		done <- loopbackOutcome{result, err}
	}()
	return ln, done
}

func TestLoopbackCallbackRedirectsAfterExchange(t *testing.T) {
	var gotCode string
	ln, done := startLoopback(t, "state-1", func(_ context.Context, code string) (*LoginResult, error) {
		gotCode = code
		return &LoginResult{AccessToken: "tok", DisplayName: "ada@example.com"}, nil
	})

	resp := getCallback(t, ln, "/callback?code=abc&state=state-1")
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusFound, resp.StatusCode)
	require.Equal(t, testCompleteURL+"?result=success", resp.Header.Get("Location"))
	require.Equal(t, "no-referrer", resp.Header.Get("Referrer-Policy"))

	got := <-done
	require.NoError(t, got.err)
	require.Equal(t, "abc", gotCode)
	require.Equal(t, "tok", got.result.AccessToken)
}

func TestLoopbackCallbackRedirectsFailedExchange(t *testing.T) {
	ln, done := startLoopback(t, "state-1", func(context.Context, string) (*LoginResult, error) {
		return nil, errors.New("invalid_grant")
	})

	resp := getCallback(t, ln, "/callback?code=abc&state=state-1")
	require.NoError(t, resp.Body.Close())
	require.Equal(t, testCompleteURL+"?result=failed", resp.Header.Get("Location"))
	require.Error(t, (<-done).err)
}

func TestLoopbackCallbackRedirectsDenied(t *testing.T) {
	ln, done := startLoopback(t, "state-1", func(context.Context, string) (*LoginResult, error) {
		t.Fatal("exchange should not run when the user denied")
		return nil, nil
	})

	resp := getCallback(t, ln, "/callback?error=access_denied&state=state-1")
	require.NoError(t, resp.Body.Close())
	require.Equal(t, testCompleteURL+"?result=denied", resp.Header.Get("Location"))
	require.Error(t, (<-done).err)
}

func TestLoopbackCallbackIgnoresWrongState(t *testing.T) {
	ln, done := startLoopback(t, "expected", func(_ context.Context, code string) (*LoginResult, error) {
		return &LoginResult{AccessToken: code}, nil
	})

	bad := getCallback(t, ln, "/callback?code=evil&state=other")
	require.NoError(t, bad.Body.Close())
	require.Equal(t, http.StatusBadRequest, bad.StatusCode)

	good := getCallback(t, ln, "/callback?code=real&state=expected")
	require.NoError(t, good.Body.Close())
	require.Equal(t, http.StatusFound, good.StatusCode)

	got := <-done
	require.NoError(t, got.err)
	require.Equal(t, "real", got.result.AccessToken)
}

func TestLoopbackLoginExchangesCode(t *testing.T) {
	var sawClientID, sawCode, sawVerifier string
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			http.NotFound(w, r)
			return
		}
		_ = r.ParseForm()
		sawClientID = r.PostForm.Get("client_id")
		sawCode = r.PostForm.Get("code")
		sawVerifier = r.PostForm.Get("code_verifier")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "nuon-token",
			"token_type":   "Bearer",
			"expires_in":   86400,
			"email":        "ada@example.com",
		})
	}))
	t.Cleanup(tokenSrv.Close)

	orig := openBrowser
	openBrowser = func(raw string) error {
		authURL, err := url.Parse(raw)
		require.NoError(t, err)
		redirect, err := url.Parse(authURL.Query().Get("redirect_uri"))
		require.NoError(t, err)
		callback := &url.URL{
			Scheme: "http",
			Host:   redirect.Host,
			Path:   redirect.Path,
			RawQuery: url.Values{
				"code":  {"the-code"},
				"state": {authURL.Query().Get("state")},
			}.Encode(),
		}
		go func() {
			deadline := time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) {
				resp, err := http.Get(callback.String())
				if err == nil {
					resp.Body.Close()
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
		}()
		return nil
	}
	t.Cleanup(func() { openBrowser = orig })

	svc := &Service{
		cfg:                 &config.Config{Interactive: true},
		authBaseURLOverride: tokenSrv.URL,
	}
	result, err := svc.loginWithNuonAuth(context.Background(), &models.ServiceCLIConfig{RootDomain: "example.test"})
	require.NoError(t, err)
	require.Equal(t, "nuon-token", result.AccessToken)
	require.Equal(t, "ada@example.com", result.DisplayName)
	require.Equal(t, nuonauth.CLIClientID, sawClientID)
	require.Equal(t, "the-code", sawCode)
	require.NotEmpty(t, sawVerifier)
}

func TestInteractiveLoginUsesDeviceCodeBeforeLoopbackRelease(t *testing.T) {
	opened := false
	orig := openBrowser
	openBrowser = func(string) error {
		opened = true
		return nil
	}
	t.Cleanup(func() { openBrowser = orig })

	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/device/authorize":
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(NuonAuthDeviceAuthorizationResponse{
				DeviceCode:      "dev",
				UserCode:        "ABCD-EFGH",
				VerificationURI: "https://auth.example.test/device",
				ExpiresIn:       30,
				Interval:        1,
			}))
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"expired_token","error_description":"device code has expired"}`))
		}
	}))
	t.Cleanup(srv.Close)

	svc := &Service{
		cfg:                 &config.Config{Interactive: true},
		authBaseURLOverride: srv.URL,
		api:                 &versionAPI{version: &nuon.APIVersion{Version: "0.19.1288"}},
	}
	_, err := svc.loginWithNuonAuth(context.Background(), &models.ServiceCLIConfig{RootDomain: "example.test"})
	require.Error(t, err)
	require.True(t, opened)
	require.Equal(t, []string{"/device/authorize", "/device/token"}, paths)
}

func TestInteractiveLoginUsesLoopbackForGitHashVersion(t *testing.T) {
	var sawPath string
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		if r.URL.Path != "/oauth/token" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "nuon-token",
			"token_type":   "Bearer",
			"expires_in":   86400,
			"email":        "ada@example.com",
		})
	}))
	t.Cleanup(tokenSrv.Close)

	orig := openBrowser
	openBrowser = func(raw string) error {
		authURL, err := url.Parse(raw)
		require.NoError(t, err)
		redirect, err := url.Parse(authURL.Query().Get("redirect_uri"))
		require.NoError(t, err)
		callback := &url.URL{
			Scheme: "http",
			Host:   redirect.Host,
			Path:   redirect.Path,
			RawQuery: url.Values{
				"code":  {"the-code"},
				"state": {authURL.Query().Get("state")},
			}.Encode(),
		}
		go func() {
			deadline := time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) {
				resp, err := http.Get(callback.String())
				if err == nil {
					resp.Body.Close()
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
		}()
		return nil
	}
	t.Cleanup(func() { openBrowser = orig })

	svc := &Service{
		cfg:                 &config.Config{Interactive: true},
		authBaseURLOverride: tokenSrv.URL,
		api:                 &versionAPI{version: &nuon.APIVersion{Version: "8fdb54d"}},
	}
	result, err := svc.loginWithNuonAuth(context.Background(), &models.ServiceCLIConfig{RootDomain: "example.test"})
	require.NoError(t, err)
	require.Equal(t, "nuon-token", result.AccessToken)
	require.Equal(t, "/oauth/token", sawPath)
}

func TestNonInteractiveLoginUsesDeviceAuthorization(t *testing.T) {
	opened := false
	orig := openBrowser
	openBrowser = func(string) error {
		opened = true
		return nil
	}
	t.Cleanup(func() { openBrowser = orig })

	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/device/authorize":
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(NuonAuthDeviceAuthorizationResponse{
				DeviceCode:      "dev",
				UserCode:        "ABCD-EFGH",
				VerificationURI: "https://auth.example.test/device",
				ExpiresIn:       30,
				Interval:        1,
			}))
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"expired_token","error_description":"device code has expired"}`))
		}
	}))
	t.Cleanup(srv.Close)

	svc := &Service{
		cfg:                 &config.Config{Interactive: false},
		authBaseURLOverride: srv.URL,
	}
	_, err := svc.loginWithNuonAuth(context.Background(), &models.ServiceCLIConfig{RootDomain: "example.test"})
	require.Error(t, err)
	require.True(t, opened)
	require.Equal(t, []string{"/device/authorize", "/device/token"}, paths)
}

func TestInteractiveLoginFallsBackWhenLoopbackCannotBind(t *testing.T) {
	orig := openBrowser
	openBrowser = func(rawURL string) error { return nil }
	t.Cleanup(func() { openBrowser = orig })

	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/device/authorize":
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(NuonAuthDeviceAuthorizationResponse{
				DeviceCode:      "dev",
				UserCode:        "ABCD-EFGH",
				VerificationURI: "https://auth.example.test/device",
				ExpiresIn:       30,
				Interval:        1,
			}))
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"expired_token","error_description":"device code has expired"}`))
		}
	}))
	t.Cleanup(srv.Close)

	svc := &Service{
		cfg:                 &config.Config{Interactive: true},
		authBaseURLOverride: srv.URL,
		listenLoopbackFn: func() (net.Listener, error) {
			return nil, errLoopbackUnavailable
		},
	}
	_, err := svc.loginWithNuonAuth(context.Background(), &models.ServiceCLIConfig{RootDomain: "example.test"})
	require.Error(t, err)
	require.NotContains(t, paths, "/oauth/authorize")
	require.Equal(t, "/device/authorize", paths[0])
}

type versionAPI struct {
	nuon.Client
	version *nuon.APIVersion
	err     error
}

func (v *versionAPI) GetVersion(context.Context) (*nuon.APIVersion, error) {
	return v.version, v.err
}

func listenLocal(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	return ln
}

func getCallback(t *testing.T, ln net.Listener, path string) *http.Response {
	t.Helper()
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	url := "http://" + ln.Addr().String() + path
	var resp *http.Response
	var err error
	for range 20 {
		resp, err = client.Get(url)
		if err == nil {
			return resp
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.NoError(t, err)
	return resp
}
