package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

// The kafka UI proxy relies on kafbat serving under SERVER_SERVLET_CONTEXT_PATH,
// so the request path must reach it unchanged — any prefix rewriting here would
// 404 every asset. Accept-Encoding has to survive for the same reason the body is
// left alone, and the session cookie must not be forwarded to the upstream.
func TestPassthroughProxy(t *testing.T) {
	const upstream = "kafka-ui.kafka.svc.cluster.local:8080"

	h := NewProxyHandler(&internal.Config{KafkaUIUrl: "http://" + upstream}, zap.NewNop())
	proxy := h.newPassthroughProxy(h.cfg.KafkaUIUrl)

	req := httptest.NewRequest(http.MethodGet, "/admin/kafka/api/clusters/nuon-stage/topics", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.AddCookie(&http.Cookie{Name: authCookie, Value: "session-token"})
	proxy.Director(req)

	if req.URL.Host != upstream {
		t.Errorf("host = %q, want %q", req.URL.Host, upstream)
	}
	if req.Host != upstream {
		t.Errorf("req.Host = %q, want %q", req.Host, upstream)
	}
	if want := "/admin/kafka/api/clusters/nuon-stage/topics"; req.URL.Path != want {
		t.Errorf("path = %q, want %q", req.URL.Path, want)
	}
	if got := req.Header.Get("Accept-Encoding"); got != "gzip" {
		t.Errorf("Accept-Encoding = %q, want it preserved", got)
	}
	if got := req.Header.Get("Cookie"); got != "" {
		t.Errorf("Cookie = %q, want it dropped", got)
	}
}

func TestAdminAPIProxyUsesVerifiedEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotEmail []string
	var gotAuth, gotCookie, gotPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEmail = r.Header.Values(adminEmailHeader)
		gotAuth = r.Header.Get("Authorization")
		gotCookie = r.Header.Get("Cookie")
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	h := NewProxyHandler(&internal.Config{AdminAPIUrl: upstream.URL}, zap.NewNop())
	h.authCache.Store("session-token", authCacheEntry{
		email:   "ada@nuon.co",
		expires: time.Now().Add(time.Minute),
	})

	e := gin.New()
	if err := h.RegisterRoutes(e); err != nil {
		t.Fatalf("RegisterRoutes: %v", err)
	}
	dashboard := httptest.NewServer(e)
	t.Cleanup(dashboard.Close)

	req, err := http.NewRequest(http.MethodPost, dashboard.URL+"/admin/v1/orgs/org_123/admin-forget", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: authCookie, Value: "session-token"})
	req.Header.Set(adminEmailHeader, "ada@nuon.co")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	if gotPath != "/v1/orgs/org_123/admin-forget" {
		t.Errorf("path = %q, want /v1/orgs/org_123/admin-forget", gotPath)
	}
	if len(gotEmail) != 1 || gotEmail[0] != "ada@nuon.co" {
		t.Errorf("X-Nuon-Admin-Email = %q, want [ada@nuon.co]", gotEmail)
	}
	if gotAuth != "Bearer session-token" {
		t.Errorf("Authorization = %q, want bearer session token", gotAuth)
	}
	if gotCookie != "" {
		t.Errorf("Cookie = %q, want it dropped", gotCookie)
	}

	omitted, err := http.NewRequest(http.MethodGet, dashboard.URL+"/admin/v1/orgs/admin-features", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	omitted.AddCookie(&http.Cookie{Name: authCookie, Value: "session-token"})
	omittedResp, err := http.DefaultClient.Do(omitted)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	omittedResp.Body.Close()
	if omittedResp.StatusCode != http.StatusNoContent {
		t.Fatalf("omitted header status = %d, want %d", omittedResp.StatusCode, http.StatusNoContent)
	}
	if len(gotEmail) != 1 || gotEmail[0] != "ada@nuon.co" {
		t.Errorf("omitted header upstream email = %q, want [ada@nuon.co]", gotEmail)
	}
}

func TestAdminAPIProxyDropsClientEmailWithoutVerifiedSession(t *testing.T) {
	var gotEmail []string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEmail = r.Header.Values(adminEmailHeader)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	h := NewProxyHandler(&internal.Config{AdminAPIUrl: upstream.URL}, zap.NewNop())
	proxy := h.newAdminAPIProxy(h.cfg.AdminAPIUrl)

	req := httptest.NewRequest(http.MethodPost, "/admin/v1/orgs/org_123/admin-forget", nil)
	req.Header.Set(adminEmailHeader, "other@nuon.co")
	req.AddCookie(&http.Cookie{Name: authCookie, Value: "session-token"})

	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if len(gotEmail) != 0 {
		t.Errorf("X-Nuon-Admin-Email = %q, want it dropped", gotEmail)
	}
}

func TestAdminAPIProxyRejectsMismatchedEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var upstreamCalls int
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls++
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	core, logs := observer.New(zap.WarnLevel)
	h := NewProxyHandler(&internal.Config{AdminAPIUrl: upstream.URL}, zap.New(core))
	h.authCache.Store("session-token", authCacheEntry{
		email:   "ada@nuon.co",
		expires: time.Now().Add(time.Minute),
	})

	e := gin.New()
	if err := h.RegisterRoutes(e); err != nil {
		t.Fatalf("RegisterRoutes: %v", err)
	}
	dashboard := httptest.NewServer(e)
	t.Cleanup(dashboard.Close)

	req, err := http.NewRequest(http.MethodPost, dashboard.URL+"/admin/v1/orgs/org_123/admin-forget", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: authCookie, Value: "session-token"})
	req.Header.Set(adminEmailHeader, "other@nuon.co")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
	if upstreamCalls != 0 {
		t.Fatalf("upstream calls = %d, want 0", upstreamCalls)
	}
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("logs = %d, want 1", len(entries))
	}
	if entries[0].Level != zap.WarnLevel || entries[0].Message != "admin proxy client email does not match verified session" {
		t.Errorf("log = %s %q, want a mismatch warning", entries[0].Level, entries[0].Message)
	}
	fields := entries[0].ContextMap()
	if fields["client_email"] != "other@nuon.co" || fields["verified_email"] != "ada@nuon.co" {
		t.Errorf("log fields = %v", fields)
	}
	if fields["path"] != "/admin/v1/orgs/org_123/admin-forget" {
		t.Errorf("path = %v", fields["path"])
	}
}

func TestRequireNuonEmailStampsVerifiedEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewProxyHandler(&internal.Config{}, zap.NewNop())
	h.authCache.Store("session-token", authCacheEntry{
		email:   "ada@nuon.co",
		expires: time.Now().Add(time.Minute),
	})

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "/admin/v1/orgs/admin-features", nil)
	req.AddCookie(&http.Cookie{Name: authCookie, Value: "session-token"})
	req.Header.Set(adminEmailHeader, "other@nuon.co")
	c.Request = req

	h.requireNuonEmail()(c)

	if c.IsAborted() {
		t.Fatal("expected @nuon.co session to proceed")
	}
	if got := verifiedEmailFrom(c.Request); got != "ada@nuon.co" {
		t.Errorf("verified email = %q, want ada@nuon.co", got)
	}
	if got := c.Request.Header.Get(adminEmailHeader); got != "other@nuon.co" {
		t.Errorf("middleware changed the client header to %q", got)
	}
}

func TestRequireNuonEmailRejectsOtherDomains(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewProxyHandler(&internal.Config{}, zap.NewNop())
	h.authCache.Store("session-token", authCacheEntry{
		email:   "ada@example.com",
		expires: time.Now().Add(time.Minute),
	})

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/admin/v1/orgs/admin-features", nil)
	req.AddCookie(&http.Cookie{Name: authCookie, Value: "session-token"})
	c.Request = req

	h.requireNuonEmail()(c)

	if !c.IsAborted() {
		t.Fatal("expected non-nuon session to be rejected")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if got := verifiedEmailFrom(c.Request); got != "" {
		t.Errorf("verified email = %q, want it unset", got)
	}
}

func TestRewriteSwaggerSpecV2(t *testing.T) {
	tests := []struct {
		name         string
		apiPrefix    string
		wantBasePath string
	}{
		{name: "admin routes through /admin", apiPrefix: "/admin", wantBasePath: "/admin"},
		{name: "public routes same-origin", apiPrefix: "", wantBasePath: "/"},
	}

	spec := []byte(`{
		"swagger": "2.0",
		"host": "ctl.nuon.us-west-2.prod.internal.nuon.co",
		"schemes": ["http"],
		"basePath": "/",
		"paths": {"/v1/installs": {}}
	}`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := rewriteSwaggerSpec(spec, tt.apiPrefix)
			if err != nil {
				t.Fatalf("rewriteSwaggerSpec: %v", err)
			}

			var doc map[string]any
			if err := json.Unmarshal(out, &doc); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if _, ok := doc["host"]; ok {
				t.Errorf("host should be dropped, got %v", doc["host"])
			}
			if _, ok := doc["schemes"]; ok {
				t.Errorf("schemes should be dropped, got %v", doc["schemes"])
			}
			if got := doc["basePath"]; got != tt.wantBasePath {
				t.Errorf("basePath = %v, want %v", got, tt.wantBasePath)
			}
		})
	}
}

func TestRewriteSwaggerSpecV3(t *testing.T) {
	spec := []byte(`{
		"openapi": "3.0.0",
		"servers": [{"url": "http://ctl.nuon.us-west-2.prod.internal.nuon.co"}],
		"paths": {"/v1/installs": {}}
	}`)

	out, err := rewriteSwaggerSpec(spec, "/admin")
	if err != nil {
		t.Fatalf("rewriteSwaggerSpec: %v", err)
	}

	var doc struct {
		Servers []struct {
			URL string `json:"url"`
		} `json:"servers"`
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(doc.Servers) != 1 || doc.Servers[0].URL != "/admin" {
		t.Errorf("servers = %+v, want single url /admin", doc.Servers)
	}
}
