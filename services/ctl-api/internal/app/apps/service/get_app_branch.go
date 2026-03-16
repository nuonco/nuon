package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

// @ID						GetAppBranch
// @Summary				get an app branch
// @Description.markdown	get_app_branch.md
// @Param					app_id			path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					latest_config	query	bool	false	"include only the latest config"	Default(false)
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.AppBranch
// @Router					/v1/apps/{app_id}/branches/{app_branch_id} [get]
func (s *service) GetAppBranch(ctx *gin.Context) {
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
	latestConfig := ctx.DefaultQuery("latest_config", "false") == "true"

	branch, err := s.getAppBranch(ctx, org.ID, appID, appBranchID, latestConfig)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, branch)
}

func (s *service) getAppBranch(ctx context.Context, orgID, appID, appBranchID string, latestConfig bool) (*app.AppBranch, error) {
	branch := app.AppBranch{}

	query := s.db.WithContext(ctx).
		Where(app.AppBranch{
			OrgID: orgID,
			AppID: appID,
		}).
		Preload("Queue")

	if latestConfig {
		// Only preload the latest config with its relationships
		query = query.Preload("Configs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_branch_configs_view_v1.created_at DESC").Limit(1)
		}).
			Preload("Configs.PublicGitVCSConfig").
			Preload("Configs.ConnectedGithubVCSConfig").
			Preload("Configs.InstallGroups", func(db *gorm.DB) *gorm.DB {
				return db.Order("\"order\" ASC")
			})
	}

	res := query.First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app branch: %w", res.Error)
	}

	return &branch, nil
}
