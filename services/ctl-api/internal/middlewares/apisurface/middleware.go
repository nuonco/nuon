package apisurface

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type Middleware struct{}

type Params struct {
	fx.In
}

func New(Params) *Middleware {
	return &Middleware{}
}

func (m *Middleware) Handler(required permissions.API) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Unauthenticated public allowlist only. Global (no org header) still
		// requires a matching api surface — runner tokens must not call ctl-api.
		if cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}
		acct, err := cctx.AccountFromGinContext(ctx)
		if err != nil {
			ctx.Error(stderr.ErrAuthorization{
				Err:         err,
				Description: "token is not allowed on this api",
			})
			ctx.Abort()
			return
		}
		api, enforce, err := authz.APIForRoles(acct.Roles)
		if err != nil {
			ctx.Error(stderr.ErrAuthorization{
				Err:         err,
				Description: "token is not allowed on this api",
			})
			ctx.Abort()
			return
		}
		if !enforce {
			ctx.Next()
			return
		}
		if api != required && !runnerStackRoute(required, api, ctx) {
			ctx.Error(stderr.ErrAuthorization{
				Err:         fmt.Errorf("token api %s cannot call %s", api, required),
				Description: "token is not allowed on this api",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// The stack API is not deployed yet, so stack_api_url points at the runner
// host and Terraform reaches these two routes there.
func runnerStackRoute(required, api permissions.API, ctx *gin.Context) bool {
	if required != permissions.APIRunner || api != permissions.APIStack {
		return false
	}
	switch ctx.Request.Method {
	case http.MethodGet:
		return ctx.FullPath() == "/v1/stacks/:install_id/config"
	case http.MethodPost:
		return ctx.FullPath() == "/v1/stacks/:install_id/phone-home"
	default:
		return false
	}
}
