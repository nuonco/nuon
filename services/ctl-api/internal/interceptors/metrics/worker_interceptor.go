package metrics

import (
	"context"

	"go.temporal.io/sdk/interceptor"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
)

var (
	_ interceptor.WorkerInterceptor          = (*workerInterceptor)(nil)
	_ interceptor.ActivityInboundInterceptor = (*actInterceptor)(nil)
)

type workerInterceptor struct {
	mw metrics.Writer
	l  *zap.Logger

	interceptor.InterceptorBase
}

// this is just to intercept and return a new interceptor
func (m *workerInterceptor) InterceptActivity(
	ctx context.Context,
	next interceptor.ActivityInboundInterceptor,
) interceptor.ActivityInboundInterceptor {
	return &actInterceptor{
		interceptor.ActivityInboundInterceptorBase{
			Next: next,
		},
		m.mw,
		m.l,
	}
}
