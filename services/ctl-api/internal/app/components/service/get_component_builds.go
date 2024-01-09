package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

// @ID GetComponentBuilds
// @Summary	get all builds for a component
// @Description.markdown	get_component_builds.md
// @Param			component_id	path	string	true	"component ID"
// @Tags			components
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.ComponentBuild
// @Router			/v1/components/{component_id}/builds [GET]
func (s *service) GetComponentBuilds(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")
	if cmpID == "" {
		ctx.Error(fmt.Errorf("component id must be passed in"))
		return
	}

	component, err := s.getComponentBuilds(ctx, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component builds: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, component)
}

func (s *service) getComponentBuilds(ctx context.Context, cmpID string) ([]app.ComponentBuild, error) {
	cmp := app.Component{}

	// query all builds that belong to the component id, starting at the component to ensure the component exists
	// via the double join.
	res := s.db.WithContext(ctx).
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("component_config_connections.created_at DESC")
		}).
		Preload("ComponentConfigs.ComponentBuilds", func(db *gorm.DB) *gorm.DB {
			return db.Order("component_builds.created_at DESC")
		}).
		Preload("ComponentConfigs.ComponentBuilds.VCSConnectionCommit").
		Preload("ComponentConfigs.ComponentBuilds.ComponentConfigConnection").
		First(&cmp, "id = ?", cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	blds := make([]app.ComponentBuild, 0)
	for _, cfg := range cmp.ComponentConfigs {
		blds = append(blds, cfg.ComponentBuilds...)
	}

	return blds, nil
}
