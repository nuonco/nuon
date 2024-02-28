package metrics

import (
	"github.com/powertoolsdev/mono/pkg/metrics"
	"go.uber.org/zap"
)

type middleware struct {
	baseMiddleware
}

func (m *middleware) Name() string {
	return "public_metrics"
}

func New(writer metrics.Writer, l *zap.Logger) *middleware {
	return &middleware{
		baseMiddleware: baseMiddleware{
			l:       l,
			writer:  writer,
			context: "public_api",
		},
	}
}
