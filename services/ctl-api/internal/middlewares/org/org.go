package org

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

const (
	orgIDHeaderKey string = "X-Nuon-Org-ID"
)

type Params struct {
	fx.In

	L  *zap.Logger
	DB *gorm.DB `name:"psql"`
}

type middleware struct {
	l  *zap.Logger
	db *gorm.DB
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
				Description: fmt.Sprint("no account was set in the middleware"),
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
		err = acct.AllPermissions.CanPerform(org.ID, perm)
		if err != nil {
			ctx.Error(stderr.ErrAuthorization{
				Err:         fmt.Errorf("unable to perform %s on org %s", perm, org.ID),
				Description: fmt.Sprintf("Please make sure you have the correct permissions for %s", org.ID),
			})
			ctx.Abort()
			return
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
		l:  params.L,
		db: params.DB,
	}
}
