package org

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

var _ middlewares.Middleware = (*runnerMiddleware)(nil)

type runnerMiddleware struct{}

func (m *runnerMiddleware) Name() string {
	return "runner_org"
}

func (m *runnerMiddleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsGlobal(ctx) || cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}

		acct, err := middlewares.AccountFromContext(ctx)
		if err != nil {
			ctx.Error(stderr.ErrSystem{
				Err:         fmt.Errorf("no runner account in request"),
				Description: "invalid runner middleware configuration",
			})
			ctx.Abort()
			return
		}

		if len(acct.OrgIDs) != 1 {
			ctx.Error(stderr.ErrAuthorization{
				Err:         fmt.Errorf("runner account associated with more than one org"),
				Description: fmt.Sprintf("please retry request with %s header", orgIDHeaderKey),
			})
			ctx.Abort()
			return
		}

		middlewares.SetOrgIDGinContext(ctx, acct.OrgIDs[0])
	}
}

func NewRunner(params Params) *runnerMiddleware {
	return &runnerMiddleware{}
}
