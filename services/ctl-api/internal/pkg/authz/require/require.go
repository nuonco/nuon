// Package require turns a route's declared resource and verb into the gin
// middleware that authorizes it.
package require

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

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

// Route returns middleware enforcing a resource and verb. paramName is the gin path
// param holding the resource ID. Runner surface only: assumes account and org are
// already in context.
//
// Not-found on any failure — a 403 would confirm the resource exists in another org.
func Route(kind permissions.ResourceKind, verb permissions.Permission, paramName string) gin.HandlerFunc {
	description := noun(kind) + " not found"

	return func(ctx *gin.Context) {
		// The response carries only the generic description: the stderr handler
		// serializes Err into the body, and a denial reason there ("is not
		// authorized") would tell the caller the resource exists, defeating the
		// not-found posture. The reason goes to the request log instead — 404s
		// are otherwise never logged server-side.
		notFound := func(err error) {
			cctx.GetLogger(ctx, zap.NewNop()).Info("route authorization denied",
				zap.String("route", ctx.FullPath()),
				zap.Error(err),
			)
			ctx.Error(stderr.ErrNotFound{Err: errors.New(description), Description: description})
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
