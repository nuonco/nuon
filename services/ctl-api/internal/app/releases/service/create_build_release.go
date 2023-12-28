package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

// @ID CreateBuildRelease
// @Summary	create a release
// @Description.markdown	create_build_release.md
// @Tags			releases
// @Accept			json
// @Param			req	body	CreateComponentReleaseRequest	true	"Input"
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.ComponentRelease
// @Router			/v1/releases [post]
func (s *service) CreateRelease(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

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

	s.hooks.Created(ctx, app.ID, org.SandboxMode)
	ctx.JSON(http.StatusCreated, app)
}

func (s *service) getBuildComponent(ctx context.Context, buildID string) (*app.Component, error) {
	build := app.ComponentBuild{}

	res := s.db.WithContext(ctx).
		Preload("ComponentConfigConnection").
		Preload("ComponentConfigConnection.Component").
		Preload("ComponentConfigConnection.Component.ComponentConfigs").
		First(&build, "id = ?", buildID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	return &build.ComponentConfigConnection.Component, nil
}
