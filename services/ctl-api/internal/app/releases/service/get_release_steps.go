package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

// @ID GetReleaseSteps
// @Summary	get a release
// @Description.markdown	get_release.md
// @Param			release_id	path	string	true	"release ID"
// @Tags			releases
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.ComponentReleaseStep
// @Router			/v1/releases/{release_id}/steps [get]
func (s *service) GetReleaseSteps(ctx *gin.Context) {
	releaseID := ctx.Param("release_id")
	steps, err := s.getReleaseSteps(ctx, releaseID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get release %s: %w", releaseID, err))
		return
	}

	ctx.JSON(http.StatusOK, steps)
}

func (s *service) getReleaseSteps(ctx context.Context, releaseID string) ([]app.ComponentReleaseStep, error) {
	var release app.ComponentRelease
	res := s.db.WithContext(ctx).
		Preload("ComponentReleaseSteps", func(db *gorm.DB) *gorm.DB {
			return db.Order("component_release_steps.created_at DESC")
		}).
		Preload("ComponentReleaseSteps.InstallDeploys").
		First(&release, "id = ?", releaseID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get release: %w", res.Error)
	}

	return release.ComponentReleaseSteps, nil
}
