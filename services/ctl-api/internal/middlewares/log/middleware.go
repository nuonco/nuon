package log

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type middleware struct {
	l   *zap.Logger
	cfg *internal.Config
}

func (m middleware) Name() string {
	return "log"
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startAt := time.Now()

		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: ctx.Writer}
		ctx.Writer = w
		childLogger := m.setCTXLogger(ctx)

		logRequests := false
		if m.cfg.ServiceDeployment == "admin" {
			logRequests = true
		} else if m.cfg.ForceDebugMode {
			logRequests = true
		} else {
			org, _ := cctx.OrgFromContext(ctx)
			if org != nil {
				logRequests = org.DebugMode
			}
		}

		if logRequests {
			reqZapFields := m.requestToZapFields(ctx, startAt)
			requestMessage := fmt.Sprintf("req: %s %s", ctx.Request.Method, ctx.FullPath())
			childLogger.Info(requestMessage, reqZapFields...)
		}

		ctx.Next()

		status := ctx.Writer.Status()
		duration := float64(time.Since(startAt).Milliseconds())
		respZapFields := m.responseToZapFields(ctx, w, startAt)
		responseMessage := fmt.Sprintf("res: %s %s %d (%.2fms)", ctx.Request.Method, ctx.FullPath(), status, duration)

		if status >= 500 {
			childLogger.Error(responseMessage, respZapFields...)
		} else {
			if logRequests {
				childLogger.Info(responseMessage, respZapFields...)
			}
		}
	}
}

// helper func to add headers to zap fields
func headerToZapField(header string) zap.Field {
	parts := strings.SplitN(header, ": ", 2)
	if len(parts) != 2 {
		return zap.String("header", header)
	}
	key := parts[0]
	value := parts[1]

	// Mask sensitive headers
	sensitiveHeaders := map[string]struct{}{
		"Authorization": {},
		"Cookie":        {},
		"Set-Cookie":    {},
	}

	if _, ok := sensitiveHeaders[key]; ok {
		value = "*"
	}

	return zap.String(fmt.Sprintf("header_%s", strings.ToLower(strings.ReplaceAll(key, "-", "_"))), value)
}

func (m *middleware) setCTXLogger(ctx *gin.Context) *zap.Logger {
	fields := []zap.Field{}
	// Add organization context
	org, err := cctx.OrgFromContext(ctx)
	if err == nil {
		fields = append(fields,
			zap.String("org_id", org.ID),
			zap.String("org_name", org.Name),
			zap.Bool("org_debug_mode", org.DebugMode),
		)
	}

	// Add account context
	acct, err := cctx.AccountFromContext(ctx)
	if err == nil {
		fields = append(fields,
			zap.String("account_id", acct.ID),
			zap.String("account_email", acct.Email),
		)
	} else {
		m.l.Debug("no account object on request")
	}

	// Add request body if configured
	if ctx.Request.Body != nil && m.cfg.LogRequestBody {
		body, err := ctx.GetRawData()
		if err == nil || len(body) > 0 {
			fields = append(fields, zap.String("request_body", string(body)))
		}
	}

	cctx.SetAPILoggerFields(ctx, fields)

	ctxLogger := cctx.GetLogger(ctx, m.l)

	return ctxLogger
}

// helper func to generate zap fields for request
func (m *middleware) requestToZapFields(ctx *gin.Context, startAt time.Time) []zap.Field {
	fields := []zap.Field{
		zap.String("method", ctx.Request.Method),
		zap.String("host", ctx.Request.Host),
		zap.String("url", ctx.Request.URL.String()),
		zap.String("path", ctx.FullPath()),
		zap.String("query", ctx.Request.URL.RawQuery),
		zap.String("ip", ctx.ClientIP()),
		zap.String("started_at", startAt.Format(time.RFC3339)),
	}

	for key, values := range ctx.Request.Header {
		for _, value := range values {
			fields = append(fields, headerToZapField(fmt.Sprintf("%s: %s", key, value)))
		}
	}

	// Add request body if configured
	if ctx.Request.Body != nil && m.cfg.LogRequestBody {
		body, err := ctx.GetRawData()
		if err == nil || len(body) > 0 {
			fields = append(fields, zap.String("request_body", string(body)))
		}
	}

	return fields
}

func (m *middleware) responseToZapFields(ctx *gin.Context, bw *responseBodyWriter, startAt time.Time) []zap.Field {
	endAt := time.Now()
	fields := []zap.Field{
		zap.String("method", ctx.Request.Method),
		zap.Int("status", ctx.Writer.Status()),
		zap.String("host", ctx.Request.Host),
		zap.String("url", ctx.Request.URL.String()),
		zap.String("path", ctx.FullPath()),
		zap.String("query", ctx.Request.URL.RawQuery),
		zap.String("ip", ctx.ClientIP()),
		zap.String("started_at", startAt.Format(time.RFC3339)),
		zap.String("ended_at", endAt.Format(time.RFC3339)),
		zap.Float64("duration_ms", float64(endAt.Sub(startAt).Milliseconds())),
	}

	for key, values := range ctx.Writer.Header() {
		for _, value := range values {
			fields = append(fields, headerToZapField(fmt.Sprintf("%s: %s", key, value)))
		}
	}

	if ctx.Writer.Status() >= 500 {
		fields = append(fields, zap.String("response_body", bw.body.String()))
	}

	return fields
}

type Params struct {
	fx.In
	L   *zap.Logger
	Cfg *internal.Config
}

func New(params Params) *middleware {
	return &middleware{
		l:   params.L,
		cfg: params.Cfg,
	}
}
