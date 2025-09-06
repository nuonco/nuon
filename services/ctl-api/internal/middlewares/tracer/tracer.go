package tracer

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

const (
	traceIDHeaderKey string = "X-Nuon-Trace-ID"
)

type Params struct {
	fx.In

	L  *zap.Logger
	DB *gorm.DB `name:"psql"`
}

type middleware struct {
	l  *zap.Logger
	db *gorm.DB
}

func (m middleware) Name() string {
	return "tracer"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traceID := ctx.Request.Header.Get(traceIDHeaderKey)
		if traceID == "" {
			u7 := uuid.Must(uuid.NewV7())
			traceID = u7.String()
		}

		cctx.SetTraceIDGinContext(ctx, traceID)

		// set trace id header for all responses
		ctx.Writer.Header().Set(traceIDHeaderKey, traceID)

		ctx.Next()
	}
}

func New(params Params) *middleware {
	return &middleware{
		l:  params.L,
		db: params.DB,
	}
}
