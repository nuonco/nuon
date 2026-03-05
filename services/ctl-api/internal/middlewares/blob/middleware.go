package blob

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type middleware struct {
	l   *zap.Logger
	svc blobstore.Service
}

func (m middleware) Name() string {
	return "blob"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		blobstore.WithBlobServiceGin(ctx, m.svc)
		ctx.Next()
	}
}

func New(l *zap.Logger, svc blobstore.Service) *middleware {
	return &middleware{
		l:   l,
		svc: svc,
	}
}
