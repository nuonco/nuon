package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"go.uber.org/zap"
)

func (m *middleware) setAccount(ctx *gin.Context) error {
	email := ctx.Request.Header.Get(emailHeaderKey)
	if email == "" {
		m.l.Debug("no admin email header found")
		return nil
	}

	m.l.Info("admin email header found", zap.String("email", email))
	acct, err := m.acctClient.FindAccount(ctx, email)
	if err != nil {
		return stderr.ErrAuthorization{
			Err:         err,
			Description: "please make sure to use a valid email",
		}
	}

	cctx.SetAccountGinContext(ctx, acct)
	return nil
}
