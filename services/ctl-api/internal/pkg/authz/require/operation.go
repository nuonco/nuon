package require

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func OperationRoute(router gin.IRouter, api permissions.API, operation permissions.PermissionID,
	resolve func(*gin.Context) (permissions.Target, error),
) *gin.RouterGroup {
	return operationGroup(router, api, func(ctx *gin.Context) {
		if legacyRequest(ctx) {
			ctx.Next()
			return
		}
		authority, err := requestAuthority(ctx, api)
		if err != nil {
			operationFailure(ctx, err)
			return
		}
		target, err := resolve(ctx)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			operationDenied(ctx, err)
			return
		}
		if err != nil {
			operationFailure(ctx, err)
			return
		}
		if match := authority.Check(operation, target); !match.Matched {
			operationDenied(ctx, fmt.Errorf("operation %s denied: %s", operation, match.Reason))
			return
		}
		ctx.Next()
	})
}

const collectionIDsKey = "nuon.authz.collection_ids"

func CollectionRoute(router gin.IRouter, api permissions.API, operation permissions.PermissionID,
	resources func(*gin.Context) (*authz.Resources, error),
) *gin.RouterGroup {
	return operationGroup(router, api, func(ctx *gin.Context) {
		if legacyRequest(ctx) {
			ctx.Next()
			return
		}
		authority, err := requestAuthority(ctx, api)
		if err != nil {
			operationFailure(ctx, err)
			return
		}
		set, err := resources(ctx)
		if err != nil {
			operationFailure(ctx, err)
			return
		}
		ids, err := authority.AuthorizedIDs(ctx, operation, set)
		if err != nil {
			operationFailure(ctx, err)
			return
		}
		ctx.Set(collectionIDsKey, ids)
		ctx.Next()
	})
}

func CollectionIDs(ctx *gin.Context) (*gorm.DB, error) {
	value, ok := ctx.Get(collectionIDsKey)
	ids, valid := value.(*gorm.DB)
	if !ok || !valid || ids == nil {
		return nil, fmt.Errorf("collection requires an authorized ID query")
	}
	return ids.Session(&gorm.Session{}), nil
}

func legacyRequest(ctx *gin.Context) bool {
	account, err := cctx.AccountFromGinContext(ctx)
	return err == nil && account.Authority != nil && !authz.HasAuthority(ctx)
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

func requestAuthority(ctx *gin.Context, api permissions.API) (*authz.Authority, error) {
	account, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	authority, err := authz.AuthorityFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := authority.ValidateIdentity(account.ID, cctx.TokenIDFromContext(ctx), orgID, api); err != nil {
		return nil, err
	}
	return authority, nil
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
