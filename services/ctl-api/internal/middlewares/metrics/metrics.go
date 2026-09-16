package metrics

import (
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
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
