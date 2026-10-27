package require

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// NotFederatedToken rejects the request if it was made with a federated
// (OIDC-exchanged) token. action completes "federated tokens cannot <action>"
// in the error message.
func NotFederatedToken(ctx context.Context, action string) error {
	if !cctx.IsFederatedToken(ctx) {
		return nil
	}

	return stderr.ErrAuthorization{
		Err:         fmt.Errorf("federated tokens cannot %s", action),
		Description: fmt.Sprintf("Tokens issued by an OIDC trust policy cannot %s. Use a user or static API token instead.", action),
	}
}

// NotFederatedTokenRoute returns a group that rejects federated tokens before
// the handler runs, so the restriction is visible at route registration
// instead of buried in the handler body.
//
//	require.NotFederatedTokenRoute(orgs, "create organizations").POST("", s.CreateOrg)
func NotFederatedTokenRoute(router gin.IRouter, action string) *gin.RouterGroup {
	return router.Group("", func(ctx *gin.Context) {
		if err := NotFederatedToken(ctx, action); err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}
		ctx.Next()
	})
}
