package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetComponentBuild
// @Summary				get a build for a component
// @Description.markdown	get_component_build.md
// @Param					component_id	path	string	true	"component ID"
// @Param					build_id		path	string	true	"build ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.ComponentBuild
// @Router					/v1/components/{component_id}/builds/{build_id} [GET]
func (s *service) GetComponentBuild(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")
	bldID := ctx.Param("build_id")

	bld, err := s.getComponentBuild(ctx, cmpID, bldID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component build: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, bld)
}

func (s *service) getComponentBuild(ctx context.Context, cmpID, bldID string) (*app.ComponentBuild, error) {
	var bld app.ComponentBuild

	// query the build in a way where it will _only_ be returned if it belongs to the component id in question
	res := s.db.WithContext(ctx).
    Preload("CreatedBy").
		Preload("VCSConnectionCommit").
		Preload("ComponentConfigConnection", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(s.db, &app.ComponentConfigConnection{}, ".created_at DESC"))
		}).
		Preload("ComponentConfigConnection.Component").
		Preload("RunnerJob", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(scopes.WithDisableViews)
		}).
		First(&bld, "id = ?", bldID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component build: %w", res.Error)
	}

	return &bld, nil
}
