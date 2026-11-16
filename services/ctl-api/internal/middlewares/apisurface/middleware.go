package apisurface

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
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
		if _, operationBased := require.OperationAPI(ctx); operationBased {
			authority, err := authz.AuthorityFromContext(ctx)
			orgID, orgErr := cctx.OrgIDFromContext(ctx)
			if err != nil || orgErr != nil || authority.ValidateIdentity(acct.ID, cctx.TokenIDFromContext(ctx), orgID, required) != nil {
				ctx.Error(stderr.ErrAuthorization{
					Err:         fmt.Errorf("token is not allowed on this api"),
					Description: "token is not allowed on this api",
				})
				ctx.Abort()
				return
			}
			ctx.Next()
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
		if api != required {
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
