package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetComponentLatestBuild
// @Summary	get latest build for a component
// @Description.markdown	get_component_latest_build.md
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
// @Success		200				{object}	app.ComponentBuild
// @Router			/v1/components/{component_id}/builds/latest [GET]
func (s *service) GetComponentLatestBuild(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	bld, err := s.getComponentLatestBuild(ctx, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component builds: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, bld)
}

func (s *service) getComponentLatestBuild(ctx *gin.Context, cmpID string) (*app.ComponentBuild, error) {
	cmp := app.Component{}

	// query all builds that belong to the component id, starting at the component to ensure the component exists
	// via the double join.
	res := s.db.WithContext(ctx).
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Table(app.ComponentConfigConnection{}.ViewName()).
				Order("component_config_connections_view.created_at DESC")
		}).
		Preload("ComponentConfigs.ComponentBuilds", func(db *gorm.DB) *gorm.DB {
			return db.Order("component_builds.created_at DESC").Limit(1)
		}).
		Preload("ComponentConfigs.ComponentBuilds.VCSConnectionCommit").
		First(&cmp, "id = ?", cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	// pull out the first (and only) component build
	for _, cfg := range cmp.ComponentConfigs {
		for _, bld := range cfg.ComponentBuilds {
			return &bld, nil
		}
	}

	return nil, fmt.Errorf("no build found for component: %w", gorm.ErrRecordNotFound)
}
