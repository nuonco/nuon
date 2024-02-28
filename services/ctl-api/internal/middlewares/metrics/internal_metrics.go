package metrics

import (
	"github.com/powertoolsdev/mono/pkg/metrics"
	"go.uber.org/zap"
)

type internalMiddleware struct {
	baseMiddleware
}

func (m *internalMiddleware) Name() string {
	return "internal_metrics"
}

func NewInternal(writer metrics.Writer, l *zap.Logger) *internalMiddleware {
	return &internalMiddleware{
		baseMiddleware: baseMiddleware{
			l:       l,
			writer:  writer,
			context: "internal_api",
		},
	}
}
