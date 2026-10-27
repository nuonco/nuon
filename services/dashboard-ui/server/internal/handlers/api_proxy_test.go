package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

func TestAPIProxyDropsUpstreamCORSHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4001")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Expose-Headers", "X-Nuon-Api-Version")
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(upstream.Close)

	h := NewAPIProxyHandler(&internal.Config{APIUrl: upstream.URL}, zap.NewNop())
	e := gin.New()
	e.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:4001")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Next()
	})
	if err := h.RegisterRoutes(e); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	resp, err := http.Post(srv.URL+"/v1/orgs", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Values("Access-Control-Allow-Origin"); len(got) != 1 || got[0] != "http://localhost:4001" {
		t.Fatalf("allow-origin = %q, want one http://localhost:4001", got)
	}
	if got := resp.Header.Values("Access-Control-Allow-Credentials"); len(got) != 1 {
		t.Fatalf("allow-credentials = %q, want one value", got)
	}
	if got := resp.Header.Values("Access-Control-Expose-Headers"); len(got) != 0 {
		t.Fatalf("expose-headers = %q, want upstream value dropped", got)
	}
}
