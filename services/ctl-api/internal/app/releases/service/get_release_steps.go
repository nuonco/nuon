package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
	"gorm.io/gorm"
)

// @ID						GetReleaseSteps
// @Summary				get a release
// @Description.markdown	get_release.md
// @Param					release_id					path	string	true	"release ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					releases
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.ComponentReleaseStep
// @Router					/v1/releases/{release_id}/steps [get]
func (s *service) GetReleaseSteps(ctx *gin.Context) {
	releaseID := ctx.Param("release_id")
	steps, err := s.getReleaseSteps(ctx, releaseID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get release %s: %w", releaseID, err))
		return
	}

	ctx.JSON(http.StatusOK, steps)
}

func (s *service) getReleaseSteps(ctx *gin.Context, releaseID string) ([]app.ComponentReleaseStep, error) {
	var release app.ComponentRelease
	res := s.db.WithContext(ctx).
		Preload("ComponentReleaseSteps", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOffsetPagination).
				Order("component_release_steps.created_at DESC")
		}).
		Preload("ComponentReleaseSteps.InstallDeploys").
		First(&release, "id = ?", releaseID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get release: %w", res.Error)
	}

	steps, err := db.HandlePaginatedResponse(ctx, release.ComponentReleaseSteps)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	release.ComponentReleaseSteps = steps
	return release.ComponentReleaseSteps, nil
}
