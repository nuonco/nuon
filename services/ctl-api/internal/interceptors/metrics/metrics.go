package metrics

import (
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
)

func New(writer metrics.Writer, l *zap.Logger) *workerInterceptor {
	return &workerInterceptor{
		mw: writer,
		l:  l,
	}
}
