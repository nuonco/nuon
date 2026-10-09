// Package require applies existing RBAC permissions to HTTP routes and
// non-HTTP callers. HTTP helpers return ordinary Gin groups whose child routes
// inherit the permission check. Install authentication and org-resolution
// middleware on the parent router before registering protected routes.
// Public and global route classification stays with the listener's middleware.
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

// Route returns a group requiring verb on the resource named by paramName in
// the request path, within the selected org. For KindOrg it checks the org itself
// and ignores paramName. The explicit verb is authoritative, regardless of the
// HTTP method. Resource denials return 404 to avoid disclosing existence; org
// denials return 403.
//
//	require.Route(api, permissions.KindStack, permissions.PermissionRead, "install_id").
//		GET("/v1/stacks/:install_id/config", handler)
func Route(router gin.IRouter, kind permissions.ResourceKind, verb permissions.Permission, paramName string) *gin.RouterGroup {
	return route(router, kind, paramName, func(*gin.Context) permissions.Permission { return verb })
}

// OrgRoute returns a group requiring permission on the selected org, inferred
// using permissions.FromRequest: GET/HEAD read, POST create, PUT/PATCH update,
// and DELETE delete. Unsupported methods fail closed at request time. Use Route
// with KindOrg and an explicit verb when the operation differs from this mapping.
//
//	org := require.OrgRoute(api)
//	org.GET("/v1/orgs/current", handler)
func OrgRoute(router gin.IRouter) *gin.RouterGroup {
	return route(router, permissions.KindOrg, "", permissions.FromRequest)
}

func route(router gin.IRouter, kind permissions.ResourceKind, paramName string, permission func(*gin.Context) permissions.Permission) *gin.RouterGroup {
	description := noun(kind) + " not found"

	guard := func(ctx *gin.Context) {
		verb := permission(ctx)
		if verb == permissions.PermissionUnknown {
			ctx.Error(stderr.ErrSystem{
				Err:         fmt.Errorf("cannot infer permission for %s; use require.Route with an explicit permission", ctx.Request.Method),
				Description: "invalid route authorization configuration",
			})
			ctx.Abort()
			return
		}

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

		if kind == permissions.KindOrg {
			if err := acct.AllPermissions.CanPerform(orgID, verb); err != nil {
				ctx.Error(PermissionDeniedError(acct, orgID, verb, ScopeFromPath(ctx.FullPath())))
				ctx.Abort()
				return
			}
			ctx.Next()
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
	return group(router, guard)
}

const declaredRouteKey = "nuon.authz.declared_route"

func group(router gin.IRouter, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	route := router.Group("", handlers...)
	// Mark the declaration before inherited middleware so it can reject missing
	// declarations. The permission guard stays after authentication/org resolution.
	route.Handlers = append(gin.HandlersChain{func(ctx *gin.Context) {
		ctx.Set(declaredRouteKey, true)
	}}, route.Handlers...)
	return route
}

// IsDeclared reports whether the matched route was registered through Route or
// OrgRoute. It is available before authentication and does not mean access was
// granted; the later permission guard must still run.
func IsDeclared(ctx *gin.Context) bool {
	return ctx.GetBool(declaredRouteKey)
}
