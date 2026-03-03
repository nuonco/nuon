package cctx

import (
	"context"

	"go.temporal.io/sdk/interceptor"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	pkgcctx "github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

var _ interceptor.ActivityInboundInterceptor = (*actInterceptor)(nil)

type actInterceptor struct {
	interceptor.ActivityInboundInterceptorBase

	blobSvc blobstore.Service
	l       *zap.Logger
}

func (a *actInterceptor) Init(outbound interceptor.ActivityOutboundInterceptor) error {
	return a.Next.Init(outbound)
}

func (a *actInterceptor) ExecuteActivity(
	ctx context.Context,
	in *interceptor.ExecuteActivityInput,
) (interface{}, error) {
	// Add blobstore service to context
	ctx = pkgcctx.SetBlobServiceContext(ctx, a.blobSvc)

	// Continue with execution
	return a.Next.ExecuteActivity(ctx, in)
}
