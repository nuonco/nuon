package require

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func UnboundedCredentialRoute(router gin.IRouter) *gin.RouterGroup {
	return router.Group("", func(ctx *gin.Context) {
		if authz.HasAuthority(ctx) {
			authority, err := authz.AuthorityFromContext(ctx)
			if err != nil {
				operationFailure(ctx, err)
				return
			}
			if authority.HasCeiling() {
				ctx.Error(stderr.ErrAuthorization{Err: errors.New("bounded credentials cannot delegate access through this endpoint"), Description: "bounded credentials cannot delegate access through this endpoint"})
				ctx.Abort()
				return
			}
		}
		ctx.Next()
	})
}

func Organization(operation permissions.PermissionID) Rule {
	return Point(operation, func(ctx *gin.Context) (permissions.Target, error) {
		org, err := cctx.OrgFromContext(ctx)
		if err != nil {
			return permissions.Target{}, err
		}
		return permissions.Target{ResolvedTarget: permissions.ResolvedTarget{
			Scope: permissions.Scope{Kind: permissions.KindOrg, OrgID: org.ID, ID: org.ID},
		}}, nil
	})
}

func Point(operation permissions.PermissionID,
	resolve func(*gin.Context) (permissions.Target, error),
) Rule {
	return All(func(ctx *gin.Context) ([]authz.Requirement, error) {
		target, err := resolve(ctx)
		if err != nil {
			return nil, err
		}
		return []authz.Requirement{{Operation: operation, Target: target}}, nil
	})
}

type Rule struct {
	check          func(*gin.Context, *authz.AccessContext) (*authz.AccessContext, error)
	notFoundDenied bool
}

func All(resolve func(*gin.Context) ([]authz.Requirement, error)) Rule {
	return Rule{notFoundDenied: true, check: func(ctx *gin.Context, access *authz.AccessContext) (*authz.AccessContext, error) {
		requirements, err := resolve(ctx)
		if err != nil {
			return nil, err
		}
		return access.Require(requirements...)
	}}
}

func (rule Rule) handler(api permissions.API) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authority, err := authz.RequestAuthority(ctx, api)
		if err != nil {
			operationFailure(ctx, err)
			return
		}
		if authority == nil {
			ctx.Next()
			return
		}
		if rule.check == nil {
			operationFailure(ctx, errors.New("missing authorization rule"))
			return
		}
		access := authority.Access()
		if _, exists := ctx.Get(accessKey); exists {
			access, err = Access(ctx)
			if err != nil {
				operationFailure(ctx, err)
				return
			}
		}
		access, err = rule.check(ctx, access)
		if errors.Is(err, authz.ErrDenied) || rule.notFoundDenied && errors.Is(err, gorm.ErrRecordNotFound) {
			operationDenied(ctx, err)
			return
		}
		var invalidRequest stderr.ErrInvalidRequest
		if rule.notFoundDenied && errors.As(err, &invalidRequest) {
			ctx.Error(invalidRequest)
			ctx.Abort()
			return
		}
		if err != nil {
			operationFailure(ctx, err)
			return
		}
		ctx.Set(accessKey, access)
		ctx.Next()
	}
}

const accessKey = "nuon.authz.access"

func Access(ctx *gin.Context) (*authz.AccessContext, error) {
	api, _ := OperationAPI(ctx)
	authority, err := authz.RequestAuthority(ctx, api)
	if err != nil {
		return nil, err
	}
	value, _ := ctx.Get(accessKey)
	access, ok := value.(*authz.AccessContext)
	if !ok || !access.Matches(authority) {
		return nil, fmt.Errorf("missing or mismatched authorized request context")
	}
	return access, nil
}

func Collection(operation permissions.PermissionID,
	resources func(*gin.Context) (*authz.Resources, error),
) Rule {
	return Rule{check: func(ctx *gin.Context, access *authz.AccessContext) (*authz.AccessContext, error) {
		if _, err := access.Collection(); err == nil {
			return nil, errors.New("a route cannot apply multiple collection authorization rules")
		}
		set, err := resources(ctx)
		if err != nil {
			return nil, err
		}
		return access.Select(ctx.Request.Context(), operation, set)
	}}
}

func CollectionIDs(ctx *gin.Context) (*gorm.DB, error) {
	selection, err := collectionSelection(ctx)
	if err != nil {
		return nil, err
	}
	return selection.IDs(), nil
}

func collectionSelection(ctx *gin.Context) (*authz.Selection, error) {
	access, err := Access(ctx)
	if err != nil {
		return nil, err
	}
	return access.Collection()
}

func CollectionScope(ctx *gin.Context, columns ...clause.Column) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		api, _ := OperationAPI(ctx)
		authority, err := authz.RequestAuthority(ctx, api)
		if err != nil {
			db.AddError(err)
			return db
		}
		if authority == nil {
			return db
		}
		selection, err := collectionSelection(ctx)
		if err != nil {
			db.AddError(err)
			return db
		}
		return selection.Scope(columns...)(db)
	}
}

const operationAPIKey = "nuon.authz.operation_api"

func OperationAPI(ctx *gin.Context) (permissions.API, bool) {
	value, _ := ctx.Get(operationAPIKey)
	api, ok := value.(permissions.API)
	return api, ok
}

func operationGroup(router gin.IRouter, api permissions.API, guard gin.HandlerFunc) *gin.RouterGroup {
	route := group(router, guard)
	route.Handlers = append(gin.HandlersChain{func(ctx *gin.Context) {
		ctx.Set(operationAPIKey, api)
	}}, route.Handlers...)
	return route
}

func operationDenied(ctx *gin.Context, err error) {
	cctx.GetLogger(ctx, zap.NewNop()).Info("operation authorization denied", zap.String("route", ctx.FullPath()), zap.Error(err))
	ctx.Error(stderr.ErrNotFound{Err: errors.New("resource not found"), Description: "resource not found"})
	ctx.Abort()
}

func operationFailure(ctx *gin.Context, err error) {
	cctx.GetLogger(ctx, zap.NewNop()).Error("operation authorization failed", zap.String("route", ctx.FullPath()), zap.Error(err))
	ctx.Error(stderr.ErrSystem{Err: errors.New("unable to authorize request"), Description: "unable to authorize request"})
	ctx.Abort()
}
