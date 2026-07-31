package cctx

import (
	"go.temporal.io/sdk/interceptor"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

func New(l *zap.Logger, blobSvc blobstore.Service, cfg *internal.Config) interceptor.WorkerInterceptor {
	return &workerInterceptor{
		blobSvc: blobSvc,
		l:       l,
	}
}
