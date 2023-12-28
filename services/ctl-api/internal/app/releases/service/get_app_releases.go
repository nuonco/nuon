package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetAppReleases
// @Summary	get all releases for an app
// @Description.markdown	get_app_releases.md
// @Param			app_id	path	string	true	"app ID"
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
// @Success		200				{array}		app.ComponentRelease
// @Router			/v1/apps/{app_id}/releases [GET]
func (s *service) GetAppReleases(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	installs, err := s.getAppReleases(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installs)
}

func (s *service) getAppReleases(ctx context.Context, appID string) ([]app.ComponentRelease, error) {
	var releases []app.ComponentRelease
	res := s.db.WithContext(ctx).
		// join component-releases to component-builds to component-config-connections to components
		Joins("JOIN component_builds ON component_builds.id=component_releases.component_build_id").
		Joins("JOIN component_config_connections ON component_config_connections.id=component_builds.component_config_connection_id").
		Joins("JOIN components ON components.id=component_config_connections.component_id").
		Where("components.app_id = ?", appID).
		Order("created_at DESC").
		Find(&releases)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to load component releases")
	}

	return releases, nil
}
