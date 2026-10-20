package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func newDDProxyTestEngine(t *testing.T, intake string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	e := gin.New()
	if err := newDDProxyHandler(zap.NewNop(), intake).RegisterRoutes(e); err != nil {
		t.Fatal(err)
	}
	return e
}

func ddpRequest(method, ddforward string) *http.Request {
	return httptest.NewRequest(method, "/api/ddp?ddforward="+url.QueryEscape(ddforward), nil)
}

func TestDDProxyForwardsIntakePaths(t *testing.T) {
	var gotPath, gotQuery string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	e := newDDProxyTestEngine(t, upstream.URL)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, ddpRequest(http.MethodPost, "/api/v2/rum?ddsource=browser&x=1"))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "ok")
	}
	if gotPath != "/api/v2/rum" {
		t.Errorf("upstream path = %q, want /api/v2/rum", gotPath)
	}
	if gotQuery != "ddsource=browser&x=1" {
		t.Errorf("upstream query = %q, want ddsource=browser&x=1", gotQuery)
	}
}

func TestDDProxyRejectsUntrustedForwards(t *testing.T) {
	var hits atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
	}))
	defer upstream.Close()

	e := newDDProxyTestEngine(t, upstream.URL)

	for _, ddforward := range []string{
		"",
		".evil.com/x",
		"@evil.com/x",
		"//evil.com/api/v2/rum",
		"http://evil.com/api/v2/rum",
		"/api/v2/rum/../../admin",
		"/other",
	} {
		t.Run(ddforward, func(t *testing.T) {
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, ddpRequest(http.MethodPost, ddforward))
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}

	if n := hits.Load(); n != 0 {
		t.Errorf("upstream received %d requests, want 0", n)
	}
}

func TestDDProxyOnlyAcceptsPost(t *testing.T) {
	var hits atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
	}))
	defer upstream.Close()

	e := newDDProxyTestEngine(t, upstream.URL)
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, ddpRequest(method, "/api/v2/rum"))
		if rec.Code < 400 {
			t.Errorf("%s status = %d, want an error", method, rec.Code)
		}
	}

	if n := hits.Load(); n != 0 {
		t.Errorf("upstream received %d requests, want 0", n)
	}
}

func TestDDProxyDoesNotFollowRedirects(t *testing.T) {
	var internalHits atomic.Int32
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		internalHits.Add(1)
	}))
	defer internal.Close()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, internal.URL+"/secret", http.StatusFound)
	}))
	defer upstream.Close()

	e := newDDProxyTestEngine(t, upstream.URL)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, ddpRequest(http.MethodPost, "/api/v2/logs"))

	if rec.Code != http.StatusFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if n := internalHits.Load(); n != 0 {
		t.Errorf("redirect target received %d requests, want 0", n)
	}
}
