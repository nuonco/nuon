package require

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	applookup "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/lookup"
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

		a, err := applookup.ByRef(ctx.Request.Context(), db, orgID, appID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Error(fmt.Errorf("unable to authorize app: %w", err))
			ctx.Abort()
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

		cctx.SetResolvedAppGinContext(ctx, appID, a)
		ctx.Next()
	}
}
