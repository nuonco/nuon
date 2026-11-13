package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func TestServerBodyLimitOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &internal.Config{MaxRequestSize: 16, RunnerMaxRequestSize: 64}

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
			a := &API{cfg: cfg, l: zap.NewNop(), name: "test", maxRequestSize: tc.override}
			if err := a.init(nil); err != nil {
				t.Fatal(err)
			}
			a.handler.POST("/", func(c *gin.Context) { c.Status(http.StatusOK) })

			rec := httptest.NewRecorder()
			a.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(make([]byte, tc.size))))
			assert.Equal(t, tc.want, rec.Code)
		})
	}
}
