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

func Route(kind permissions.ResourceKind, verb permissions.Permission, paramName string) gin.HandlerFunc {
	description := noun(kind) + " not found"

	return func(ctx *gin.Context) {
		// why: The response carries only the generic description: the stderr handler
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

		obj := permissions.Object(orgID, kind, id)
		if err := acct.AllPermissions.CanPerform(obj, verb); err != nil {
			notFound(fmt.Errorf("account %s cannot %s %s %s: %w", acct.ID, verb, kind, id, err))
			return
		}

		ctx.Next()
	}
}
