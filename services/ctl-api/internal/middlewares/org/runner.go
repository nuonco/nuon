package org

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

var _ middlewares.Middleware = (*runnerMiddleware)(nil)

type runnerMiddleware struct {
	l *zap.Logger
}

func (m *runnerMiddleware) Name() string {
	return "runner_org"
}

func (m *runnerMiddleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsGlobal(ctx) || cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}

		acct, err := cctx.AccountFromContext(ctx)
		if err != nil {
			ctx.Error(stderr.ErrSystem{
				Err:         fmt.Errorf("no runner account in request"),
				Description: "invalid runner middleware configuration",
			})
			ctx.Abort()
			return
		}

		if len(acct.OrgIDs) > 1 {
			m.l.Warn("runner associated with more than one org",
				zap.String("org_ids", strings.Join(acct.OrgIDs, ",")),
			)
			ctx.Error(stderr.ErrAuthorization{
				Err:         fmt.Errorf("runner account associated with more than one org"),
				Description: fmt.Sprintf("please retry request correct runner account"),
			})
			ctx.Abort()
			return
		}
		if len(acct.OrgIDs) < 1 {
			ctx.Error(stderr.ErrAuthorization{
				Err:         fmt.Errorf("runner account not associated any org"),
				Description: fmt.Sprintf("please retry request correct runner account"),
			})
			ctx.Abort()
			return
		}

		cctx.SetOrgIDGinContext(ctx, acct.OrgIDs[0])
		cctx.SetOrgGinContext(ctx, acct.Orgs[0])
	}
}

func NewRunner(params Params) *runnerMiddleware {
	return &runnerMiddleware{}
}
