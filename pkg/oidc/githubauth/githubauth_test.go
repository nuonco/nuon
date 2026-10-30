package githubauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

type fakeGitHub struct {
	token        string
	installs     []int64
	tokenStatus  int
	tokenBody    string
	gotCode      string
	gotAuthToken string
}

func (f *fakeGitHub) start(t *testing.T) *Client {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Accept") != "application/json" {
			t.Errorf("unexpected token request: %s accept=%q", r.Method, r.Header.Get("Accept"))
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		f.gotCode = r.PostForm.Get("code")
		w.Header().Set("Content-Type", "application/json")
		if f.tokenStatus != 0 {
			w.WriteHeader(f.tokenStatus)
		}
		if f.tokenBody != "" {
			_, _ = w.Write([]byte(f.tokenBody))
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"` + f.token + `","token_type":"bearer"}`))
	})
	mux.HandleFunc("/user/installations", func(w http.ResponseWriter, r *http.Request) {
		f.gotAuthToken = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		var items []string
		for _, id := range f.installs {
			items = append(items, `{"id":`+itoa(id)+`}`)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total_count":` + itoa(int64(len(items))) + `,"installations":[` + strings.Join(items, ",") + `]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &Client{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		HTTPClient:   srv.Client(),
		TokenURL:     srv.URL + "/login/oauth/access_token",
		APIBaseURL:   srv.URL + "/",
	}
}

func itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}

func TestAuthorizeURL(t *testing.T) {
	got := AuthorizeURL("abc", "https://app.example.test/connect/oauth", "n0nce")
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "github.com" || u.Path != "/login/oauth/authorize" {
		t.Fatalf("unexpected url %q", got)
	}
	q := u.Query()
	if q.Get("client_id") != "abc" || q.Get("state") != "n0nce" || q.Get("redirect_uri") != "https://app.example.test/connect/oauth" {
		t.Fatalf("unexpected query %q", u.RawQuery)
	}
	if q.Has("scope") {
		t.Fatalf("scope must not be requested: %q", u.RawQuery)
	}
}

func TestVerifyInstallation(t *testing.T) {
	ctx := context.Background()

	t.Run("owned", func(t *testing.T) {
		gh := &fakeGitHub{token: "user-token", installs: []int64{7, 42}}
		c := gh.start(t)
		if err := c.VerifyInstallation(ctx, "code-1", "42"); err != nil {
			t.Fatal(err)
		}
		if gh.gotCode != "code-1" || gh.gotAuthToken != "user-token" {
			t.Fatalf("code=%q token=%q", gh.gotCode, gh.gotAuthToken)
		}
	})

	t.Run("not owned", func(t *testing.T) {
		gh := &fakeGitHub{token: "user-token", installs: []int64{7}}
		c := gh.start(t)
		err := c.VerifyInstallation(ctx, "code-1", "42")
		if !errors.Is(err, ErrInstallationNotOwned) {
			t.Fatalf("err = %v, want ErrInstallationNotOwned", err)
		}
	})

	t.Run("no installations", func(t *testing.T) {
		gh := &fakeGitHub{token: "user-token"}
		c := gh.start(t)
		if err := c.VerifyInstallation(ctx, "code-1", "42"); !errors.Is(err, ErrInstallationNotOwned) {
			t.Fatalf("err = %v, want ErrInstallationNotOwned", err)
		}
	})

	t.Run("bad code", func(t *testing.T) {
		gh := &fakeGitHub{tokenBody: `{"error":"bad_verification_code","error_description":"The code passed is incorrect or expired."}`}
		c := gh.start(t)
		err := c.VerifyInstallation(ctx, "stale", "42")
		if err == nil || errors.Is(err, ErrInstallationNotOwned) || !strings.Contains(err.Error(), "bad_verification_code") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("token endpoint error", func(t *testing.T) {
		gh := &fakeGitHub{tokenStatus: http.StatusBadGateway, tokenBody: "nope"}
		c := gh.start(t)
		if err := c.VerifyInstallation(ctx, "code-1", "42"); err == nil || errors.Is(err, ErrInstallationNotOwned) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("invalid install id", func(t *testing.T) {
		gh := &fakeGitHub{token: "user-token", installs: []int64{42}}
		c := gh.start(t)
		if err := c.VerifyInstallation(ctx, "code-1", "forty-two"); err == nil {
			t.Fatal("expected error for non-numeric installation id")
		}
		if gh.gotCode != "" {
			t.Fatal("code must not be exchanged for an invalid installation id")
		}
	})

	t.Run("unconfigured", func(t *testing.T) {
		c := &Client{}
		if err := c.VerifyInstallation(ctx, "code-1", "42"); !errors.Is(err, ErrNotConfigured) {
			t.Fatalf("err = %v, want ErrNotConfigured", err)
		}
	})
}
