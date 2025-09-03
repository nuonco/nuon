package panicker

import (
	"errors"
	"net/http/httputil"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type Params struct {
	fx.In
	L      *zap.Logger
	Writer metrics.Writer
}

type middleware struct {
	l      *zap.Logger
	writer metrics.Writer
}

func (m middleware) Name() string {
	return "panicker"
}

func (m middleware) Handler() gin.HandlerFunc {
	return gin.CustomRecovery(func(ctx *gin.Context, err any) {
		attrs := []zap.Field{
			zap.String("path", ctx.FullPath()),
			zap.String("method", ctx.Request.Method),
		}

		httpRequest, _ := httputil.DumpRequest(ctx.Request, false)
		headers := strings.Split(string(httpRequest), "\r\n")
		for idx, header := range headers {
			current := strings.Split(header, ":")
			if current[0] == "Authorization" {
				headers[idx] = current[0] + ": *"
			}

			attrs = append(attrs, zap.String("header", headers[idx]))
		}

		m.l.Error("api request panic", attrs...)
		metricsCtx, err := cctx.MetricsContextFromGinContext(ctx)
		if err != nil {
			m.l.Error("no metrics context found")
			return
		}
		metricsCtx.IsPanic = true

		ctx.Error(stderr.ErrSystem{
			Err:         errors.New("internal error occurred"),
			Description: "Please try the request again",
		})
	})
}

func New(params Params) *middleware {
	return &middleware{
		l:      params.L,
		writer: params.Writer,
	}
}
