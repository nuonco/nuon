package metrics

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/metrics"
)

func New(writer metrics.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTS := time.Now()
		status := "ok"

		c.Next()
		if len(c.Errors) > 0 {
			status = "err"
		}
		tags := []string{
			"status:" + status,
			"endpoint:" + fmt.Sprintf("%s-%s", c.Request.Method, c.Request.URL.Path),
		}

		writer.Incr("api.request.status", 1, tags)
		writer.Timing("api.request.latency", time.Since(startTS), tags)
	}
}
