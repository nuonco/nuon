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

func NativeRoute(router gin.IRouter, api permissions.API, operation permissions.PermissionID,
	load func(*gin.Context) (*authz.NativeAuthority, error),
	resolve func(*gin.Context) (permissions.NativeTarget, error),
) *gin.RouterGroup {
	return group(router, func(ctx *gin.Context) {
		authority, err := nativeAuthority(ctx, api, load)
		if err != nil {
			nativeFailure(ctx, err)
			return
		}
		target, err := resolve(ctx)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			nativeDenied(ctx, err)
			return
		}
		if err != nil {
			nativeFailure(ctx, err)
			return
		}
		if match := authority.Check(operation, target); !match.Matched {
			nativeDenied(ctx, fmt.Errorf("operation %s denied: %s", operation, match.Reason))
			return
		}
		ctx.Next()
	})
}

const nativeCollectionIDsKey = "nuon.authz.native_collection_ids"

func NativeCollectionRoute(router gin.IRouter, api permissions.API, operation permissions.PermissionID,
	load func(*gin.Context) (*authz.NativeAuthority, error),
	resources func(*gin.Context) (*authz.NativeResources, error),
) *gin.RouterGroup {
	return group(router, func(ctx *gin.Context) {
		authority, err := nativeAuthority(ctx, api, load)
		if err != nil {
			nativeFailure(ctx, err)
			return
		}
		set, err := resources(ctx)
		if err != nil {
			nativeFailure(ctx, err)
			return
		}
		ids, err := authority.AuthorizedIDs(ctx, operation, set)
		if err != nil {
			nativeFailure(ctx, err)
			return
		}
		ctx.Set(nativeCollectionIDsKey, ids)
		ctx.Next()
	})
}

func NativeCollectionIDs(ctx *gin.Context) (*gorm.DB, error) {
	value, ok := ctx.Get(nativeCollectionIDsKey)
	ids, valid := value.(*gorm.DB)
	if !ok || !valid || ids == nil {
		return nil, fmt.Errorf("native collection requires an authorized ID query")
	}
	return ids.Session(&gorm.Session{}), nil
}

func nativeAuthority(ctx *gin.Context, api permissions.API, load func(*gin.Context) (*authz.NativeAuthority, error)) (*authz.NativeAuthority, error) {
	account, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	authority, err := load(ctx)
	if err != nil {
		return nil, err
	}
	if err := authority.ValidateIdentity(account.ID, cctx.TokenIDFromContext(ctx), orgID, api); err != nil {
		return nil, err
	}
	return authority, nil
}

func nativeDenied(ctx *gin.Context, err error) {
	cctx.GetLogger(ctx, zap.NewNop()).Info("native route authorization denied", zap.String("route", ctx.FullPath()), zap.Error(err))
	ctx.Error(stderr.ErrNotFound{Err: errors.New("resource not found"), Description: "resource not found"})
	ctx.Abort()
}

func nativeFailure(ctx *gin.Context, err error) {
	cctx.GetLogger(ctx, zap.NewNop()).Error("native route authorization failed", zap.String("route", ctx.FullPath()), zap.Error(err))
	ctx.Error(stderr.ErrSystem{Err: errors.New("unable to authorize request"), Description: "unable to authorize request"})
	ctx.Abort()
}
