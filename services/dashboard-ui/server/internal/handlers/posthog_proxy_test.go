package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

func newTestPostHogProxy(t *testing.T, upstreamURL string) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	h := NewPostHogProxyHandler(&internal.Config{PostHogKey: "test-key", PostHogHost: upstreamURL}, zap.NewNop())
	e := gin.New()
	if err := h.RegisterRoutes(e); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)
	return srv
}

func TestPostHogProxyAllowsKnownPaths(t *testing.T) {
	var gotPath, gotQuery string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(upstream.Close)

	srv := newTestPostHogProxy(t, upstream.URL)

	resp, err := http.Post(srv.URL+"/ingest/e/?ip=1", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if gotPath != "/e/" {
		t.Errorf("upstream path = %q, want /e/", gotPath)
	}
	if gotQuery != "ip=1" {
		t.Errorf("upstream query = %q, want ip=1", gotQuery)
	}
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
}

func TestPostHogProxyRejectsPathsOutsideAllowlist(t *testing.T) {
	var called bool
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	srv := newTestPostHogProxy(t, upstream.URL)

	resp, err := http.Get(srv.URL + "/ingest/api/users/@me/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	if called {
		t.Error("upstream was called for a disallowed path")
	}
}

func TestPostHogProxyRejectsPathTraversal(t *testing.T) {
	var called bool
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	srv := newTestPostHogProxy(t, upstream.URL)

	resp, err := http.Get(srv.URL + "/ingest/e/..%2f..%2fadmin")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	if called {
		t.Error("upstream was called for a path-traversal attempt")
	}
}

func TestPostHogProxyEscapesEmbeddedQueryDelimiter(t *testing.T) {
	var gotPath, gotQuery string
	var called bool
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	srv := newTestPostHogProxy(t, upstream.URL)

	resp, err := http.Get(srv.URL + "/ingest/e/foo%3Finjected=1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if !called {
		t.Fatal("upstream was never called")
	}
	if gotQuery != "" {
		t.Errorf("upstream query = %q, want empty (no injected query)", gotQuery)
	}
	if !strings.Contains(gotPath, "foo?injected=1") {
		t.Errorf("upstream path = %q, want the literal '?' preserved in the path", gotPath)
	}
}

func TestPostHogProxyDoesNotFollowRedirects(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "https://evil.example.com/steal")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(upstream.Close)

	srv := newTestPostHogProxy(t, upstream.URL)

	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(srv.URL + "/ingest/e/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("status = %d, want 302", resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != "" {
		t.Errorf("Location = %q, want it dropped from the response", got)
	}
}

func TestPostHogProxyRejectsOversizedBody(t *testing.T) {
	var called bool
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	srv := newTestPostHogProxy(t, upstream.URL)

	body := strings.NewReader(strings.Repeat("a", maxIngestBodyBytes+1))
	resp, err := http.Post(srv.URL+"/ingest/e/", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", resp.StatusCode)
	}
	if called {
		t.Error("upstream was called with an oversized body")
	}
}

func TestPostHogProxyEnforcesConcurrencyLimit(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	gin.SetMode(gin.TestMode)
	h := NewPostHogProxyHandler(&internal.Config{PostHogKey: "test-key", PostHogHost: upstream.URL}, zap.NewNop())
	for i := 0; i < ingestConcurrencyLimit; i++ {
		h.inflight <- struct{}{}
	}

	e := gin.New()
	if err := h.RegisterRoutes(e); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/ingest/e/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", resp.StatusCode)
	}
}

func TestPostHogProxyNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewPostHogProxyHandler(&internal.Config{}, zap.NewNop())
	e := gin.New()
	if err := h.RegisterRoutes(e); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/ingest/e/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
}
