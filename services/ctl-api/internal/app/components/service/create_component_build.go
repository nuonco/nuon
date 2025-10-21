package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
)

type CreateComponentBuildRequest struct {
	GitRef    *string `validate:"required_unless=UseLatest true" json:"git_ref"`
	UseLatest bool    `validate:"required_without=GitRef" json:"use_latest"`
}

func (c *CreateComponentBuildRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						CreateAppComponentBuild
// @Summary				create component build
// @Description.markdown	create_component_build.md
// @Param					component_id	path	string						true	"component ID"
// @Param					req				body	CreateComponentBuildRequest	true	"Input"
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
// @Success				201	{object}	app.ComponentBuild
// @Router					/v1/apps/{app_id}/components/{component_id}/builds [POST]
func (s *service) CreateAppComponentBuild(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	cmpID := ctx.Param("component_id")
	cmp, err := s.getAppComponent(ctx, appID, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app component: %w", err))
		return
	}

	var req CreateComponentBuildRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	bld, err := s.helpers.CreateComponentBuild(ctx, cmp.ID, req.UseLatest, req.GitRef)
	if err != nil {
		ctx.Error(err)
		return
	}
	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type:    signals.OperationBuild,
		BuildID: bld.ID,
	})

	ctx.JSON(http.StatusCreated, bld)
}

// @ID						CreateComponentBuild
// @Summary				create component build
// @Description.markdown	create_component_build.md
// @Param					component_id	path	string						true	"component ID"
// @Param					req				body	CreateComponentBuildRequest	true	"Input"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated     true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.ComponentBuild
// @Router					/v1/components/{component_id}/builds [POST]
func (s *service) CreateComponentBuild(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	var req CreateComponentBuildRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	bld, err := s.helpers.CreateComponentBuild(ctx, cmpID, req.UseLatest, req.GitRef)
	if err != nil {
		ctx.Error(err)
		return
	}
	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type:    signals.OperationBuild,
		BuildID: bld.ID,
	})

	ctx.JSON(http.StatusCreated, bld)
}
