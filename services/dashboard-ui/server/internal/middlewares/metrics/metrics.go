package metrics

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/apiroutes"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/handlers"
)

type middleware struct {
	l          *zap.Logger
	writer     metrics.Writer
	classifier *apiroutes.Classifier
}

func New(cfg *internal.Config, l *zap.Logger, writer metrics.Writer, classifier *apiroutes.Classifier) *middleware {
	return &middleware{l: l, writer: writer, classifier: classifier}
}

func (m *middleware) Name() string {
	return "metrics"
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTS := time.Now()
		reqPath := c.Request.URL.Path

		c.Next()

		path := c.FullPath()
		if path == "" {
			return
		}

		status := "ok"
		if len(c.Errors) > 0 {
			status = "err"
		}

		if path == handlers.APIProxyRoutePattern {
			path = m.classifier.Match(reqPath)
		}

		endpoint := strings.ReplaceAll(path, "-", "_")
		statusCodeClass := fmt.Sprintf("%dxx", c.Writer.Status()/100)

		tags := []string{
			"status:" + status,
			"status_code_class:" + statusCodeClass,
			"endpoint:" + endpoint,
			"method:" + c.Request.Method,
			"context:bff",
		}

		m.writer.Timing("ui.request.latency", time.Since(startTS), tags)
		m.writer.Gauge("ui.request.size", float64(c.Request.ContentLength), tags)
	}
}
