package require

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type customerTargetKey struct{}

func CustomerTarget(ctx context.Context) (permissions.ResolvedTarget, bool) {
	target, ok := customerContext(ctx).Value(customerTargetKey{}).(permissions.ResolvedTarget)
	return target, ok
}

func customerContext(ctx context.Context) context.Context {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		requestCtx := ginCtx.Request.Context()
		if account, err := cctx.AccountFromGinContext(ginCtx); err == nil {
			requestCtx = cctx.SetAccountContext(requestCtx, account)
		}
		if org, err := cctx.OrgFromContext(ginCtx); err == nil {
			requestCtx = cctx.SetOrgContext(requestCtx, org)
		}
		return requestCtx
	}
	return ctx
}

func CustomerRead(ctx context.Context, db *gorm.DB, logger *zap.Logger, req authz.ReadRequest) (permissions.ResolvedTarget, error) {
	ctx = customerContext(ctx)
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil || orgID != req.Target.OrgID {
		return permissions.ResolvedTarget{}, customerDenied(req.Target.Kind)
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client := authz.New(authz.Params{DB: db})
	ctx, err = client.PrepareCustomer(ctx, orgID)
	log := cctx.GetLogger(ctx, logger).With(zap.String("org_id", orgID), zap.String("permission", string(req.Permission)))
	if err != nil {
		log.Warn("customer read preparation failed", zap.Error(err))
		return permissions.ResolvedTarget{}, customerDenied(req.Target.Kind)
	}
	decision, err := client.AuthorizeRead(ctx, req)
	log.Info("customer read authorization", zap.Bool("allowed", decision.Allowed), zap.String("reason", string(decision.Reason)), zap.Error(err))
	if err != nil || !decision.Allowed {
		return permissions.ResolvedTarget{}, customerDenied(req.Target.Kind)
	}
	return decision.Target, nil
}

func CustomerMutation(ctx context.Context, db *gorm.DB, logger *zap.Logger, req authz.MutationRequest, mutate func(*gorm.DB, permissions.ResolvedTarget) error) error {
	ctx = customerContext(ctx)
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil || orgID != req.Target.OrgID {
		return customerDenied(permissions.KindOrg)
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client := authz.New(authz.Params{DB: db})
	ctx, err = client.PrepareCustomer(ctx, orgID)
	log := cctx.GetLogger(ctx, logger).With(zap.String("org_id", orgID), zap.String("permission", string(req.Permission)))
	if err != nil {
		log.Warn("customer mutation preparation failed", zap.Error(err))
		return customerDenied(permissions.KindOrg)
	}
	decision, err := client.MutateCustomer(ctx, req, mutate)
	if err != nil {
		log.Warn("customer mutation failed", zap.Error(err))
		return err
	}
	log.Info("customer mutation authorization", zap.Bool("allowed", decision.Allowed), zap.String("reason", string(decision.Reason)))
	if !decision.Allowed {
		return customerDenied(permissions.KindOrg)
	}
	return nil
}

func CustomerReadRoute(router gin.IRouter, db *gorm.DB, logger *zap.Logger, permission permissions.PermissionID, kind permissions.ResourceKind, param string) *gin.RouterGroup {
	route := group(router, func(ctx *gin.Context) {
		orgID, err := cctx.OrgIDFromContext(ctx)
		if err != nil {
			ctx.Error(customerDenied(kind))
			ctx.Abort()
			return
		}
		id := orgID
		if kind != permissions.KindOrg {
			id, err = CustomerResourceID(ctx, db, kind, orgID, ctx.Param(param))
			if err != nil {
				ctx.Error(err)
				ctx.Abort()
				return
			}
		}
		target, err := CustomerRead(ctx, db, cctx.GetLogger(ctx, logger), authz.ReadRequest{
			Permission: permission, Target: permissions.Scope{Kind: kind, OrgID: orgID, ID: id},
		})
		if err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}
		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), customerTargetKey{}, target))
	})
	return CustomerRoute(route)
}

func CustomerRoute(router gin.IRouter) *gin.RouterGroup {
	route := group(router)
	route.Handlers = append(gin.HandlersChain{func(ctx *gin.Context) { ctx.Set("nuon.authz.customer", true) }}, route.Handlers...)
	return route
}

func IsCustomerRoute(ctx *gin.Context) bool {
	return ctx.GetBool("nuon.authz.customer")
}

func CustomerResourceID(ctx context.Context, db *gorm.DB, kind permissions.ResourceKind, orgID, ref string) (string, error) {
	db = db.WithContext(ctx).Scopes(scopes.WithDisableViews).Select("id")
	var id string
	var err error
	switch kind {
	case permissions.KindInstall:
		var row app.Install
		err = db.Where(app.Install{OrgID: orgID}).Where(db.Where(app.Install{ID: ref}).Or(app.Install{Name: ref})).First(&row).Error
		id = row.ID
	case permissions.KindApp:
		var row app.App
		err = db.Where(app.App{OrgID: orgID}).Where(db.Where(app.App{ID: ref}).Or(app.App{Name: ref})).First(&row).Error
		id = row.ID
	default:
		return "", customerDenied(kind)
	}
	if ref == "" || err != nil {
		return "", customerDenied(kind)
	}
	return id, nil
}

func customerDenied(kind permissions.ResourceKind) error {
	if kind == permissions.KindOrg {
		return stderr.ErrAuthorization{Err: errors.New("access denied"), Description: "access denied"}
	}
	description := noun(kind) + " not found"
	return stderr.ErrNotFound{Err: errors.New(description), Description: description}
}
