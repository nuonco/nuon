package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
)

// @ID						GetBuild
// @Summary				get a build
// @Description.markdown	get_component_build.md
// @Param					build_id	path	string	true	"build ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated			true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.ComponentBuild
// @Router					/v1/components/builds/{build_id} [GET]
func (s *service) GetBuild(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	bldID := ctx.Param("build_id")

	bld, err := s.getBuild(ctx, org.ID, bldID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org build: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, bld)
}

func (s *service) getBuild(ctx context.Context, orgID, bldID string) (*app.ComponentBuild, error) {
	var bld app.ComponentBuild

	// query the build in a way where it will _only_ be returned if it belongs to the component id in question
	res := s.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Preload("RunnerJob").
		Preload("LogStream").
		Preload("ComponentConfigConnection", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(s.db, &app.ComponentConfigConnection{}, ".created_at DESC"))
		}).
		Where("org_id = ?", orgID).
		First(&bld, "id = ?", bldID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get build: %w", res.Error)
	}

	return &bld, nil
}
