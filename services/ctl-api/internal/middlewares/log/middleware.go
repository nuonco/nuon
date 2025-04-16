package log

import (
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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

func (m middleware) Handler() gin.HandlerFunc {
	return ginzap.GinzapWithConfig(m.l, &ginzap.Config{
		UTC:        true,
		TimeFormat: time.RFC3339,
		Context: func(c *gin.Context) []zapcore.Field {
			fields := make([]zapcore.Field, 0)

			org, err := cctx.OrgFromContext(c)
			if err == nil {
				fields = append(fields,
					zap.String("org_id", org.ID),
					zap.String("org_name", org.Name),
					zap.Bool("org_debug_mode", org.DebugMode),
				)
			}

			acct, err := cctx.AccountFromContext(c)
			if err == nil {
				fields = append(fields,
					zap.String("account_id", acct.ID),
					zap.String("account_email", acct.Email),
				)
			} else {
				m.l.Debug("no account object on request")
			}

			if c.Request.Body != nil && m.cfg.LogRequestBody {
				body, err := c.GetRawData()
				if err == nil || len(body) > 0 {
					fields = append(fields, zap.String("request_body", string(body)))
				}
			}

			return fields
		},
		Skipper: func(c *gin.Context) bool {
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
