package runnerscope

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type Middleware struct{}

func New() *Middleware {
	return &Middleware{}
}

func (m *Middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}

		acct, err := cctx.AccountFromGinContext(ctx)
		if err != nil {
			_, cookieErr := ctx.Cookie("X-Nuon-Auth")
			if ctx.GetHeader("Authorization") == "" && ctx.Query("token") == "" && cookieErr != nil {
				ctx.Next()
				return
			}
			ctx.Error(stderr.ErrAuthentication{
				Err:         err,
				Description: "unable to authenticate account on public API",
			})
			ctx.Abort()
			return
		}
		if acct.AccountType != app.AccountTypeService {
			ctx.Next()
			return
		}
		for _, role := range acct.Roles {
			if role.RoleType != app.RoleTypeRunner {
				continue
			}
			ctx.Error(stderr.ErrAuthorization{
				Err:         errors.New("runner account used on public API"),
				Description: "runner credentials are only valid on the runner API",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
