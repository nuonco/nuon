package cctx

import (
	"context"

	"go.temporal.io/sdk/interceptor"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

var _ interceptor.ActivityInboundInterceptor = (*actInterceptor)(nil)

type actInterceptor struct {
	interceptor.ActivityInboundInterceptorBase

	blobSvc         blobstore.Service
	l               *zap.Logger
	blobReadEnabled bool
}

func (a *actInterceptor) Init(outbound interceptor.ActivityOutboundInterceptor) error {
	return a.Next.Init(outbound)
}

func (a *actInterceptor) ExecuteActivity(
	ctx context.Context,
	in *interceptor.ExecuteActivityInput,
) (interface{}, error) {
	ctx = blobstore.WithBlobService(ctx, a.blobSvc)
	ctx = blobstore.WithBlobReadEnabled(ctx, a.blobReadEnabled)

	return a.Next.ExecuteActivity(ctx, in)
}
