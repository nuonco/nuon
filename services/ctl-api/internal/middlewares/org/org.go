package org

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

const (
	orgIDHeaderKey string = "X-Nuon-Org-ID"
	orgCtxKey      string = "org"
)

func FromContext(ctx *gin.Context) (*app.Org, error) {
	org, exists := ctx.Get(orgCtxKey)
	if !exists {
		return nil, fmt.Errorf("org was not set on middleware context")
	}

	return org.(*app.Org), nil
}

func New(writer metrics.Writer, db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orgID := ctx.Request.Header.Get(orgIDHeaderKey)
		if orgID == "" {
			ctx.Error(fmt.Errorf("org header missing, please set %s", orgIDHeaderKey))
			return
		}

		org := app.Org{}
		res := db.WithContext(ctx).First(&org, "id = ?", orgID)
		if res.Error != nil {
			ctx.Error(fmt.Errorf("unable to get org: %w", res.Error))
			return
		}

		ctx.Set(orgCtxKey, &org)
		ctx.Next()
	}
}
