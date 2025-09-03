package log

import (
	"bytes"
	"errors"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// Custom response writer to capture response body and status code
type responseBodyWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *responseBodyWriter) Write(data []byte) (int, error) {
	// Write to our buffer to capture the response
	w.body.Write(data)
	// Also write to the original writer
	return w.ResponseWriter.Write(data)
}

func (w *responseBodyWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

type middleware struct {
	l   *zap.Logger
	cfg *internal.Config
}

func (m middleware) Name() string {
	return "log"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a custom response writer to capture response body
		responseBody := &bytes.Buffer{}
		customWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           responseBody,
			statusCode:     200, // default status
		}
		c.Writer = customWriter

		// Continue with ginzap middleware
		ginzapHandler := ginzap.GinzapWithConfig(m.l, &ginzap.Config{
			UTC:        true,
			TimeFormat: time.RFC3339,
			Context: func(c *gin.Context) []zapcore.Field {
				fields := make([]zapcore.Field, 0)

				// Add organization context
				org, err := cctx.OrgFromContext(c)
				if err == nil {
					fields = append(fields,
						zap.String("org_id", org.ID),
						zap.String("org_name", org.Name),
						zap.Bool("org_debug_mode", org.DebugMode),
					)
				}

				// Add account context
				acct, err := cctx.AccountFromContext(c)
				if err == nil {
					fields = append(fields,
						zap.String("account_id", acct.ID),
						zap.String("account_email", acct.Email),
					)
				} else {
					m.l.Debug("no account object on request")
				}

				// Add request body if configured
				if c.Request.Body != nil && m.cfg.LogRequestBody {
					body, err := c.GetRawData()
					if err == nil || len(body) > 0 {
						fields = append(fields, zap.String("request_body", string(body)))
					}
				}

				// Add response body for 5xx status codes
				if customWriter.statusCode >= 500 && customWriter.statusCode < 600 {
					if responseBody.Len() > 0 {
						fields = append(fields, zap.String("response_body", responseBody.String()))
					}
				}

				return fields
			},
			Skipper: func(c *gin.Context) bool {
				// some errors should never be logged, such as auth issues because they are spammy and could
				// correspond to actually bad runners or bad actors.
				if len(c.Errors) > 0 {
					var errAuth stderr.ErrAuthentication
					if errors.As(c.Errors[0], &errAuth) {
						return true
					}
				}
				//
				if m.cfg.ForceDebugMode {
					return false
				}
				org, err := cctx.OrgFromContext(c)
				if err == nil {
					if org.DebugMode {
						return false
					}
				}
				if len(c.Errors) > 0 {
					return false
				}
				return true
			},
		})

		// Execute the ginzap handler
		ginzapHandler(c)
	}
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

