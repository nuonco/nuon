package metrics

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"go.uber.org/zap"
)

type middleware struct {
	l      *zap.Logger
	writer metrics.Writer
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTS := time.Now()
		status := "ok"
		path := c.FullPath()
		endpoint := fmt.Sprintf("%s-%s", c.Request.Method, path)

		c.Set(ContextKey, &MetricContext{
			Endpoint: endpoint,
			Method:   c.Request.Method,
			Context:  "public_api",
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
		}

		m.writer.Incr("api.request.status", 1, tags)
		m.writer.Timing("api.request.latency", time.Since(startTS), tags)
	}
}

func (m *middleware) Name() string {
	return "metrics"
}

func New(writer metrics.Writer, l *zap.Logger) *middleware {
	return &middleware{
		writer: writer,
		l:      l,
	}
}
