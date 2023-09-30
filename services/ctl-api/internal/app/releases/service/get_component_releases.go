package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@BasePath	/v1/components
// Get all releases for a component
//	@Summary	get all releases for a component
//	@Schemes
//	@Description	get all releases for a component
//	@Param			component_id	path	string	true	"component ID"
//	@Tags			releases
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	app.ComponentRelease
//	@Router			/v1/components/{component_id}/releases [GET]
func (s *service) GetComponentReleases(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	installs, err := s.getComponentReleases(ctx, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component releases: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installs)
}

func (s *service) getComponentReleases(ctx context.Context, componentID string) ([]app.ComponentRelease, error) {
	var releases []app.ComponentRelease
	res := s.db.WithContext(ctx).
		// join component-releases to component-builds to component-config-connections to components
		Joins("JOIN component_builds ON component_builds.id=component_releases.component_build_id").
		Joins("JOIN component_config_connections ON component_config_connections.id=component_builds.component_config_connection_id").
		Joins("JOIN components ON components.id=component_config_connections.component_id").
		Where("components.id = ?", componentID).
		Order("created_at DESC").
		Find(&releases)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to load component releases")
	}

	return releases, nil
}
