package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@BasePath	/v1/releases
//
// Create get a release
//
//	@Summary	get a release
//	@Schemes
//	@Description	get a release
//	@Param			release_id	path	string	true	"release ID"
//	@Tags			releases
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{object}	app.ComponentRelease
//	@Router			/v1/releases/{release_id} [get]
func (s *service) GetRelease(ctx *gin.Context) {
	releaseID := ctx.Param("release_id")
	app, err := s.getRelease(ctx, releaseID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get release %s: %w", releaseID, err))
		return
	}

	ctx.JSON(http.StatusOK, app)
}

func (s *service) getRelease(ctx context.Context, releaseID string) (*app.ComponentRelease, error) {
	release := app.ComponentRelease{}
	res := s.db.WithContext(ctx).
		Preload("ComponentBuild").
		Preload("ComponentReleaseSteps").
		First(&release, "id = ?", releaseID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get release: %w", res.Error)
	}
	release.TotalComponentReleaseSteps = len(release.ComponentReleaseSteps)

	return &release, nil
}
