package cctx

import (
	"context"

	"go.temporal.io/sdk/interceptor"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

var (
	_ interceptor.WorkerInterceptor          = (*workerInterceptor)(nil)
	_ interceptor.ActivityInboundInterceptor = (*actInterceptor)(nil)
)

type workerInterceptor struct {
	blobSvc blobstore.Service
	l       *zap.Logger

	interceptor.InterceptorBase
}

// InterceptActivity intercepts activity execution to inject context values
func (w *workerInterceptor) InterceptActivity(
	ctx context.Context,
	next interceptor.ActivityInboundInterceptor,
) interceptor.ActivityInboundInterceptor {
	return &actInterceptor{
		interceptor.ActivityInboundInterceptorBase{
			Next: next,
		},
		w.blobSvc,
		w.l,
	}
}
