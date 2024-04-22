package metrics

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"go.uber.org/zap"
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

		c.Set(ContextKey, &MetricContext{
			Endpoint: endpoint,
			Method:   c.Request.Method,
			Context:  m.context,
		})

		c.Next()

		if len(c.Errors) > 0 {
			status = "err"
		}

		statusCodeClass := fmt.Sprintf("%dxx", c.Writer.Status()/100)
		tags := []string{
			"status:" + status,
			"status_code_class:" + statusCodeClass,
			"endpoint:" + endpoint,
			"method:" + c.Request.Method,
			"context:" + m.context,
		}

		m.writer.Incr("api.request.status", tags)
		m.writer.Timing("api.request.latency", time.Since(startTS), tags)
	}
}
