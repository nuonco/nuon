package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

// @ID						GetAppBranchPreviewSources
// @Summary				list preview sources for an app branch
// @Description			Returns open pull requests targeting the branch and other git branches in the repo
// @Tags					apps
// @Param					app_id			path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	helpers.ListPreviewSourcesResult
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/preview-sources [get]
func (s *service) GetAppBranchPreviewSources(ctx *gin.Context) {
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

	var branch app.AppBranch
	res := s.db.WithContext(ctx).
		Where(app.AppBranch{OrgID: org.ID, AppID: appID}).
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", res.Error))
		return
	}

	var config app.AppBranchConfig
	res = s.db.WithContext(ctx).
		Preload("ConnectedGithubVCSConfig.VCSConnection").
		Preload("PublicGitVCSConfig").
		Where("app_branch_id = ?", appBranchID).
		Order("config_number DESC").
		First(&config)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find branch config: %w", res.Error))
		return
	}

	sources, err := s.helpers.ListPreviewSources(ctx, &branch, &config)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list preview sources: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, sources)
}
