package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

// @ID						GetInstallGroupRun
// @Summary				get a specific install group run
// @Description			Returns a single install group run with full details
// @Tags					apps
// @Param					app_id					path	string	true	"app ID"
// @Param					app_branch_id			path	string	true	"app branch ID"
// @Param					run_id					path	string	true	"app branch run ID"
// @Param					install_group_run_id	path	string	true	"install group run ID"
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.InstallGroupRun
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/runs/{run_id}/install-group-runs/{install_group_run_id} [get]
func (s *service) GetInstallGroupRun(ctx *gin.Context) {
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
	installGroupRunID := ctx.Param("install_group_run_id")

	var branch app.AppBranch
	res := s.db.WithContext(ctx).
		Where(app.AppBranch{OrgID: org.ID, AppID: appID}).
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", res.Error))
		return
	}

	var groupRun app.InstallGroupRun
	res = s.db.WithContext(ctx).
		Preload("InstallGroup").
		Where(app.InstallGroupRun{OrgID: org.ID}).
		First(&groupRun, "id = ?", installGroupRunID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find install group run: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, groupRun)
}
