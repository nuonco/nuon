package require

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// AppInOrg rejects requests whose :app_id is not an app in the selected org.
// Org-level "all" makes permission object checks succeed for arbitrary app IDs
// under the header org, so route handlers must also prove ownership.
func AppInOrg(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsGlobal(ctx) || cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}

		appID := ctx.Param("app_id")
		if appID == "" {
			ctx.Next()
			return
		}

		orgID, err := cctx.OrgIDFromContext(ctx)
		if err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}

		var n int64
		err = db.WithContext(ctx.Request.Context()).
			Model(&app.App{}).
			Where("org_id = ?", orgID).
			Where("id = ? OR name = ?", appID, appID).
			Limit(1).
			Count(&n).Error
		if err != nil {
			ctx.Error(fmt.Errorf("unable to authorize app: %w", err))
			ctx.Abort()
			return
		}
		if n < 1 {
			cctx.GetLogger(ctx, zap.NewNop()).Info("route authorization denied",
				zap.String("route", ctx.FullPath()),
				zap.Error(fmt.Errorf("app %s is not in org %s", appID, orgID)),
			)
			ctx.Error(stderr.ErrNotFound{
				Err:         errors.New("app not found"),
				Description: "app not found",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
