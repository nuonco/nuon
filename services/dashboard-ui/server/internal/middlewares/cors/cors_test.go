package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

func TestCORSAllowsConfiguredOriginOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name          string
		appURL        string
		origin        string
		allowLoopback bool
		want          string
	}{
		{
			name:   "configured origin",
			appURL: "https://app.byoc.example.com/",
			origin: "https://app.byoc.example.com",
			want:   "https://app.byoc.example.com",
		},
		{
			name:   "sibling subdomain",
			appURL: "https://app.byoc.example.com",
			origin: "https://api.byoc.example.com",
		},
		{
			name:   "foreign origin",
			appURL: "https://app.byoc.example.com",
			origin: "https://app.other.example",
		},
		{
			name:   "missing app url",
			origin: "https://app.byoc.example.com",
		},
		{
			name:          "loopback flag allows a different port",
			appURL:        "http://localhost:4000",
			origin:        "http://localhost:4001",
			allowLoopback: true,
			want:          "http://localhost:4001",
		},
		{
			name:          "loopback flag allows 127.0.0.1",
			appURL:        "http://localhost:4000",
			origin:        "http://127.0.0.1:4010",
			allowLoopback: true,
			want:          "http://127.0.0.1:4010",
		},
		{
			name:          "loopback flag rejects a non-loopback origin",
			appURL:        "http://localhost:4000",
			origin:        "http://evil.example",
			allowLoopback: true,
		},
		{
			name:          "https localhost is not a loopback origin",
			appURL:        "http://localhost:4000",
			origin:        "https://localhost:4001",
			allowLoopback: true,
		},
		{
			name:   "empty app url rejects a loopback origin",
			origin: "http://localhost:4001",
		},
		{
			name:   "invalid app url rejects a loopback origin",
			appURL: "not a url",
			origin: "http://localhost:4001",
		},
		{
			name:   "localhost app without the flag rejects another port",
			appURL: "http://localhost:4000",
			origin: "http://localhost:4001",
		},
		{
			name:   "deployed app rejects a loopback origin",
			appURL: "https://app.byoc.example.com",
			origin: "http://localhost:4001",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := New(&internal.Config{AppUrl: tc.appURL, CORSAllowLoopbackOrigins: tc.allowLoopback}, zap.NewNop())
			e := gin.New()
			e.Use(m.Handler())
			e.GET("/v1/orgs", func(c *gin.Context) { c.Status(http.StatusOK) })

			req := httptest.NewRequest(http.MethodGet, "/v1/orgs", nil)
			req.Header.Set("Origin", tc.origin)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tc.want {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tc.want)
			}
			wantCreds := ""
			if tc.want != "" {
				wantCreds = "true"
			}
			if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != wantCreds {
				t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, wantCreds)
			}
		})
	}
}
