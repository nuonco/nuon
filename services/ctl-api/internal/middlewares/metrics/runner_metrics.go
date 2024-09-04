package metrics

import (
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
)

type runnerMiddleware struct {
	baseMiddleware
}

func (m *runnerMiddleware) Name() string {
	return "runner_metrics"
}

func NewRunner(writer metrics.Writer, l *zap.Logger) *runnerMiddleware {
	return &runnerMiddleware{
		baseMiddleware: baseMiddleware{
			l:       l,
			writer:  writer,
			context: "runner_api",
		},
	}
}
