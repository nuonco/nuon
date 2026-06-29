package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

// @ID						GetAppBranchRunInstallGroups
// @Summary				get install group deployments for an app branch run
// @Description			Returns install config updates triggered by a specific app branch run, grouped by install group
// @Tags					apps
// @Param					app_id			path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					run_id			path	string	true	"app branch run ID"
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.InstallConfigUpdate
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/runs/{run_id}/install-groups [get]
func (s *service) GetAppBranchRunInstallGroups(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureAppBranches)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check feature: %w", err))
		return
	}
	if !enabled {
		ctx.Error(features.ErrFeatureNotEnabled(app.OrgFeatureAppBranches))
		return
	}

	appID := ctx.Param("app_id")
	appBranchID := ctx.Param("app_branch_id")
	runID := ctx.Param("run_id")

	var branch app.AppBranch
	res := s.db.WithContext(ctx).
		Where(app.AppBranch{OrgID: org.ID, AppID: appID}).
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", res.Error))
		return
	}

	var run app.AppBranchRun
	res = s.db.WithContext(ctx).
		Where(app.AppBranchRun{AppBranchID: appBranchID}).
		First(&run, "id = ?", runID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch run: %w", res.Error))
		return
	}

	var updates []app.InstallConfigUpdate
	res = s.db.WithContext(ctx).
		Preload("Install").
		Preload("Workflow").
		Preload("Workflow.Steps").
		Where(app.InstallConfigUpdate{AppBranchRunID: runID}).
		Order("created_at ASC").
		Find(&updates)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install config updates: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, updates)
}
