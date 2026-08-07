package tracer

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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
	tracer := otel.Tracer("github.com/nuonco/nuon/services/ctl-api/internal/middlewares/tracer")

	return func(ctx *gin.Context) {
		traceID := ctx.Request.Header.Get(traceIDHeaderKey)
		if traceID == "" {
			u7 := uuid.Must(uuid.NewV7())
			traceID = u7.String()
		}

		cctx.SetTraceIDGinContext(ctx, traceID)

		// set trace id header for all responses
		ctx.Writer.Header().Set(traceIDHeaderKey, traceID)

		route := ctx.FullPath()
		if route == "" {
			route = "unmatched"
		}

		// Downstream instrumentation (e.g. the gorm tracing plugin) parents to this span
		// through the request context; gin's own Value does not expose it because
		// engine.ContextWithFallback is off.
		spanCtx, span := tracer.Start(ctx.Request.Context(),
			ctx.Request.Method+" "+route,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", ctx.Request.Method),
				attribute.String("http.route", route),
				attribute.String("nuon.trace_id", traceID),
			),
		)
		defer span.End()
		ctx.Request = ctx.Request.WithContext(spanCtx)

		ctx.Next()

		status := ctx.Writer.Status()
		span.SetAttributes(attribute.Int("http.status_code", status))

		if metricCtx, err := cctx.MetricsContextFromGinContext(ctx); err == nil {
			span.SetAttributes(
				attribute.String("nuon.org_id", metricCtx.OrgID),
				attribute.String("nuon.namespace", metricCtx.Namespace),
				attribute.String("nuon.context", metricCtx.Context),
			)
		}

		if status >= 500 {
			span.SetStatus(codes.Error, http.StatusText(status))
		}
		if len(ctx.Errors) > 0 {
			span.RecordError(ctx.Errors[0].Err)
		}
	}
}

func New(params Params) *middleware {
	return &middleware{
		l:  params.L,
		db: params.DB,
	}
}
