package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

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
