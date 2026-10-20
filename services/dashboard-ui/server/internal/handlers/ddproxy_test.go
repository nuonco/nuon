package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestDDProxyUpstreamURL(t *testing.T) {
	h := NewDDProxyHandler(zap.NewNop())

	tests := []struct {
		name      string
		ddforward string
		want      string
	}{
		{
			name:      "rum",
			ddforward: "/api/v2/rum?ddsource=browser&dd-api-key=pub123",
			want:      "https://browser-intake-us5-datadoghq.com/api/v2/rum?ddsource=browser&dd-api-key=pub123",
		},
		{
			name:      "logs",
			ddforward: "/api/v2/logs?ddsource=browser",
			want:      "https://browser-intake-us5-datadoghq.com/api/v2/logs?ddsource=browser",
		},
		{
			name:      "replay",
			ddforward: "/api/v2/replay?ddsource=browser",
			want:      "https://browser-intake-us5-datadoghq.com/api/v2/replay?ddsource=browser",
		},
		{name: "userinfo host swap", ddforward: "@evil.example/api/v2/rum"},
		{name: "host suffix", ddforward: ".evil.example/api/v2/rum"},
		{name: "port suffix", ddforward: ":8080/api/v2/rum"},
		{name: "protocol relative", ddforward: "//evil.example/api/v2/rum"},
		{name: "absolute url", ddforward: "https://evil.example/api/v2/rum"},
		{name: "metadata", ddforward: "http://169.254.169.254/latest/meta-data/"},
		{name: "traversal", ddforward: "/api/v2/rum/../../v1/validate"},
		{name: "other path", ddforward: "/api/v1/validate"},
		{name: "no leading slash", ddforward: "api/v2/rum"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := h.upstreamURL(tt.ddforward)
			if tt.want == "" {
				if ok {
					t.Fatalf("upstreamURL(%q) = %q, want rejected", tt.ddforward, got)
				}
				return
			}
			if !ok || got != tt.want {
				t.Fatalf("upstreamURL(%q) = %q, %v; want %q", tt.ddforward, got, ok, tt.want)
			}
		})
	}
}

func TestDDProxyHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotPath, gotMethod string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		if r.URL.Path == "/api/v2/logs" {
			http.Redirect(w, r, "http://169.254.169.254/latest/meta-data/", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer upstream.Close()

	upstreamURL, _ := url.Parse(upstream.URL)
	h := NewDDProxyHandler(zap.NewNop())
	h.intake = url.URL{Scheme: upstreamURL.Scheme, Host: upstreamURL.Host}

	e := gin.New()
	if err := h.RegisterRoutes(e); err != nil {
		t.Fatal(err)
	}

	do := func(method, ddforward string) *httptest.ResponseRecorder {
		target := "/api/ddp?ddforward=" + url.QueryEscape(ddforward)
		req := httptest.NewRequest(method, target, strings.NewReader("{}"))
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}

	if rec := do(http.MethodPost, "/api/v2/rum?ddsource=browser"); rec.Code != http.StatusAccepted {
		t.Fatalf("rum status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if gotPath != "/api/v2/rum" || gotMethod != http.MethodPost {
		t.Fatalf("upstream got %s %s, want POST /api/v2/rum", gotMethod, gotPath)
	}

	gotPath = ""
	if rec := do(http.MethodPost, "@evil.example/api/v2/rum"); rec.Code != http.StatusBadRequest {
		t.Fatalf("host injection status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if gotPath != "" {
		t.Fatalf("host injection reached upstream at %q", gotPath)
	}

	if rec := do(http.MethodGet, "/api/v2/rum"); rec.Code != http.StatusNotFound {
		t.Fatalf("GET status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	rec := do(http.MethodPost, "/api/v2/logs")
	if rec.Code != http.StatusFound {
		t.Fatalf("redirect status = %d, want %d (not followed)", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); loc != "" {
		t.Fatalf("Location = %q, want it dropped", loc)
	}
}
