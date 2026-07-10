package blob

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type middleware struct {
	l               *zap.Logger
	svc             blobstore.Service
	blobReadEnabled bool
}

func (m middleware) Name() string {
	return "blob"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqCtx := blobstore.WithBlobService(ctx.Request.Context(), m.svc)
		reqCtx = blobstore.WithBlobReadEnabled(reqCtx, m.blobReadEnabled)
		ctx.Request = ctx.Request.WithContext(reqCtx)
		ctx.Next()
	}
}

func New(l *zap.Logger, svc blobstore.Service, cfg *internal.Config) *middleware {
	return &middleware{
		l:               l,
		svc:             svc,
		blobReadEnabled: cfg.BlobReadEnabled,
	}
}
