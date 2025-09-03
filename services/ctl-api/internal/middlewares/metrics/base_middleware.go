package metrics

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

const (
	targetLatency time.Duration = time.Millisecond * 50
)

type baseMiddleware struct {
	l       *zap.Logger
	writer  metrics.Writer
	context string
}

func (m *baseMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTS := time.Now()
		status := "ok"
		path := c.FullPath()

		// https://docs.datadoghq.com/getting_started/tagging/ dashes are not permitted in tags
		path = strings.ReplaceAll(path, "-", "_")
		endpoint := fmt.Sprintf("%s__%s", c.Request.Method, path)

		ctxObj := &cctx.MetricContext{
			Endpoint: endpoint,
			Method:   c.Request.Method,
			Context:  m.context,
		}
		c.Set(keys.MetricsKey, ctxObj)

		c.Next()

		if len(c.Errors) > 0 {
			status = "err"
		}

		withinTargetLatency := time.Since(startTS) < targetLatency
		statusCodeClass := fmt.Sprintf("%dxx", c.Writer.Status()/100)
		tags := []string{
			"status:" + status,
			"status_code_class:" + statusCodeClass,
			"endpoint:" + endpoint,
			"method:" + c.Request.Method,
			"context:" + m.context,
			"within_target_latency:" + strconv.FormatBool(withinTargetLatency),
			"is_panic:" + strconv.FormatBool(ctxObj.IsPanic),
		}

		m.writer.Incr("api.request.status", tags)
		m.writer.Gauge("api.request.size", float64(c.Request.ContentLength), tags)
		m.writer.Gauge("gorm_operation.endpoint_count", float64(ctxObj.DBQueryCount), tags)
		m.writer.Timing("api.request.latency", time.Since(startTS), tags)
	}
}
