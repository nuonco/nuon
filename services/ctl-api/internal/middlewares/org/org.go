package org

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/global"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/public"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

type middleware struct {
	l  *zap.Logger
	db *gorm.DB
}

func (m middleware) Name() string {
	return "org"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if global.IsGlobal(ctx) || public.IsPublic(ctx) {
			ctx.Next()
			return
		}

		orgID := ctx.Request.Header.Get(orgIDHeaderKey)
		if orgID == "" {
			ctx.Error(stderr.ErrAuthorization{
				Err:         fmt.Errorf("required header %s not found", orgIDHeaderKey),
				Description: fmt.Sprintf("please retry request with %s header", orgIDHeaderKey),
			})
			ctx.Abort()
			return
		}

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

		SetGinContext(ctx, &org)

		metricCtx, err := metrics.FromContext(ctx)
		if err == nil {
			metricCtx.OrgID = orgID
		}

		ctx.Next()
	}
}

func New(l *zap.Logger, db *gorm.DB) *middleware {
	return &middleware{
		l:  l,
		db: db,
	}
}
