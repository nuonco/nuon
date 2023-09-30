package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@BasePath	/v1/releases
// Create a release from a build
//	@Summary	create a release
//	@Schemes
//	@Description	create a release for a build
//	@Tags			releases
//	@Accept			json
//	@Param			req	body	CreateComponentReleaseRequest	true	"Input"
//	@Produce		json
//	@Success		201	{object}	app.ComponentRelease
//	@Router			/v1/releases [post]
func (s *service) CreateRelease(ctx *gin.Context) {
	var req CreateComponentReleaseRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	cmp, err := s.getBuildComponent(ctx, req.BuildID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	app, err := s.createRelease(ctx, cmp.ID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	s.hooks.Created(ctx, app.ID)
	ctx.JSON(http.StatusCreated, app)
}

func (s *service) getBuildComponent(ctx context.Context, buildID string) (*app.Component, error) {
	component := app.Component{}
	res := s.db.WithContext(ctx).
		Raw(`
select cmp.*
from components cmp
join component_config_connections cfg on cmp.id = cfg.component_id
join component_builds bld on cfg.id = bld.component_config_connection_id
where bld.id = @buildID
`, sql.Named("buildID", buildID)).
		First(&component)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	component.ConfigVersions = len(component.ComponentConfigs)

	return &component, nil
}
