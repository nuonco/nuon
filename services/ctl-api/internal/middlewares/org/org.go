package org

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

const (
	orgIDHeaderKey string = "X-Nuon-Org-ID"
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
	return "org"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsGlobal(ctx) || cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}

		orgID := ctx.Request.Header.Get(orgIDHeaderKey)
		if orgID == "" {
			orgID = ctx.Query("org_id")
		}

		if orgID == "" {
			ctx.Error(stderr.ErrAuthorization{
				Err:         fmt.Errorf("org ID was empty"),
				Description: fmt.Sprintf("please retry request with %s header or org query param", orgIDHeaderKey),
			})
			ctx.Abort()
			return
		}

		acct, err := cctx.AccountFromGinContext(ctx)
		if err != nil {
			ctx.Error(stderr.ErrAuthorization{
				Err:         fmt.Errorf("no account identified"),
				Description: "no account was set in the middleware",
			})
			ctx.Abort()
			return
		}

		// make sure org exists
		org := app.Org{}
		res := m.db.WithContext(ctx).
			Preload("NotificationsConfig").
			First(&org, "id = ?", orgID)
		if res.Error != nil {
			ctx.Error(stderr.ErrAuthorization{
				Err:         fmt.Errorf("org %s was not found", orgID),
				Description: "please make sure org ID is set properly",
			})
			ctx.Abort()
			return
		}

		// make sure account has access to org
		perm := permissions.FromRequest(ctx)
		object := permissions.RequestObject(ctx, org.ID)
		orgErr := acct.AllPermissions.CanPerform(object, perm)

		if m.cfg.ResourceGrantsEnabled {
			// Split membership from authorization. An org-wide grant takes the
			// fast path; otherwise a member is authorized at the resource level
			// (or, for a grant-filtered collection, in the handler). Anything
			// else fails closed.
			if orgErr == nil {
				cctx.SetOrgAuthorized(ctx, true)
			} else {
				if !acct.HasOrg(org.ID) {
					ctx.Error(stderr.ErrAuthorization{
						Err:         fmt.Errorf("account has no access to org %s", org.ID),
						Description: fmt.Sprintf("Please make sure you have access to %s", org.ID),
					})
					ctx.Abort()
					return
				}
				cctx.SetOrgAuthorized(ctx, false)

				if !isFilteredCollection(ctx) {
					if err := m.authorizeResource(ctx, acct, org.ID, perm); err != nil {
						ctx.Error(err)
						ctx.Abort()
						return
					}
				}
			}
		} else {
			if orgErr != nil {
				ctx.Error(stderr.ErrAuthorization{
					Err:         fmt.Errorf("unable to perform %s on object %s", perm, object),
					Description: fmt.Sprintf("Please make sure you have the correct permissions for %s", object),
				})
				ctx.Abort()
				return
			}
			// Legacy path: reaching here means the account has org-wide access,
			// so downstream grant-scope filtering must no-op.
			cctx.SetOrgAuthorized(ctx, true)
		}

		cctx.SetOrgGinContext(ctx, &org)
		metricCtx, err := cctx.MetricsContextFromGinContext(ctx)
		if err == nil {
			metricCtx.OrgID = orgID
		}

		ctx.Next()
	}
}

// authorizeResource resolves the most specific resource named in the path
// (install, then app; name-or-id, org-scoped) and authorizes it via the walk-up
// primitive. A deferred request that names no grantable resource fails closed.
func (m middleware) authorizeResource(ctx *gin.Context, acct *app.Account, orgID string, perm permissions.Permission) error {
	chain, resolved, err := m.resourceChain(ctx, orgID)
	if err != nil {
		return stderr.ErrAuthorization{
			Err:         fmt.Errorf("unable to resolve resource in org %s: %w", orgID, err),
			Description: "the requested resource could not be found in this org",
		}
	}
	if !resolved {
		return stderr.ErrAuthorization{
			Err:         fmt.Errorf("no org-wide permission and no grantable resource in path"),
			Description: fmt.Sprintf("Please make sure you have the correct permissions for %s", orgID),
		}
	}
	if err := authz.Authorize(acct.AllPermissions, acct.OrgTypeGrants(orgID), chain, perm); err != nil {
		return stderr.ErrAuthorization{
			Err:         fmt.Errorf("unable to perform %s on the requested resource", perm),
			Description: "you do not have access to the requested resource",
		}
	}
	return nil
}

// resourceChain builds the ownership chain of the most specific resource named
// in the path (install first, then app, then org-owned resources like webhooks
// and VCS connections), each link tagged with its grant resource type so
// wildcard grants can authorize by tier. resolved is false when the route names
// no grantable resource.
func (m middleware) resourceChain(ctx *gin.Context, orgID string) (chain []authz.Link, resolved bool, err error) {
	if raw := ctx.Param("install_id"); raw != "" {
		var inst app.Install
		res := m.db.WithContext(ctx).
			Where("org_id = ?", orgID).
			Where(m.db.Where("id = ?", raw).Or("name = ?", raw)).
			First(&inst)
		if res.Error != nil {
			return nil, false, res.Error
		}
		return []authz.Link{
			{Type: string(app.GrantResourceTypeInstall), ID: inst.ID},
			{Type: string(app.GrantResourceTypeApp), ID: inst.AppID},
			{Type: string(app.GrantResourceTypeOrg), ID: orgID},
		}, true, nil
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
		return []authz.Link{
			{Type: string(app.GrantResourceTypeApp), ID: a.ID},
			{Type: string(app.GrantResourceTypeOrg), ID: orgID},
		}, true, nil
	}

	if route, ok := matchOrgResourceRoute(ctx.FullPath()); ok {
		orgLink := authz.Link{Type: string(app.GrantResourceTypeOrg), ID: orgID}

		raw := ctx.Param(route.idParam)
		if raw == "" {
			return []authz.Link{
				{Type: string(route.resourceType)},
				orgLink,
			}, true, nil
		}

		res := m.db.WithContext(ctx).
			Where("org_id = ?", orgID).
			Where("id = ?", raw).
			First(route.model())
		if res.Error != nil {
			return nil, false, res.Error
		}
		return []authz.Link{
			{Type: string(route.resourceType), ID: raw},
			orgLink,
		}, true, nil
	}

	return nil, false, nil
}

func New(params Params) *middleware {
	return &middleware{
		l:   params.L,
		db:  params.DB,
		cfg: params.Cfg,
	}
}
