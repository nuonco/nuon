package headers

import (
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/metrics"
)

func New(writer metrics.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	}
}
