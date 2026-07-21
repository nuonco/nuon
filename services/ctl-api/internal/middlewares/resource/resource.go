// Package resource provides the object-level authorization middleware that runs
// after the org gate. For a request the org gate deferred (member without an
// org-wide permission), it resolves the resource named in the path and walks up
// its ownership chain looking for a grant that satisfies the request. Requests
// the org gate already authorized org-wide skip straight through.
package resource

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type Params struct {
	fx.In

	L   *zap.Logger
	DB  *gorm.DB `name:"psql"`
	Cfg *internal.Config
}

type middleware struct {
	l   *zap.Logger
	db  *gorm.DB
	cfg *internal.Config
}

func (m middleware) Name() string {
	return "resource"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Inert unless the feature is enabled. When disabled the org gate never
		// defers, so this middleware would no-op anyway; the guard makes that
		// explicit and cheap.
		if !m.cfg.ResourceGrantsEnabled {
			ctx.Next()
			return
		}

		// Public/global routes and requests already authorized org-wide need no
		// resource-level check.
		if cctx.IsGlobal(ctx) || cctx.IsPublic(ctx) || cctx.OrgAuthorized(ctx) {
			ctx.Next()
			return
		}

		// Collection endpoints authorize a deferred grantee by filtering their
		// result set in the handler, not by gating the whole route — so they are
		// let through here even when their path carries a resource param (e.g.
		// GET /v1/apps/:app_id/installs, where an install-only grantee has no app
		// grant to satisfy a gate). This check precedes resource resolution.
		if isFilteredCollection(ctx) {
			ctx.Next()
			return
		}

		org, err := cctx.OrgFromContext(ctx)
		if err != nil {
			// No org in context (e.g. a non-org-scoped route). Nothing to
			// authorize at the resource level; the org gate governs it.
			ctx.Next()
			return
		}

		chain, resolved, err := m.chainForRequest(ctx, org.ID)
		if err != nil {
			m.deny(ctx, fmt.Sprintf("unable to resolve resource in org %s", org.ID))
			return
		}

		// A deferred request that names no grantable resource and is not a
		// filtered collection fails closed.
		if !resolved {
			m.deny(ctx, "insufficient permissions for this org")
			return
		}

		acct, err := cctx.AccountFromGinContext(ctx)
		if err != nil {
			m.deny(ctx, "no account was identified")
			return
		}

		perm := permissions.FromRequest(ctx)
		if err := authz.Authorize(acct.AllPermissions, chain, perm); err != nil {
			m.deny(ctx, fmt.Sprintf("unable to perform %s on the requested resource", perm))
			return
		}

		ctx.Next()
	}
}

// chainForRequest resolves the most specific resource named in the path
// (install first, then app) scoped to the org, returning its ownership chain.
// resolved is false when the route names no grantable resource.
func (m middleware) chainForRequest(ctx *gin.Context, orgID string) (chain []string, resolved bool, err error) {
	if raw := ctx.Param("install_id"); raw != "" {
		var inst app.Install
		res := m.db.WithContext(ctx).
			Where("org_id = ?", orgID).
			Where(m.db.Where("id = ?", raw).Or("name = ?", raw)).
			First(&inst)
		if res.Error != nil {
			return nil, false, res.Error
		}
		return []string{inst.ID, inst.AppID, orgID}, true, nil
	}

	if raw := ctx.Param("app_id"); raw != "" {
		var a app.App
		res := m.db.WithContext(ctx).
			Where("org_id = ?", orgID).
			Where(m.db.Where("id = ?", raw).Or("name = ?", raw)).
			First(&a)
		if res.Error != nil {
			return nil, false, res.Error
		}
		return []string{a.ID, orgID}, true, nil
	}

	return nil, false, nil
}

func (m middleware) deny(ctx *gin.Context, description string) {
	ctx.Error(stderr.ErrAuthorization{
		Err:         fmt.Errorf("resource authorization failed"),
		Description: description,
	})
	ctx.Abort()
}

func New(params Params) *middleware {
	return &middleware{
		l:   params.L,
		db:  params.DB,
		cfg: params.Cfg,
	}
}
