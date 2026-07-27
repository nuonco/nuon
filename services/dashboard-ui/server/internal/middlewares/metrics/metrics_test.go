package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/handlers"
)

type fakeWriter struct {
	timingTags []string
}

func (f *fakeWriter) Incr(string, []string)                           {}
func (f *fakeWriter) Decr(string, []string)                           {}
func (f *fakeWriter) Timing(_ string, _ time.Duration, tags []string) { f.timingTags = tags }
func (f *fakeWriter) Gauge(string, float64, []string)                 {}
func (f *fakeWriter) Count(string, int64, []string)                   {}
func (f *fakeWriter) Distribution(string, float64, []string)          {}
func (f *fakeWriter) Event(*statsd.Event)                             {}
func (f *fakeWriter) Flush()                                          {}

func endpointTag(tags []string) string {
	for _, tag := range tags {
		if len(tag) >= len("endpoint:") && tag[:len("endpoint:")] == "endpoint:" {
			return tag[len("endpoint:"):]
		}
	}
	return ""
}

func TestMetricsEndpointTag(t *testing.T) {
	cases := []struct {
		name    string
		pattern string
		reqPath string
		want    string
	}{
		{"proxied api route normalized", handlers.APIProxyRoutePattern, "/v1/apps/app98e2wpzdxwoey393edtqj45/installs", "/v1/apps/{app_id}/installs"},
		{"non-proxy route uses full path", "/api/orgs/:orgId/health", "/api/orgs/org123/health", "/api/orgs/:orgId/health"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			fw := &fakeWriter{}
			m := New(&internal.Config{}, zap.NewNop(), fw)

			e := gin.New()
			e.Use(m.Handler())
			e.Any(tc.pattern, func(c *gin.Context) { c.Status(http.StatusOK) })

			req := httptest.NewRequest(http.MethodGet, tc.reqPath, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if got := endpointTag(fw.timingTags); got != tc.want {
				t.Errorf("endpoint tag = %q, want %q (tags=%v)", got, tc.want, fw.timingTags)
			}
		})
	}
}
