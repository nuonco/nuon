package headers

import (
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"go.uber.org/zap"
)

type middleware struct {
	l *zap.Logger
}

func (m *middleware) Name() string {
	return "headers"
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	}
}

func New(writer metrics.Writer, l *zap.Logger) *middleware {
	return &middleware{
		l: l,
	}
}
