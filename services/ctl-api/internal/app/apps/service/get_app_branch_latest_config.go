package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetAppBranchLatestConfig
// @Summary				get latest app branch config
// @Description			Returns the latest AppBranchConfig ordered by config_number (descending)
// @Tags					apps
// @Param					app_id			path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.AppBranchConfig
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/latest-config [get]
func (s *service) GetAppBranchLatestConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Feature flag checks
	if !org.Features[string(app.OrgFeatureAppBranches)] {
		ctx.Error(fmt.Errorf("app branches feature not enabled for this organization"))
		return
	}

	appID := ctx.Param("app_id")
	appBranchID := ctx.Param("app_branch_id")

	// Verify branch exists and belongs to this org/app
	var branch app.AppBranch
	res := s.db.WithContext(ctx).
		Where(app.AppBranch{
			OrgID: org.ID,
			AppID: appID,
		}).
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", res.Error))
		return
	}

	// Get latest config
	var config app.AppBranchConfig
	res = s.db.WithContext(ctx).
		Preload("ConnectedGithubVCSConfig").
		Preload("PublicGitVCSConfig").
		Preload("InstallGroups").
		Where("app_branch_id = ?", appBranchID).
		Order("config_number DESC").
		First(&config)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get latest app branch config: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, config)
}
