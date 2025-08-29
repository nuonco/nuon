package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetComponentReleases
// @BasePath				/v1/components
// @Summary				get all releases for a component
// @Description.markdown	get_component_releases.md
// @Param					component_id				path	string	true	"component ID"
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
// @Success				200	{array}		app.ComponentRelease
// @Router					/v1/components/{component_id}/releases [GET]
func (s *service) GetComponentReleases(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	installs, err := s.getComponentReleases(ctx, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component releases: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installs)
}

func (s *service) getComponentReleases(ctx *gin.Context, componentID string) ([]app.ComponentRelease, error) {
	var releases []app.ComponentRelease
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
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

	releases, err := db.HandlePaginatedResponse(ctx, releases)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return releases, nil
}
