package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type PreviewInstallCandidatesResponse struct {
	Installs []app.Install `json:"installs"`
}

// @ID						GetAppBranchPreviewInstallCandidates
// @Summary				list preview install candidates for an app branch
// @Description			Lists all installs on the app for preview run selection (includes installs on other branches)
// @Tags					apps
// @Param					app_id			path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					config_id		query	string	false	"branch config ID (defaults to latest)"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	PreviewInstallCandidatesResponse
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/preview-install-candidates [get]
func (s *service) GetAppBranchPreviewInstallCandidates(ctx *gin.Context) {
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
	configID := ctx.Query("config_id")

	var branch app.AppBranch
	res := s.db.WithContext(ctx).
		Where(app.AppBranch{OrgID: org.ID, AppID: appID}).
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", res.Error))
		return
	}

	var config app.AppBranchConfig
	if configID != "" {
		res = s.db.WithContext(ctx).
			Where("app_branch_id = ?", appBranchID).
			First(&config, "id = ?", configID)
	} else {
		res = s.db.WithContext(ctx).
			Where("app_branch_id = ?", appBranchID).
			Order("config_number DESC").
			First(&config)
	}
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find branch config: %w", res.Error))
		return
	}

	previewCfg := helpers.BranchPreviewConfigOrDefault(&config)
	installs, err := s.helpers.ListPreviewInstallCandidates(ctx, appID, appBranchID, previewCfg)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list preview install candidates: %w", err))
		return
	}
	if installs == nil {
		installs = []app.Install{}
	}

	ctx.JSON(http.StatusOK, PreviewInstallCandidatesResponse{Installs: installs})
}
