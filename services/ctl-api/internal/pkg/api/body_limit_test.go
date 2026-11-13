package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/size"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// Mounts the same "error" and "size" middlewares the deployed configs list, so a configured
// middleware re-applying max_request_size would show up as a 413 on the runner server.
func TestServerBodyLimitOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &internal.Config{MaxRequestSize: 16, RunnerMaxRequestSize: 64}
	l := zap.NewNop()
	mws := []middlewares.Middleware{
		stderr.New(l, cfg),
		size.New(size.Params{Cfg: cfg, L: l}),
	}

	for name, tc := range map[string]struct {
		override int64
		size     int
		want     int
	}{
		"public server uses max_request_size":            {size: 32, want: http.StatusRequestEntityTooLarge},
		"runner server uses runner_max_request_size":     {override: cfg.RunnerMaxRequestSize, size: 32, want: http.StatusOK},
		"runner server still rejects past its own limit": {override: cfg.RunnerMaxRequestSize, size: 65, want: http.StatusRequestEntityTooLarge},
	} {
		t.Run(name, func(t *testing.T) {
			a := &API{
				cfg:                   cfg,
				l:                     l,
				name:                  "test",
				maxRequestSize:        tc.override,
				middlewares:           mws,
				configuredMiddlewares: []string{"error", "size"},
			}
			require.NoError(t, a.init(nil))
			require.NoError(t, a.registerMiddlewares())
			a.handler.POST("/", func(c *gin.Context) {
				if _, err := c.GetRawData(); err != nil {
					c.Error(err)
					return
				}
				c.Status(http.StatusOK)
			})

			rec := httptest.NewRecorder()
			a.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(make([]byte, tc.size))))
			assert.Equal(t, tc.want, rec.Code)
		})
	}
}
