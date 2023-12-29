package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
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

// @ID CreateComponentBuild
// @Summary	create component build
// @Description.markdown	create_component_build.md
// @Param			component_id	path	string						true	"component ID"
// @Param			req				body	CreateComponentBuildRequest	true	"Input"
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
// @Success		201				{object}	app.ComponentBuild
// @Router			/v1/components/{component_id}/builds [POST]
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

	bld, err := s.createComponentBuild(ctx, cmpID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component build: %w", err))
		return
	}
	s.hooks.BuildCreated(ctx, cmpID, bld.ID)
	ctx.JSON(http.StatusCreated, bld)
}

func (s *service) createComponentBuild(ctx context.Context, cmpID string, req *CreateComponentBuildRequest) (*app.ComponentBuild, error) {
	var vcsCommit *app.VCSConnectionCommit
	if req.UseLatest {
		var err error
		vcsCommit, err = s.helpers.GetComponentCommit(ctx, cmpID)
		if err != nil {
			return nil, fmt.Errorf("unable to get latest commit for connection: %w", err)
		}
	}
	gitRef := req.GitRef
	if vcsCommit != nil {
		gitRef = generics.ToPtr(vcsCommit.SHA)
	}

	bld := app.ComponentBuild{
		Status:            "queued",
		StatusDescription: "queued and waiting for runner to pick up",
		GitRef:            gitRef,
	}
	if vcsCommit != nil {
		bld.VCSConnectionCommitID = generics.ToPtr(vcsCommit.ID)
	}

	cmp := app.ComponentConfigConnection{}
	err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(1).
		First(&cmp, "component_id = ?", cmpID).Association("ComponentBuilds").Append(&bld)
	if err != nil {
		return nil, fmt.Errorf("unable to create build for component: %w", err)
	}
	return &bld, nil
}
