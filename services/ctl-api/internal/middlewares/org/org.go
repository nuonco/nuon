package org

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

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

		// authorize from the URL: the route's path params name the resource's
		// ownership chain, and any link carrying a sufficient grant allows the
		// request. Org is the last link of every chain, so managed org roles
		// authorize everywhere exactly as they did before scoped permissions.
		perm := permissions.FromRequest(ctx)
		chain := chainFromParams(ctx.Param, org.ID)
		authorizedBy, err := authz.Decide(acct.AllPermissions, acct.OrgTypeGrants(org.ID), chain, perm)
		if err != nil {
			m.logDecision(ctx, acct, org.ID, perm, chain, "deny", authorizedBy)

			if !acct.HasOrg(org.ID) {
				ctx.Error(stderr.ErrAuthorization{
					Err:         fmt.Errorf("account has no access to org %s", org.ID),
					Description: "you do not have access to this organization",
				})
				ctx.Abort()
				return
			}

			ctx.Error(permissionDeniedError(acct, org.ID, perm, scopeFromPath(ctx.FullPath())))
			ctx.Abort()
			return
		}
		m.logDecision(ctx, acct, org.ID, perm, chain, "allow", authorizedBy)

		cctx.SetOrgGinContext(ctx, &org)
		metricCtx, err := cctx.MetricsContextFromGinContext(ctx)
		if err == nil {
			metricCtx.OrgID = orgID
		}

		ctx.Next()
	}
}

// logDecision records why a request was allowed or denied. Denials log at info
// so they are visible without turning on debug logging for the whole service;
// allows log at debug, since every authorized request produces one.
func (m middleware) logDecision(ctx *gin.Context, acct *app.Account, orgID string, perm permissions.Permission, chain []authz.Link, decision, authorizedBy string) {
	fields := []zap.Field{
		zap.String("route", ctx.FullPath()),
		zap.String("method", ctx.Request.Method),
		zap.String("permission", string(perm)),
		zap.String("chain", authz.ChainString(chain)),
		zap.String("decision", decision),
		zap.String("authorized_by", authorizedBy),
		zap.String("account_id", acct.ID),
		zap.String("org_id", orgID),
	}

	if decision == "allow" {
		m.l.Debug("authz decision", fields...)
		return
	}

	m.l.Info("authz decision", append(fields,
		zap.String("missing", fmt.Sprintf("needs %s on %s", perm, authz.ChainString(chain))),
	)...)
}

func New(params Params) *middleware {
	return &middleware{
		l:  params.L,
		db: params.DB,
	}
}
