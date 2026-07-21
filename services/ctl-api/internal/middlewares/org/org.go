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
			// Split membership from authorization: an org-wide grant takes the
			// fast path; a member without one defers to a downstream resource
			// middleware (which fails closed if none authorizes the request).
			if orgErr != nil {
				if !acct.HasOrg(org.ID) {
					ctx.Error(stderr.ErrAuthorization{
						Err:         fmt.Errorf("account has no access to org %s", org.ID),
						Description: fmt.Sprintf("Please make sure you have access to %s", org.ID),
					})
					ctx.Abort()
					return
				}
				cctx.SetOrgAuthorized(ctx, false)
			} else {
				cctx.SetOrgAuthorized(ctx, true)
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

func New(params Params) *middleware {
	return &middleware{
		l:   params.L,
		db:  params.DB,
		cfg: params.Cfg,
	}
}
