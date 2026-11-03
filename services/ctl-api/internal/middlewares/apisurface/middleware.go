package apisurface

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
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
		credential, ok := cctx.VerifiedCredentialFromContext(ctx.Request.Context())
		limits := permissions.CredentialLimits{Profile: credential.Profile, API: credential.Audience, Ceiling: credential.Ceiling}
		if !ok || credential.AccountID == "" || credential.AccountID != acct.ID || credential.TokenID == "" ||
			credential.Surface != cctx.SurfacePublicAPI || credential.Audience != required ||
			required != permissions.APICtl || limits.Validate() != nil {
			ctx.Error(stderr.ErrAuthorization{
				Err:         fmt.Errorf("verified credential is not permitted on %s", required),
				Description: "token is not allowed on this api",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
