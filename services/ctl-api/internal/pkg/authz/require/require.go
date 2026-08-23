// Package require turns a route's declared resource and verb into the gin
// middleware that authorizes it.
package require

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// nouns names the resource in the not-found error, so the message matches what
// the handlers say. A stack route is addressed by its install.
var nouns = map[permissions.ResourceKind]string{
	permissions.KindApp:     "app",
	permissions.KindInstall: "install",
	permissions.KindStack:   "install",
}

func noun(kind permissions.ResourceKind) string {
	if n, ok := nouns[kind]; ok {
		return n
	}

	return string(kind)
}

// Route declares the resource and verb a route operates on and returns the
// middleware enforcing it. paramName is the gin path param holding the
// resource ID. Runner-surface only for now: it assumes account and org are
// already in context from the surface's engine-wide middlewares.
//
// Not-found on any failure, matching the handlers: a 403 would confirm the
// resource exists in another org.
func Route(kind permissions.ResourceKind, verb permissions.Permission, paramName string) gin.HandlerFunc {
	description := noun(kind) + " not found"

	return func(ctx *gin.Context) {
		notFound := func(err error) {
			ctx.Error(stderr.ErrNotFound{Err: err, Description: description})
			ctx.Abort()
		}

		acct, err := cctx.AccountFromGinContext(ctx)
		if err != nil {
			notFound(fmt.Errorf("unable to resolve account from request: %w", err))
			return
		}

		orgID, err := cctx.OrgIDFromContext(ctx)
		if err != nil {
			notFound(fmt.Errorf("unable to resolve org from request: %w", err))
			return
		}

		id := ctx.Param(paramName)
		if id == "" {
			notFound(fmt.Errorf("no %s in request path", paramName))
			return
		}

		// The declared verb, not the one inferred from the request method: the
		// route says what it does, the method is only a hint.
		obj := permissions.Object(orgID, kind, id)
		if err := acct.AllPermissions.CanPerform(obj, verb); err != nil {
			notFound(fmt.Errorf("account %s cannot %s %s %s: %w", acct.ID, verb, kind, id, err))
			return
		}

		ctx.Next()
	}
}
