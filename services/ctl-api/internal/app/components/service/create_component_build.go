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

// @BasePath /v1/components
// Create component build
// @Summary create component build
// @Schemes
// @Description create component build
// @Param component_id path string true "component ID"
// @Param req body CreateComponentBuildRequest true "Input"
// @Tags components
// @Accept json
// @Produce json
// @Success 201 {object} app.ComponentBuild
// @Router /v1/components/{component_id}/builds [POST]
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

	component, err := s.createComponentBuild(ctx, cmpID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component build: %w", err))
		return
	}
	s.hooks.BuildCreated(ctx, component.ID)
	ctx.JSON(http.StatusCreated, component)
}

func (s *service) createComponentBuild(ctx context.Context, cmpID string, req *CreateComponentBuildRequest) (*app.ComponentBuild, error) {
	var vcsCommit *app.VCSConnectionCommit
	if req.UseLatest {
		var err error
		vcsCommit, err = s.getComponentConnectionCommit(ctx, cmpID)
		if err != nil {
			return nil, fmt.Errorf("unable to get latest commit for connection: %w", err)
		}
	}

	bld := app.ComponentBuild{
		Status: "queued",
		GitRef: req.GitRef,
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
