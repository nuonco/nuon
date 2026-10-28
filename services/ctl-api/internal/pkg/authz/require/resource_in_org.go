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

// PathOrgMatches rejects requests whose :org_id path param does not match the
// selected org from the auth header. Handlers often load the header org and
// ignore the path, which lets callers probe arbitrary org IDs with a 200.
func PathOrgMatches() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsGlobal(ctx) || cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}

		pathOrgID := ctx.Param("org_id")
		if pathOrgID == "" {
			ctx.Next()
			return
		}

		orgID, err := cctx.OrgIDFromContext(ctx)
		if err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}
		if pathOrgID != orgID {
			cctx.GetLogger(ctx, zap.NewNop()).Info("route authorization denied",
				zap.String("route", ctx.FullPath()),
				zap.Error(fmt.Errorf("path org %s does not match selected org %s", pathOrgID, orgID)),
			)
			ctx.Error(stderr.ErrNotFound{
				Err:         errors.New("org not found"),
				Description: "org not found",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// ComponentInOrg rejects requests whose :component_id is not in the selected org.
func ComponentInOrg(db *gorm.DB) gin.HandlerFunc {
	return resourceInOrg(db, "component_id", "component", &app.Component{})
}

// RunnerInOrg rejects requests whose :runner_id is not in the selected org.
func RunnerInOrg(db *gorm.DB) gin.HandlerFunc {
	return resourceInOrg(db, "runner_id", "runner", &app.Runner{})
}

// WorkflowInOrg rejects requests whose :workflow_id is not in the selected org.
func WorkflowInOrg(db *gorm.DB) gin.HandlerFunc {
	return resourceInOrg(db, "workflow_id", "workflow", &app.Workflow{})
}

// InstallWorkflowInOrg rejects requests whose :install_workflow_id is not in the selected org.
func InstallWorkflowInOrg(db *gorm.DB) gin.HandlerFunc {
	return resourceInOrg(db, "install_workflow_id", "workflow", &app.Workflow{})
}

func resourceInOrg(db *gorm.DB, param, label string, model any) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsGlobal(ctx) || cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}

		resourceID := ctx.Param(param)
		if resourceID == "" {
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
			Model(model).
			Where("org_id = ?", orgID).
			Where("id = ?", resourceID).
			Limit(1).
			Count(&n).Error
		if err != nil {
			ctx.Error(fmt.Errorf("unable to authorize %s: %w", label, err))
			ctx.Abort()
			return
		}
		if n < 1 {
			cctx.GetLogger(ctx, zap.NewNop()).Info("route authorization denied",
				zap.String("route", ctx.FullPath()),
				zap.Error(fmt.Errorf("%s %s is not in org %s", label, resourceID, orgID)),
			)
			ctx.Error(stderr.ErrNotFound{
				Err:         fmt.Errorf("%s not found", label),
				Description: fmt.Sprintf("%s not found", label),
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
