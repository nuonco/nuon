package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

// @ID						DeleteAppBranch
// @Summary				delete an app branch
// @Description			Deletes an app branch and all associated configs, runs, and install group runs.
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
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/apps/{app_id}/branches/{app_branch_id} [DELETE]
func (s *service) DeleteAppBranch(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	enabled, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check features: %w", err))
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
		Where(app.AppBranch{
			OrgID: org.ID,
			AppID: appID,
		}).
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", res.Error))
		return
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		if err := tx.Model(&app.InstallAppBranchConnection{}).
			Where(app.InstallAppBranchConnection{
				AppBranchID: appBranchID,
				Active:      true,
			}).
			Updates(map[string]interface{}{
				"active":         false,
				"deactivated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("unable to deactivate install branch connections: %w", err)
		}

		if err := tx.Model(&app.Install{}).
			Where("app_branch_id = ?", appBranchID).
			Update("app_branch_id", nil).Error; err != nil {
			return fmt.Errorf("unable to release installs from branch: %w", err)
		}

		if err := tx.Delete(&branch).Error; err != nil {
			return fmt.Errorf("unable to delete app branch: %w", err)
		}

		return nil
	}); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
